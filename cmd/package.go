package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/kubesphere/ksbuilder/pkg/extension"
)

type packageOptions struct {
	skipDependencyUpdate bool
}

func defaultPackageOptions() *packageOptions {
	return &packageOptions{}
}

func packageExtensionCmd() *cobra.Command {
	o := defaultPackageOptions()

	cmd := &cobra.Command{
		Use:   "package",
		Short: "package an extension",
		Args:  cobra.ExactArgs(1),
		RunE:  o.packageCmd,
	}
	cmd.Flags().BoolVar(&o.skipDependencyUpdate, "skip-dependency-update", false, "skip running helm dependency update before packaging")
	return cmd
}

func (o *packageOptions) packageCmd(_ *cobra.Command, args []string) error {
	pwd, _ := os.Getwd()
	p := args[0]
	if !path.IsAbs(p) {
		p = path.Join(pwd, p)
	}
	fmt.Printf("package extension %s\n", args[0])

	chartFilename, err := extension.PackageToPath(p, pwd, o.skipDependencyUpdate)
	if err != nil {
		return err
	}
	fmt.Printf("package saved to %s\n", chartFilename)
	return nil
}
