package extension

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kubesphere/ksbuilder/pkg/extension/spec"
)

func TestInferCapabilities_NoExtensionYaml(t *testing.T) {
	dir := t.TempDir()
	cap, err := InferCapabilities(dir)
	if err != nil {
		t.Fatalf("InferCapabilities: %v", err)
	}
	if cap != nil {
		t.Errorf("InferCapabilities should return nil for empty dir, got %+v", cap)
	}
}

func TestInferCapabilities_FrontendOnly(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "fe")
	s := &spec.Spec{
		Name:         "fe",
		Version:      "0.1.0",
		Mode:         spec.ModeStandard,
		Capabilities: spec.Capabilities{Frontend: true, Backend: false},
		Metadata:     spec.Metadata{Category: "database"},
		Permissions:  spec.PermDefault,
	}
	if err := CreateFromSpec(root, s); err != nil {
		t.Fatalf("CreateFromSpec: %v", err)
	}
	cap, err := InferCapabilities(root)
	if err != nil {
		t.Fatalf("InferCapabilities: %v", err)
	}
	if cap == nil {
		t.Fatal("InferCapabilities returned nil for valid extension")
	}
	if cap.Name != "fe" {
		t.Errorf("Name = %q, want fe", cap.Name)
	}
	if !cap.HasFrontend {
		t.Error("HasFrontend = false, want true")
	}
	if cap.HasBackend {
		t.Error("HasBackend = true, want false")
	}
}

func TestExtendAddBackend(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "fe-ext")
	s := &spec.Spec{
		Name:         "fe-ext",
		Version:      "0.1.0",
		Mode:         spec.ModeStandard,
		Capabilities: spec.Capabilities{Frontend: true, Backend: false},
		Metadata:     spec.Metadata{Category: "database"},
		Permissions:  spec.PermDefault,
	}
	if err := CreateFromSpec(root, s); err != nil {
		t.Fatalf("CreateFromSpec: %v", err)
	}
	if err := ExtendAddBackend(root); err != nil {
		t.Fatalf("ExtendAddBackend: %v", err)
	}
	extData, err := os.ReadFile(filepath.Join(root, "extension.yaml"))
	if err != nil {
		t.Fatalf("read extension.yaml: %v", err)
	}
	if !strings.Contains(string(extData), "installationMode: Multicluster") {
		t.Error("after ExtendAddBackend, installationMode should be Multicluster")
	}
	verifyFiles(t, root, true, true)
	verifyPackage(t, root)
}

func TestExtendAddFrontend(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "be-ext")
	s := &spec.Spec{
		Name:         "be-ext",
		Version:      "0.1.0",
		Mode:         spec.ModeStandard,
		Capabilities: spec.Capabilities{Frontend: false, Backend: true},
		Metadata:     spec.Metadata{Category: "observability"},
		Permissions:  spec.PermDefault,
	}
	if err := CreateFromSpec(root, s); err != nil {
		t.Fatalf("CreateFromSpec: %v", err)
	}
	if err := ExtendAddFrontend(root); err != nil {
		t.Fatalf("ExtendAddFrontend: %v", err)
	}
	extData, err := os.ReadFile(filepath.Join(root, "extension.yaml"))
	if err != nil {
		t.Fatalf("read extension.yaml: %v", err)
	}
	if !strings.Contains(string(extData), "installationMode: Multicluster") {
		t.Error("after ExtendAddFrontend, installationMode should stay Multicluster (backend present)")
	}
	verifyFiles(t, root, true, true)
	verifyPackage(t, root)
}
