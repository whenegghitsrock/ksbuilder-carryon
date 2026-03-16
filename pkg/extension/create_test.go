package extension

import (
	"os"
	"path/filepath"
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

	verifyFiles(t, root, false, true)
	verifyPackage(t, root)
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
	// Must have frontend scaffold
	for _, p := range []string{"frontend/Dockerfile", "frontend/index.html", "frontend/Makefile"} {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}
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
	}
	if wantBackend {
		if _, err := os.Stat(filepath.Join(root, "charts", "backend", "Chart.yaml")); err != nil {
			t.Errorf("missing charts/backend: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "backend", "Dockerfile")); err != nil {
			t.Errorf("missing backend scaffold: %v", err)
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
