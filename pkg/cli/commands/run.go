package commands

import (
	"github.com/devopstoday11/sigrun/pkg/cli/commands/add"
	initCmd "github.com/devopstoday11/sigrun/pkg/cli/commands/init"
	"github.com/devopstoday11/sigrun/pkg/cli/commands/sign"
	"github.com/devopstoday11/sigrun/pkg/cli/commands/update"
	"github.com/spf13/cobra"
)

func Run() error {
	cli := &cobra.Command{
		Use: "sigrun",
		Short: "Sign your artifacts source code or container images using " +
			"Sigstore tools, Save the Signatures you want to use, and Validate & Control the " +
			"deployments to allow only the known Signatures.",
	}

	cli.AddCommand(initCmd.Command(), add.Command(), sign.Command(), update.Command())

	if err := cli.Execute(); err != nil {
		return err
	}

	return nil
}
