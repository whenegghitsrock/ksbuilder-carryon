package extension

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"sigs.k8s.io/yaml"

	"github.com/kubesphere/ksbuilder/pkg/api"
)

// CurrentCapabilities describes what an existing extension already has.
type CurrentCapabilities struct {
	HasFrontend bool
	HasBackend  bool
	Name        string // from extension.yaml
}

// InferCapabilities reads extension.yaml and directory structure to determine capabilities.
// Returns nil if dir is not a valid extendable extension (no extension.yaml or app/simple mode).
func InferCapabilities(dir string) (*CurrentCapabilities, error) {
	metadata, err := api.LoadMetadata(dir, api.WithEncodeIcon(false))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	// Only standard-mode extensions are extendable (have frontend/backend subcharts)
	if len(metadata.Dependencies) == 0 {
		return nil, nil
	}
	cap := &CurrentCapabilities{Name: metadata.Name}
	for _, d := range metadata.Dependencies {
		if d.Name == "frontend" {
			cap.HasFrontend = true
		}
		if d.Name == "backend" {
			cap.HasBackend = true
		}
	}
	// App/simple extensions have deps like "base" but no frontend/backend → not extendable
	if !cap.HasFrontend && !cap.HasBackend {
		return nil, nil
	}
	return cap, nil
}

// extendEnsureBothDeps ensures extension.yaml has both frontend and backend dependencies.
// No-op if both already present. Caller must have at least one of frontend/backend in deps.
func extendEnsureBothDeps(root string) error {
	extPath := filepath.Join(root, "extension.yaml")
	data, err := os.ReadFile(extPath)
	if err != nil {
		return err
	}
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}
	// Check if both deps present (handle dependencies or Dependencies key)
	var deps []interface{}
	if v, ok := raw["dependencies"]; ok {
		deps, _ = v.([]interface{})
	}
	if v, ok := raw["Dependencies"]; ok {
		deps, _ = v.([]interface{})
	}
	hasFrontend, hasBackend := false, false
	for _, d := range deps {
		m, _ := d.(map[string]interface{})
		if n, _ := m["name"].(string); n == "frontend" {
			hasFrontend = true
		}
		if n, _ := m["name"].(string); n == "backend" {
			hasBackend = true
		}
	}
	if hasFrontend && hasBackend {
		return nil
	}
	newDeps := []map[string]interface{}{
		{"name": "frontend", "tags": []string{"extension"}},
		{"name": "backend", "tags": []string{"agent"}},
	}
	delete(raw, "Dependencies")
	raw["dependencies"] = newDeps
	out, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	return os.WriteFile(extPath, out, 0644)
}

// extendSetValuesSection enables a section (frontend or backend) in values.yaml.
// repoSuffix is e.g. "-frontend" or "-api" for the image repository.
func extendSetValuesSection(root, section, repoSuffix, name string) error {
	valsPath := filepath.Join(root, "values.yaml")
	valsData, err := os.ReadFile(valsPath)
	if err != nil {
		return err
	}
	var vals map[string]interface{}
	if err := yaml.Unmarshal(valsData, &vals); err != nil {
		return err
	}
	repo := "kubespheredev/" + name + repoSuffix
	if sec, ok := vals[section].(map[string]interface{}); ok {
		sec["enabled"] = true
		if img, ok := sec["image"].(map[string]interface{}); ok {
			img["repository"] = repo
			img["tag"] = "latest"
		}
	} else {
		vals[section] = map[string]interface{}{
			"enabled": true,
			"image": map[string]interface{}{
				"repository": repo,
				"tag":        "latest",
			},
		}
	}
	valsOut, err := yaml.Marshal(vals)
	if err != nil {
		return err
	}
	return os.WriteFile(valsPath, valsOut, 0644)
}

// ExtendAddBackend adds backend capability to an existing frontend-only extension.
// root must contain extension.yaml, values.yaml, charts/frontend.
func ExtendAddBackend(root string) error {
	metadata, err := api.LoadMetadata(root, api.WithEncodeIcon(false))
	if err != nil {
		return err
	}
	name := metadata.Name
	config := struct {
		Name        string
		HasFrontend bool
		HasBackend  bool
	}{Name: name, HasFrontend: true, HasBackend: true}

	// 1. Ensure extension.yaml has both frontend and backend deps
	if err := extendEnsureBothDeps(root); err != nil {
		return err
	}

	// 2. Update values.yaml backend section
	if err := extendSetValuesSection(root, "backend", "-api", name); err != nil {
		return err
	}

	// 3. Copy charts/backend
	if err := copySubtree(Templates, "templates/charts/backend", filepath.Join(root, "charts", "backend"), config); err != nil {
		return err
	}
	// 4. Copy backend scaffold
	if err := copySubtreeWithRename(Templates, "templates/backend", filepath.Join(root, "backend"), config, map[string]string{"go.mod.tpl": "go.mod"}); err != nil {
		return err
	}
	// 5. Regenerate root Makefile
	makefile, err := fs.ReadFile(Templates, "templates/Makefile")
	if err != nil {
		return err
	}
	t, err := template.New("Makefile").Delims("[[", "]]").Parse(string(makefile))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, config); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "Makefile"), buf.Bytes(), 0644)
}

// ExtendAddFrontend adds frontend capability to an existing backend-only extension.
// root must contain extension.yaml, values.yaml, charts/backend.
func ExtendAddFrontend(root string) error {
	metadata, err := api.LoadMetadata(root, api.WithEncodeIcon(false))
	if err != nil {
		return err
	}
	name := metadata.Name
	config := struct {
		Name        string
		HasFrontend bool
		HasBackend  bool
	}{Name: name, HasFrontend: true, HasBackend: true}

	// 1. Ensure extension.yaml has both frontend and backend deps
	if err := extendEnsureBothDeps(root); err != nil {
		return err
	}

	// 2. Update values.yaml frontend section
	valsPath := filepath.Join(root, "values.yaml")
	valsData, err := os.ReadFile(valsPath)
	if err != nil {
		return err
	}
	var vals map[string]interface{}
	if err := yaml.Unmarshal(valsData, &vals); err != nil {
		return err
	}
	if fe, ok := vals["frontend"].(map[string]interface{}); ok {
		fe["enabled"] = true
		if img, ok := fe["image"].(map[string]interface{}); ok {
			img["repository"] = "kubespheredev/" + name + "-frontend"
			img["tag"] = "latest"
		}
	} else {
		vals["frontend"] = map[string]interface{}{
			"enabled": true,
			"image": map[string]interface{}{
				"repository": "kubespheredev/" + name + "-frontend",
				"tag":        "latest",
			},
		}
	}
	valsOut, err := yaml.Marshal(vals)
	if err != nil {
		return err
	}
	if err := os.WriteFile(valsPath, valsOut, 0644); err != nil {
		return err
	}

	// 3. Copy charts/frontend
	if err := copySubtree(Templates, "templates/charts/frontend", filepath.Join(root, "charts", "frontend"), config); err != nil {
		return err
	}
	// 4. Copy frontend scaffold (no rename needed)
	if err := copySubtree(Templates, "templates/frontend", filepath.Join(root, "frontend"), config); err != nil {
		return err
	}
	// 5. Regenerate root Makefile
	makefile, err := fs.ReadFile(Templates, "templates/Makefile")
	if err != nil {
		return err
	}
	t, err := template.New("Makefile").Delims("[[", "]]").Parse(string(makefile))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, config); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "Makefile"), buf.Bytes(), 0644)
}
