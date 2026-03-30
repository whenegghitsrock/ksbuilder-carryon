package generator

import (
	"strings"
	"testing"

	"github.com/kubesphere/ksbuilder/pkg/extension/spec"
)

func TestExtensionYAMLInstallationMode(t *testing.T) {
	frontendOnly := &spec.Spec{
		Name: "x", Version: "0.1.0", Mode: spec.ModeStandard,
		Capabilities: spec.Capabilities{Frontend: true, Backend: false},
		Metadata:     spec.Metadata{Category: "c"},
		Permissions:  spec.PermDefault,
	}
	out, err := ExtensionYAML(frontendOnly)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "installationMode: HostOnly") {
		t.Errorf("frontend-only: want installationMode HostOnly, got %q", string(out))
	}

	withBackend := &spec.Spec{
		Name: "y", Version: "0.1.0", Mode: spec.ModeStandard,
		Capabilities: spec.Capabilities{Frontend: false, Backend: true},
		Metadata:     spec.Metadata{Category: "c"},
		Permissions:  spec.PermDefault,
	}
	out, err = ExtensionYAML(withBackend)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "installationMode: Multicluster") {
		t.Errorf("backend present: want installationMode Multicluster, got %q", string(out))
	}
}

func TestExtensionYAMLForChartMode(t *testing.T) {
	out, err := ExtensionYAMLForChartMode("my-app")
	if err != nil {
		t.Fatalf("ExtensionYAMLForChartMode: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "name: my-app") {
		t.Errorf("expected name: my-app, got %q", s)
	}
	if !strings.Contains(s, "demoauthor") {
		t.Error("expected demoauthor in output")
	}
	if !strings.Contains(s, "apiVersion:") {
		t.Error("expected apiVersion (lowercase)")
	}
	if strings.Contains(s, "dependencies:") {
		t.Error("app/simple extension.yaml should not have dependencies")
	}
}
