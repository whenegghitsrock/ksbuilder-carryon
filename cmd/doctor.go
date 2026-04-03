package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kubesphere/ksbuilder/pkg/utils"
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

func checkHelm() checkResult {
	_, err := exec.LookPath("helm")
	if err != nil {
		return checkResult{ok: false, msg: "Helm not found", hint: "install Helm: https://helm.sh/docs/intro/install"}
	}
	out, err := exec.Command("helm", "version", "--short").Output()
	if err != nil {
		return checkResult{ok: false, msg: "helm version failed", hint: "run 'helm version'"}
	}
	return checkResult{ok: true, msg: "Helm " + strings.TrimSpace(string(out))}
}

func checkKubectl() checkResult {
	_, err := exec.LookPath("kubectl")
	if err != nil {
		return checkResult{ok: false, msg: "kubectl not found", hint: "install kubectl: https://kubernetes.io/docs/tasks/tools/"}
	}
	out, err := exec.Command("kubectl", "version", "--client", "--short").Output()
	if err != nil {
		return checkResult{ok: true, msg: "kubectl (version check skipped)"}
	}
	ver := strings.TrimSpace(string(out))
	if len(ver) > 0 && ver[0] == 'v' {
		return checkResult{ok: true, msg: "kubectl " + ver}
	}
	return checkResult{ok: true, msg: "kubectl installed"}
}

func checkKsbuilder(version string) checkResult {
	return checkResult{ok: true, msg: "ksbuilder " + version}
}

func checkKubeconfig(cmd *cobra.Command) checkResult {
	flagVal, _ := cmd.Root().PersistentFlags().GetString("kubeconfig")
	path := utils.ResolveKubeconfig(flagVal)
	_, err := utils.BuildClientFromFlags(path)
	if err != nil {
		return checkResult{
			ok:   false,
			msg:  fmt.Sprintf("kubeconfig: %s (%v)", path, err),
			hint: "check cluster is running and kubeconfig points to correct context",
		}
	}
	return checkResult{ok: true, msg: fmt.Sprintf("kubeconfig: %s (cluster reachable)", path)}
}

func runDoctor(cmd *cobra.Command, version string) error {
	fmt.Println("ksbuilder doctor")
	fmt.Println()

	checks := []checkResult{
		checkGo(),
		checkHelm(),
		checkKubectl(),
		checkKubeconfig(cmd),
		checkKsbuilder(version),
	}

	anyFail := false
	for _, r := range checks {
		if r.ok {
			fmt.Println("✓", r.msg)
		} else {
			fmt.Println("✗", r.msg)
			if r.hint != "" {
				fmt.Println("  Hint:", r.hint)
			}
			anyFail = true
		}
	}
	fmt.Println()
	if anyFail {
		fmt.Println("Some checks failed.")
		return fmt.Errorf("doctor found issues")
	}
	fmt.Println("All checks passed.")
	return nil
}
