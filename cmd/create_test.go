package cmd

import (
	"strings"
	"testing"
)

func TestCreateStandardTemplateTypeMapping(t *testing.T) {
	cases := []struct {
		idx  int
		want string
	}{
		{idx: 0, want: "standard"},
		{idx: 1, want: "frontend"},
		{idx: 2, want: "backend"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			got := overlayTypeFromTemplateIndex(tc.idx)
			if got != tc.want {
				t.Fatalf("overlayTypeFromTemplateIndex(%d) = %q, want %q", tc.idx, got, tc.want)
			}
		})
	}
}

func TestCreateSuccessHintTextStable(t *testing.T) {
	hint := createSuccessHint(false, false, false, "")
	if !strings.Contains(hint, "ksbuilder package .") {
		t.Fatalf("create success hint should mention packaging command, got %q", hint)
	}
}

func TestCreateSuccessHintIncludesCdStep(t *testing.T) {
	hint := createSuccessHint(true, true, true, "my-extension")
	if !strings.Contains(hint, "cd my-extension") {
		t.Fatalf("create success hint should mention cd step, got %q", hint)
	}
}

func TestCreateSuccessHintStandardMode(t *testing.T) {
	hint := createSuccessHint(true, true, true, "my-extension")
	if !strings.Contains(hint, "make build-frontend build-backend") {
		t.Errorf("standard mode hint should mention make build: %q", hint)
	}
	if !strings.Contains(hint, "make push") {
		t.Errorf("standard mode hint should mention make push: %q", hint)
	}
	if !strings.Contains(hint, "ksbuilder package .") {
		t.Errorf("standard mode hint should mention package: %q", hint)
	}
	if !strings.Contains(hint, "ksbuilder publish .") {
		t.Errorf("standard mode hint should mention publish: %q", hint)
	}
	if !strings.Contains(hint, "yarn create ks-project") {
		t.Errorf("standard mode with frontend should mention create-ks-project: %q", hint)
	}
	if !strings.Contains(hint, "frontend/README.md") {
		t.Errorf("standard mode with frontend should point to frontend README: %q", hint)
	}
}
