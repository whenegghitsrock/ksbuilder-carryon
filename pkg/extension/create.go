package extension

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"hauler.dev/go/hauler/pkg/archives"
	"helm.sh/helm/v3/pkg/chart/loader"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/ksbuilder/pkg/api"
	"github.com/kubesphere/ksbuilder/pkg/extension/generator"
	"github.com/kubesphere/ksbuilder/pkg/extension/spec"
)

type Config struct {
	Name     string
	Category string
	Author   string
	Email    string
	URL      string
}
type ConfigSimple struct {
	Name string
}
type ConfigApp struct {
	Name           string
	Maintainers    string
	AppName        string
	Version        string
	AppHome        string
	Abstraction    string
	AppVersionName string
	ZipName        string
}

//go:embed templates templates/.helmignore templates/frontend
//go:embed templates/frontend/.editorconfig templates/frontend/.eslintignore templates/frontend/.eslintrc.js templates/frontend/.gitignore templates/frontend/.npmrc templates/frontend/.prettierignore templates/frontend/.prettierrc.js
var Templates embed.FS

//go:embed templatessimple
var Templatessimple embed.FS

//go:embed templatesapp
var Templatesapp embed.FS

func copyZipFile(srcPath, dstPath string) error {
	fileName := filepath.Base(srcPath)
	dstFilePath := filepath.Join(dstPath, fileName)
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		_ = srcFile.Close()
	}(srcFile)
	dstFile, err := os.Create(dstFilePath)
	if err != nil {
		return err
	}
	defer func(dstFile *os.File) {
		_ = dstFile.Close()
	}(dstFile)
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func generateRandomString() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	const charset = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 5)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
func CreateApp(chartPath string) error {
	pwd, _ := os.Getwd()
	f, err := os.ReadFile(chartPath)
	if err != nil {
		return err
	}
	chartPack, err := loader.LoadArchive(bytes.NewReader(f))
	if err != nil {
		return err
	}
	root := path.Join(pwd, chartPack.Name())
	appName := fmt.Sprintf("%s-%s", chartPack.Name(), generateRandomString())
	appVersionName := fmt.Sprintf("%s-%s-%s", appName, chartPack.AppVersion(), generateRandomString())
	err = os.MkdirAll(root, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(root, "charts", "base", "files"), 0755)
	if err != nil {
		return err
	}

	err = copyZipFile(chartPath, filepath.Join(root, "charts", "base", "files"))
	if err != nil {
		return err
	}

	extensionConfig := ConfigApp{
		Name:           chartPack.Name(),
		AppName:        appName,
		Version:        chartPack.Metadata.Version,
		AppHome:        chartPack.Metadata.Home,
		Abstraction:    chartPack.Metadata.Description,
		AppVersionName: appVersionName,
		ZipName:        filepath.Base(chartPath),
	}
	if len(chartPack.Metadata.Maintainers) > 0 {
		extensionConfig.Maintainers = chartPack.Metadata.Maintainers[0].Name
	} else {
		extensionConfig.Maintainers = "admin"
	}

	err = Create(root, extensionConfig, Templatesapp, "templatesapp")
	if err != nil {
		return err
	}
	if err := copySharedFromTemplates(root, extensionConfig); err != nil {
		return fmt.Errorf("copy shared templates: %w", err)
	}
	extYAML, err := generator.ExtensionYAMLForChartMode(chartPack.Name())
	if err != nil {
		return fmt.Errorf("generate extension.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, "extension.yaml"), extYAML, 0644); err != nil {
		return err
	}
	return nil
}

func CreateSimple(chartPath string) error {
	pwd, _ := os.Getwd()
	f, err := os.ReadFile(chartPath)
	if err != nil {
		return err
	}
	chartPack, err := loader.LoadArchive(bytes.NewReader(f))
	if err != nil {
		return err
	}
	root := path.Join(pwd, chartPack.Name())
	err = os.MkdirAll(root, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(root, "charts"), 0755)
	if err != nil {
		return err
	}

	err = archives.Unarchive(context.Background(), chartPath, filepath.Join(root, "charts"))
	if err != nil {
		return fmt.Errorf("failed to extract chart: %w", err)
	}
	extensionConfig := ConfigSimple{
		Name: chartPack.Name(),
	}
	err = Create(root, extensionConfig, Templatessimple, "templatessimple")
	if err != nil {
		return err
	}
	if err := copySharedFromTemplates(root, extensionConfig); err != nil {
		return fmt.Errorf("copy shared templates: %w", err)
	}
	extYAML, err := generator.ExtensionYAMLForChartMode(chartPack.Name())
	if err != nil {
		return fmt.Errorf("generate extension.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, "extension.yaml"), extYAML, 0644); err != nil {
		return err
	}
	return nil
}

func Create(p string, config any, temp embed.FS, trimPrefix string) error {
	return fs.WalkDir(temp, ".", func(templatePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		innerPath := strings.TrimPrefix(templatePath, trimPrefix)
		if d.IsDir() {
			if innerPath != "" { // Ignore the templates parent directory
				if err = os.MkdirAll(filepath.Join(p, innerPath), 0755); err != nil {
					return err
				}
			}
			return nil
		}

		t, err := template.New(path.Base(templatePath)).Delims("[[", "]]").ParseFS(temp, templatePath)
		if err != nil {
			return err
		}
		f, err := os.Create(filepath.Join(p, innerPath))
		if err != nil {
			return err
		}
		defer f.Close() // nolint
		return t.Execute(f, config)
	})
}

// CreateFromSpec creates an extension from a declarative spec (standard mode).
func CreateFromSpec(root string, s *spec.Spec) error {
	if s.Mode != spec.ModeStandard {
		return fmt.Errorf("CreateFromSpec only supports mode=standard, got %s", s.Mode)
	}
	if !s.HasFrontend() && !s.HasBackend() {
		return fmt.Errorf("at least one of frontend or backend must be enabled")
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}
	config := struct {
		Name        string
		HasFrontend bool
		HasBackend  bool
	}{Name: s.Name, HasFrontend: s.HasFrontend(), HasBackend: s.HasBackend()}

	if err := s.Write(root); err != nil {
		return err
	}
	extYAML, err := generator.ExtensionYAML(s)
	if err != nil {
		return fmt.Errorf("generate extension.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, "extension.yaml"), extYAML, 0644); err != nil {
		return err
	}
	permYAML, err := generator.PermissionsYAML(s)
	if err != nil {
		return fmt.Errorf("generate permissions.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, "permissions.yaml"), permYAML, 0644); err != nil {
		return err
	}
	valsYAML, err := generator.ValuesYAML(s)
	if err != nil {
		return fmt.Errorf("generate values.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, "values.yaml"), valsYAML, 0644); err != nil {
		return err
	}
	if s.HasFrontend() {
		if err := copySubtree(Templates, "templates/charts/frontend", filepath.Join(root, "charts", "frontend"), config); err != nil {
			return fmt.Errorf("copy frontend chart: %w", err)
		}
	}
	if s.HasBackend() {
		if err := copySubtree(Templates, "templates/charts/backend", filepath.Join(root, "charts", "backend"), config); err != nil {
			return fmt.Errorf("copy backend chart: %w", err)
		}
	}
	if err := copySubtree(Templates, "templates/static", filepath.Join(root, "static"), config); err != nil {
		return fmt.Errorf("copy static: %w", err)
	}
	if s.HasFrontend() {
		if err := ensureFrontendWorkspace(root, config); err != nil {
			return fmt.Errorf("prepare frontend workspace: %w", err)
		}
	}
	if s.HasBackend() {
		if err := copySubtreeWithRename(Templates, "templates/backend", filepath.Join(root, "backend"), config, map[string]string{"go.mod.tpl": "go.mod"}); err != nil {
			return fmt.Errorf("copy backend scaffold: %w", err)
		}
	}
	if s.HasFrontend() || s.HasBackend() {
		makefile, err := fs.ReadFile(Templates, "templates/Makefile")
		if err != nil {
			return fmt.Errorf("read Makefile template: %w", err)
		}
		t, err := template.New("Makefile").Delims("[[", "]]").Parse(string(makefile))
		if err != nil {
			return fmt.Errorf("parse Makefile template: %w", err)
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, config); err != nil {
			return fmt.Errorf("render Makefile: %w", err)
		}
		if err := os.WriteFile(filepath.Join(root, "Makefile"), buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("write Makefile: %w", err)
		}
	}
	for _, name := range []string{"README.md", "README_zh.md", "CHANGELOG.md", "CHANGELOG_zh.md", ".helmignore"} {
		data, err := fs.ReadFile(Templates, "templates/"+name)
		if err != nil {
			continue
		}
		t, err := template.New(name).Delims("[[", "]]").Parse(string(data))
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, config); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, name), buf.Bytes(), 0644); err != nil {
			return err
		}
	}
	return nil
}

const (
	frontendTemplateDir         = "templates/frontend"
	frontendTemplateScaffoldRel = "extensions/scaffold" // repo-only; materialized as frontend/extensions/<extName>/
)

// copyFrontendWorkspace copies templates/frontend to dest except the embedded scaffold subtree
// (extensions/scaffold), which is materialized separately under frontend/extensions/<extName>/.
func copyFrontendWorkspace(f embed.FS, destDir string, config any) error {
	return fs.WalkDir(f, frontendTemplateDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == frontendTemplateDir {
			if d.IsDir() {
				return os.MkdirAll(destDir, 0755)
			}
			return nil
		}
		rel, _ := strings.CutPrefix(p, frontendTemplateDir+"/")
		if rel == "" {
			return nil
		}
		if rel == frontendTemplateScaffoldRel || strings.HasPrefix(rel, frontendTemplateScaffoldRel+"/") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		dstPath := filepath.Join(destDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		data, err := fs.ReadFile(f, p)
		if err != nil {
			return err
		}
		t, err := template.New(filepath.Base(p)).Delims("[[", "]]").Parse(string(data))
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}
		out, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer func() { _ = out.Close() }()
		return t.Execute(out, config)
	})
}

// materializeFrontendWorkspace writes templates/frontend under the extension's **frontend/** directory:
// workspace files at frontend/, scaffold from templates/frontend/extensions/scaffold/ → frontend/extensions/<name>/.
func materializeFrontendWorkspace(root, extName string, tmplConfig any) error {
	frontendRoot := filepath.Join(root, "frontend")
	if err := copyFrontendWorkspace(Templates, frontendRoot, tmplConfig); err != nil {
		return fmt.Errorf("copy frontend template: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(frontendRoot, "configs", "extensions"), 0755); err != nil {
		return err
	}
	extDir := filepath.Join(frontendRoot, "extensions", extName)
	scaffoldSrc := frontendTemplateDir + "/" + frontendTemplateScaffoldRel
	if err := copySubtree(Templates, scaffoldSrc, extDir, tmplConfig); err != nil {
		return fmt.Errorf("copy frontend scaffold into extensions/%s: %w", extName, err)
	}
	wrapper := `REGISTRY ?= docker.io
NAMESPACE ?= kubespheredev
NAME ?= ` + extName + `
TAG ?= latest
TAG := $(or $(TAG),latest)

.PHONY: build build-assets push

build:
	$(MAKE) -C extensions/$(NAME) build NAME=$(NAME) REGISTRY=$(REGISTRY) NAMESPACE=$(NAMESPACE) TAG=$(TAG)

build-assets:
	$(MAKE) -C extensions/$(NAME) build-assets NAME=$(NAME) REGISTRY=$(REGISTRY) NAMESPACE=$(NAMESPACE) TAG=$(TAG)

push:
	$(MAKE) -C extensions/$(NAME) push NAME=$(NAME) REGISTRY=$(REGISTRY) NAMESPACE=$(NAMESPACE) TAG=$(TAG)
`
	if err := os.WriteFile(filepath.Join(frontendRoot, "Makefile"), []byte(wrapper), 0644); err != nil {
		return err
	}
	return nil
}

func ensureFrontendWorkspace(root string, config struct {
	Name        string
	HasFrontend bool
	HasBackend  bool
}) error {
	return materializeFrontendWorkspace(root, config.Name, config)
}

// copySharedFromTemplates copies static/, README*, CHANGELOG*, .helmignore from templates to root.
// Used by CreateApp and CreateSimple to avoid duplicating these files in templatesapp/templatessimple.
func copySharedFromTemplates(root string, config any) error {
	if err := copySubtree(Templates, "templates/static", filepath.Join(root, "static"), config); err != nil {
		return err
	}
	for _, name := range []string{"README.md", "README_zh.md", "CHANGELOG.md", "CHANGELOG_zh.md", ".helmignore"} {
		data, err := fs.ReadFile(Templates, "templates/"+name)
		if err != nil {
			continue
		}
		t, err := template.New(name).Delims("[[", "]]").Parse(string(data))
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, config); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, name), buf.Bytes(), 0644); err != nil {
			return err
		}
	}
	return nil
}

func copySubtree(f embed.FS, srcPrefix, destDir string, config any) error {
	return fs.WalkDir(f, srcPrefix, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == srcPrefix {
			if d.IsDir() {
				return os.MkdirAll(destDir, 0755)
			}
			return nil
		}
		rel, _ := strings.CutPrefix(p, srcPrefix+"/")
		if rel == "" {
			return nil
		}
		dstPath := filepath.Join(destDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		data, err := fs.ReadFile(f, p)
		if err != nil {
			return err
		}
		t, err := template.New(filepath.Base(p)).Delims("[[", "]]").Parse(string(data))
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}
		out, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer func() { _ = out.Close() }()
		return t.Execute(out, config)
	})
}

func copySubtreeWithRename(f embed.FS, srcPrefix, destDir string, config any, renames map[string]string) error {
	return fs.WalkDir(f, srcPrefix, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == srcPrefix {
			if d.IsDir() {
				return os.MkdirAll(destDir, 0755)
			}
			return nil
		}
		rel, _ := strings.CutPrefix(p, srcPrefix+"/")
		if rel == "" {
			return nil
		}
		dstRel := rel
		if renames != nil {
			if newName, ok := renames[filepath.Base(rel)]; ok {
				dstRel = filepath.Join(filepath.Dir(rel), newName)
			}
		}
		dstPath := filepath.Join(destDir, dstRel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		data, err := fs.ReadFile(f, p)
		if err != nil {
			return err
		}
		t, err := template.New(filepath.Base(p)).Delims("[[", "]]").Parse(string(data))
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}
		out, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer func() { _ = out.Close() }()
		return t.Execute(out, config)
	})
}

func CreateAppChart(p string, name string, chart []byte) error {
	var cmName = fmt.Sprintf("application-%s-chart", name)

	var cm = corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: api.KubeSphereSystem,
		},
		BinaryData: map[string][]byte{
			api.ConfigMapDataKey: chart,
		},
	}
	cmByte, err := yaml.Marshal(cm)
	if err != nil {
		return err
	}

	filePath := path.Join(p, "application-package.yaml")
	return os.WriteFile(filePath, cmByte, 0644)
}
