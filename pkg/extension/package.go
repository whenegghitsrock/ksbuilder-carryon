package extension

import (
	"fmt"
	"os"
	"path"

	"github.com/otiai10/copy"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/ksbuilder/pkg/api"
)

// PackageToPath packages an extension directory into a .tgz and returns the absolute path.
// If outputDir is empty, a temp dir is used; caller must clean up after upload.
func PackageToPath(extPath, outputDir string, skipDependencyUpdate bool) (string, error) {
	pwd, _ := os.Getwd()
	if !path.IsAbs(extPath) {
		extPath = path.Join(pwd, extPath)
	}
	if outputDir == "" {
		d, err := os.MkdirTemp("", "ksbuilder-publish-")
		if err != nil {
			return "", err
		}
		outputDir = d
	}

	opt := copy.Options{
		OnSymlink: func(_ string) copy.SymlinkAction { return copy.Deep },
	}
	tempDir, err := os.MkdirTemp("", "chart")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	if err = copy.Copy(extPath, tempDir, opt); err != nil {
		return "", err
	}

	metadata, err := api.LoadMetadata(extPath)
	if err != nil {
		return "", err
	}
	sanitizedDeps, err := api.SanitizeDependencies(extPath, metadata.Dependencies)
	if err != nil {
		return "", err
	}
	metadata.Dependencies = sanitizedDeps

	chartMetadata, err := yaml.Marshal(metadata.ToChartYaml())
	if err != nil {
		return "", err
	}
	if err = os.WriteFile(tempDir+"/Chart.yaml", chartMetadata, 0644); err != nil {
		return "", err
	}

	if !skipDependencyUpdate && len(metadata.Dependencies) > 0 {
		settings := cli.New()
		man := &downloader.Manager{
			Out:              os.Stdout,
			ChartPath:        tempDir,
			Keyring:          "",
			SkipUpdate:       false,
			Getters:          getter.All(settings),
			RepositoryConfig: settings.RepositoryConfig,
			RepositoryCache:  settings.RepositoryCache,
			Debug:            settings.Debug,
		}
		if err := man.Update(); err != nil {
			return "", fmt.Errorf("failed to update chart dependencies: %w", err)
		}
	}

	ch, err := loader.LoadDir(tempDir)
	if err != nil {
		return "", err
	}
	return chartutil.Save(ch, outputDir)
}
