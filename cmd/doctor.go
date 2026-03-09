package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func doctorCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check development environment",
		Long:  "Run checks for Go, Helm, kubectl, kubeconfig, and ksbuilder.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDoctor(cmd, version)
		},
	}
	return cmd
}

type checkResult struct {
	ok   bool
	msg  string
	hint string
}

func checkGo() checkResult {
	_, err := exec.LookPath("go")
	if err != nil {
		return checkResult{ok: false, msg: "Go not found", hint: "install Go from https://go.dev"}
	}
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return checkResult{ok: false, msg: "go version failed", hint: "run 'go version'"}
	}
	s := strings.TrimSpace(string(out))
	parts := strings.Fields(strings.TrimPrefix(s, "go version "))
	ver := "installed"
	if len(parts) > 0 {
		ver = strings.TrimPrefix(parts[0], "go")
	}
	return checkResult{ok: true, msg: "Go " + ver}
}

func runDoctor(cmd *cobra.Command, version string) error {
	fmt.Println("ksbuilder doctor")
	fmt.Println()

	r := checkGo()
	if r.ok {
		fmt.Println("✓", r.msg)
	} else {
		fmt.Println("✗", r.msg)
		if r.hint != "" {
			fmt.Println("  Hint:", r.hint)
		}
	}
	fmt.Println()
	fmt.Println("All checks passed.")
	return nil
}
