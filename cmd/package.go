package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/ksbuilder/pkg/api"
)

type packageOptions struct {
	skipDependencyUpdate bool
}

func defaultPackageOptions() *packageOptions {
	return &packageOptions{}
}

func packageExtensionCmd() *cobra.Command {
	o := defaultPackageOptions()

	cmd := &cobra.Command{
		Use:   "package",
		Short: "package an extension",
		Args:  cobra.ExactArgs(1),
		RunE:  o.packageCmd,
	}
	cmd.Flags().BoolVar(&o.skipDependencyUpdate, "skip-dependency-update", false, "skip running helm dependency update before packaging")
	return cmd
}

func (o *packageOptions) packageCmd(_ *cobra.Command, args []string) error {
	pwd, _ := os.Getwd()
	p := args[0]
	if !path.IsAbs(p) {
		p = path.Join(pwd, p)
	}
	fmt.Printf("package extension %s\n", args[0])

	tempDir, err := os.MkdirTemp("", "chart")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir) // nolint

	// Configure copy to follow symlinks and copy their target content
	opt := copy.Options{
		OnSymlink: func(_ string) copy.SymlinkAction {
			return copy.Deep
		},
	}
	if err = copy.Copy(p, tempDir, opt); err != nil {
		return err
	}

	metadata, err := api.LoadMetadata(p)
	if err != nil {
		return err
	}

	// Ensure all dependencies have valid version for helm dependency update.
	// Local subcharts (no repository) often omit version in extension.yaml; inject from charts/<name>/Chart.yaml.
	sanitizedDeps, err := sanitizeDependencies(p, metadata.Dependencies)
	if err != nil {
		return err
	}
	metadata.Dependencies = sanitizedDeps

	chartMetadata, err := yaml.Marshal(metadata.ToChartYaml())
	if err != nil {
		return err
	}

	if err = os.WriteFile(tempDir+"/Chart.yaml", chartMetadata, 0644); err != nil {
		return err
	}

	if !o.skipDependencyUpdate && len(metadata.Dependencies) > 0 {
		settings := cli.New()
		p := getter.All(settings)
		man := &downloader.Manager{
			Out:              os.Stdout,
			ChartPath:        tempDir,
			Keyring:          "",
			SkipUpdate:       false,
			Getters:          p,
			RepositoryConfig: settings.RepositoryConfig,
			RepositoryCache:  settings.RepositoryCache,
			Debug:            settings.Debug,
		}
		if err := man.Update(); err != nil {
			return fmt.Errorf("failed to update chart dependencies: %w\nHint: run 'helm dependency update' in the extension directory, or use --skip-dependency-update", err)
		}
	}

	ch, err := loader.LoadDir(tempDir)
	if err != nil {
		return err
	}
	chartFilename, err := chartutil.Save(ch, pwd)
	if err != nil {
		return err
	}
	fmt.Printf("package saved to %s\n", chartFilename)
	return nil
}

// sanitizeDependencies ensures each dependency has a valid version for helm dependency update.
// Local subcharts (repository empty) often omit version in extension.yaml; infer from charts/<name>/Chart.yaml.
// Remote deps without version return an error.
func sanitizeDependencies(extPath string, deps []*chart.Dependency) ([]*chart.Dependency, error) {
	if len(deps) == 0 {
		return deps, nil
	}
	result := make([]*chart.Dependency, len(deps))
	for i, dep := range deps {
		d := *dep
		if d.Version != "" {
			result[i] = &d
			continue
		}
		if d.Repository != "" {
			return nil, fmt.Errorf("dependency %q in extension.yaml must have 'version' for remote charts (repository is set)\nHint: add version and repository, or use --skip-dependency-update if deps are local only", d.Name)
		}
		// Local subchart: try to read version from charts/<name>/Chart.yaml
		chartPath := filepath.Join(extPath, "charts", d.Name, "Chart.yaml")
		data, err := os.ReadFile(chartPath)
		if err == nil {
			var meta struct {
				Version string `yaml:"version"`
			}
			if err := yaml.Unmarshal(data, &meta); err == nil && meta.Version != "" {
				d.Version = meta.Version
				result[i] = &d
				continue
			}
		}
		d.Version = ">=0.0.0"
		result[i] = &d
	}
	return result, nil
}
