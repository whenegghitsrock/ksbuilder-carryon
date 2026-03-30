package extension

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kubesphere/ksbuilder/pkg/extension/spec"
)

func TestCreateFromSpec_StandardBoth(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "verify-ext")

	s := &spec.Spec{
		Name:         "verify-ext",
		Version:      "0.1.0",
		Mode:         spec.ModeStandard,
		Capabilities: spec.Capabilities{Frontend: true, Backend: true},
		Metadata: spec.Metadata{
			Category: "ai-machine-learning",
			Author:   "tester",
			Email:    "test@example.com",
		},
		Permissions: spec.PermDefault,
	}

	if err := CreateFromSpec(root, s); err != nil {
		t.Fatalf("CreateFromSpec: %v", err)
	}

	verifyExtensionInstallationMode(t, root, "Multicluster")
	verifyFiles(t, root, true, true)
	verifyPackage(t, root)
}

func TestCreateFromSpec_FrontendOnly(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "frontend-only-ext")

	s := &spec.Spec{
		Name:         "frontend-only-ext",
		Version:      "0.1.0",
		Mode:         spec.ModeStandard,
		Capabilities: spec.Capabilities{Frontend: true, Backend: false},
		Metadata:     spec.Metadata{Category: "database"},
		Permissions:  spec.PermDefault,
	}

	if err := CreateFromSpec(root, s); err != nil {
		t.Fatalf("CreateFromSpec: %v", err)
	}

	// extension.yaml should use lowercase keys (apiVersion, not APIVersion) per template style
	extYAML, err := os.ReadFile(filepath.Join(root, "extension.yaml"))
	if err != nil {
		t.Fatalf("read extension.yaml: %v", err)
	}
	extStr := string(extYAML)
	if !strings.Contains(extStr, "apiVersion:") {
		t.Error("extension.yaml should contain apiVersion: (lowercase)")
	}
	if strings.Contains(extStr, "APIVersion:") {
		t.Error("extension.yaml should not contain APIVersion: (PascalCase)")
	}

	verifyExtensionInstallationMode(t, root, "HostOnly")
	verifyFiles(t, root, true, false)
	verifyPackage(t, root)
}

func TestCreateFromSpec_BackendOnly(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "backend-only-ext")

	s := &spec.Spec{
		Name:         "backend-only-ext",
		Version:      "0.1.0",
		Mode:         spec.ModeStandard,
		Capabilities: spec.Capabilities{Frontend: false, Backend: true},
		Metadata:     spec.Metadata{Category: "observability"},
		Permissions:  spec.PermDefault,
	}

	if err := CreateFromSpec(root, s); err != nil {
		t.Fatalf("CreateFromSpec: %v", err)
	}

	verifyExtensionInstallationMode(t, root, "Multicluster")
	verifyFiles(t, root, false, true)
	verifyPackage(t, root)
}

func verifyExtensionInstallationMode(t *testing.T, root, want string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "extension.yaml"))
	if err != nil {
		t.Fatalf("read extension.yaml: %v", err)
	}
	if !strings.Contains(string(data), "installationMode: "+want) {
		t.Errorf("extension.yaml installationMode want %q", want)
	}
}

// frontendHelloScaffoldRelPaths matches templates under pkg/extension/templates/frontend (Hello World layout + runtime dist).
var frontendHelloScaffoldRelPaths = []string{
	"frontend/Dockerfile",
	"frontend/index.html",
	"frontend/Makefile",
	"frontend/README.md",
	"frontend/package.json",
	"frontend/package-lock.json",
	"frontend/dist/index.js",
	"frontend/src/App.jsx",
	"frontend/src/iframe.jsx",
	"frontend/src/index.js",
	"frontend/src/routes/index.js",
	"frontend/src/locales/index.js",
	"frontend/src/locales/en/index.js",
	"frontend/src/locales/en/base.json",
	"frontend/src/locales/zh/index.js",
	"frontend/src/locales/zh/base.json",
}

func verifyFrontendHelloScaffold(t *testing.T, root string) {
	t.Helper()
	for _, p := range frontendHelloScaffoldRelPaths {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}
}

func TestCreateFromSpec_CopiesFrontendScaffold(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "scaffold-ext")
	s := &spec.Spec{
		Name:         "scaffold-ext",
		Version:      "0.1.0",
		Mode:         spec.ModeStandard,
		Capabilities: spec.Capabilities{Frontend: true, Backend: false},
		Metadata:     spec.Metadata{Category: "database"},
		Permissions:  spec.PermDefault,
	}
	if err := CreateFromSpec(root, s); err != nil {
		t.Fatalf("CreateFromSpec: %v", err)
	}
	verifyFrontendHelloScaffold(t, root)
	if _, err := os.Stat(filepath.Join(root, "Makefile")); err != nil {
		t.Errorf("missing root Makefile: %v", err)
	}
	// Must not have backend scaffold
	if _, err := os.Stat(filepath.Join(root, "backend")); err == nil {
		t.Error("backend scaffold should not exist for frontend-only")
	}
}

func TestCreateFromSpec_CopiesBackendScaffold(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "backend-ext")
	s := &spec.Spec{
		Name:         "backend-ext",
		Version:      "0.1.0",
		Mode:         spec.ModeStandard,
		Capabilities: spec.Capabilities{Frontend: false, Backend: true},
		Metadata:     spec.Metadata{Category: "observability"},
		Permissions:  spec.PermDefault,
	}
	if err := CreateFromSpec(root, s); err != nil {
		t.Fatalf("CreateFromSpec: %v", err)
	}
	for _, p := range []string{"backend/Dockerfile", "backend/main.go", "backend/go.mod", "backend/Makefile"} {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "Makefile")); err != nil {
		t.Errorf("missing root Makefile: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "frontend")); err == nil {
		t.Error("frontend scaffold should not exist for backend-only")
	}
}

func verifyFiles(t *testing.T, root string, wantFrontend, wantBackend bool) {
	t.Helper()
	required := []string{
		".ksbuilder.yaml",
		".helmignore",
		"extension.yaml",
		"permissions.yaml",
		"values.yaml",
		"static/favicon.svg",
		"static/screenshots/screenshot.png",
	}
	for _, name := range required {
		p := filepath.Join(root, name)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing required file %s: %v", name, err)
		}
	}
	if wantFrontend {
		if _, err := os.Stat(filepath.Join(root, "charts", "frontend", "Chart.yaml")); err != nil {
			t.Errorf("missing charts/frontend: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "frontend", "Dockerfile")); err != nil {
			t.Errorf("missing frontend scaffold: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "frontend", "templates")); err == nil {
			t.Error("frontend/templates should not exist (Helm templates are in charts/frontend/templates)")
		}
	}
	if wantBackend {
		if _, err := os.Stat(filepath.Join(root, "charts", "backend", "Chart.yaml")); err != nil {
			t.Errorf("missing charts/backend: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "backend", "Dockerfile")); err != nil {
			t.Errorf("missing backend scaffold: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "backend", "templates")); err == nil {
			t.Error("backend/templates should not exist (Helm templates are in charts/backend/templates)")
		}
	}
	if wantFrontend || wantBackend {
		if _, err := os.Stat(filepath.Join(root, "Makefile")); err != nil {
			t.Errorf("missing root Makefile: %v", err)
		}
	}
}

func verifyPackage(t *testing.T, root string) {
	t.Helper()
	outDir := t.TempDir()
	tgz, err := PackageToPath(root, outDir, true)
	if err != nil {
		t.Fatalf("PackageToPath: %v", err)
	}
	if tgz == "" {
		t.Fatal("PackageToPath returned empty path")
	}
	if _, err := os.Stat(tgz); err != nil {
		t.Fatalf("packaged tgz missing: %v", err)
	}
}

func getTestdataPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata")
}

func TestCreateApp_ExtensionYAML(t *testing.T) {
	chartPath := filepath.Join(getTestdataPath(), "minimal-chart-0.1.0.tgz")
	if _, err := os.Stat(chartPath); err != nil {
		t.Skip("testdata chart fixture not found, run: cd pkg/extension/testdata && helm package minimal-chart")
	}
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWd) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	absChartPath, _ := filepath.Abs(chartPath)
	if err := CreateApp(absChartPath); err != nil {
		t.Fatalf("CreateApp: %v", err)
	}
	extPath := filepath.Join(dir, "minimal-chart", "extension.yaml")
	data, err := os.ReadFile(extPath)
	if err != nil {
		t.Fatalf("read extension.yaml: %v", err)
	}
	if !strings.Contains(string(data), "name: minimal-chart") {
		t.Error("extension.yaml should contain extension name")
	}
	if !strings.Contains(string(data), "demoauthor") {
		t.Error("extension.yaml should contain demoauthor")
	}
}

func TestCreateSimple_ExtensionYAML(t *testing.T) {
	chartPath := filepath.Join(getTestdataPath(), "minimal-chart-0.1.0.tgz")
	if _, err := os.Stat(chartPath); err != nil {
		t.Skip("testdata chart fixture not found, run: cd pkg/extension/testdata && helm package minimal-chart")
	}
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWd) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	absChartPath, _ := filepath.Abs(chartPath)
	if err := CreateSimple(absChartPath); err != nil {
		t.Fatalf("CreateSimple: %v", err)
	}
	extPath := filepath.Join(dir, "minimal-chart", "extension.yaml")
	data, err := os.ReadFile(extPath)
	if err != nil {
		t.Fatalf("read extension.yaml: %v", err)
	}
	if !strings.Contains(string(data), "name: minimal-chart") {
		t.Error("extension.yaml should contain extension name")
	}
	if !strings.Contains(string(data), "demoauthor") {
		t.Error("extension.yaml should contain demoauthor")
	}
}
