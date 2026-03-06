package utils

import (
	"os"
	"path/filepath"
)

// ResolveKubeconfig returns the kubeconfig path in order: flagValue > KUBECONFIG env > ~/.kube/config
func ResolveKubeconfig(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if v := os.Getenv("KUBECONFIG"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kube", "config")
}
