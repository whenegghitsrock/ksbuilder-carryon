package api

import (
	"fmt"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"
)

// SanitizeDependencies ensures each dependency has a version. Local subcharts
// infer version from charts/<name>/Chart.yaml. Remote deps without version return an error.
func SanitizeDependencies(extPath string, deps []*chart.Dependency) ([]*chart.Dependency, error) {
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
		if dep.Repository != "" {
			return nil, fmt.Errorf("dependency %q must have version for remote charts", dep.Name)
		}
		chartPath := filepath.Join(extPath, "charts", dep.Name, "Chart.yaml")
		data, err := os.ReadFile(chartPath)
		if err == nil {
			var meta struct{ Version string `yaml:"version"` }
			if yaml.Unmarshal(data, &meta) == nil && meta.Version != "" {
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
