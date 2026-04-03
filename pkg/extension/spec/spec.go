package spec

import (
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

// SpecFileName is the declarative spec file name in extension root.
const SpecFileName = ".ksbuilder.yaml"

// Mode is the extension creation mode.
type Mode string

const (
	ModeStandard Mode = "standard" // frontend and/or backend from scratch
	ModeApp      Mode = "app"      // from helm chart, Application CR
	ModeSimple   Mode = "simple"   // from helm chart, link/iframe
)

// PermissionsLevel defines permission presets.
type PermissionsLevel string

const (
	PermDefault PermissionsLevel = "default" // standard ClusterRole + Role
	PermApp     PermissionsLevel = "app"     // + Application, ConfigMap, RBAC
	PermSimple  PermissionsLevel = "simple"  // + JSBundle, ConfigMap, clusterroles
)

// Capabilities describes which components to include (standard mode).
type Capabilities struct {
	Frontend bool `yaml:"frontend"`
	Backend  bool `yaml:"backend"`
}

// Metadata holds extension metadata.
type Metadata struct {
	DisplayName map[string]string `yaml:"displayName"` // zh, en
	Description map[string]string `yaml:"description"`
	Category    string            `yaml:"category"`
	Keywords    []string          `yaml:"keywords,omitempty"`
	Author      string            `yaml:"author"`
	Email       string            `yaml:"email"`
	URL         string            `yaml:"url"`
}

// Spec is the declarative extension specification.
type Spec struct {
	Name         string           `yaml:"name"`
	Version      string           `yaml:"version"`
	Mode         Mode             `yaml:"mode"`
	Capabilities Capabilities     `yaml:"capabilities,omitempty"` // for mode=standard
	Metadata     Metadata         `yaml:"metadata"`
	Permissions  PermissionsLevel `yaml:"permissions"`

	// App-only fields (when mode=app)
	AppName        string `yaml:"appName,omitempty"`
	AppVersionName string `yaml:"appVersionName,omitempty"`
	ZipName        string `yaml:"zipName,omitempty"`
	Maintainers    string `yaml:"maintainers,omitempty"`
	AppHome        string `yaml:"appHome,omitempty"`
	Abstraction    string `yaml:"abstraction,omitempty"`
}

// HasFrontend returns true if the spec includes frontend.
func (s *Spec) HasFrontend() bool {
	switch s.Mode {
	case ModeStandard:
		return s.Capabilities.Frontend
	case ModeApp, ModeSimple:
		return false
	default:
		return false
	}
}

// HasBackend returns true if the spec includes backend.
func (s *Spec) HasBackend() bool {
	switch s.Mode {
	case ModeStandard:
		return s.Capabilities.Backend
	case ModeApp, ModeSimple:
		return false
	default:
		return false
	}
}

// Read loads a spec from the given directory (looks for .ksbuilder.yaml).
func Read(dir string) (*Spec, error) {
	path := filepath.Join(dir, SpecFileName)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var spec Spec
	if err := yaml.Unmarshal(b, &spec); err != nil {
		return nil, fmt.Errorf("parse %s: %w", SpecFileName, err)
	}
	return &spec, nil
}

// Write persists the spec to the given directory.
func (s *Spec) Write(dir string) error {
	b, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, SpecFileName)
	return os.WriteFile(path, b, 0644)
}

// Exists checks if a spec file exists in dir.
func Exists(dir string) bool {
	path := filepath.Join(dir, SpecFileName)
	_, err := os.Stat(path)
	return err == nil
}
