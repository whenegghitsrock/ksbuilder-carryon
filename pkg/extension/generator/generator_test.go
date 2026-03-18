package generator

import (
	"strings"
	"testing"
)

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
