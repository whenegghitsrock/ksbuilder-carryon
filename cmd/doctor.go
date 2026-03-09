package cmd

import (
	"fmt"

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

func runDoctor(cmd *cobra.Command, version string) error {
	fmt.Println("ksbuilder doctor")
	fmt.Println()
	// 占位：后续 Task 实现各检查
	fmt.Println("All checks passed.")
	return nil
}
