package cli

import (
	"os"

	"github.com/devopstoday11/sigrun/pkg/cli/add"
	"github.com/devopstoday11/sigrun/pkg/cli/configure"
	initCmd "github.com/devopstoday11/sigrun/pkg/cli/init"
	"github.com/devopstoday11/sigrun/pkg/cli/list"
	"github.com/devopstoday11/sigrun/pkg/cli/sign"
	"github.com/devopstoday11/sigrun/pkg/cli/update"
	cosignCLI "github.com/sigstore/cosign/cmd/cosign/cli"
	"github.com/spf13/cobra"
)

func Run() error {
	cli := &cobra.Command{
		Use: "sigrun",
		Short: "Sign your artifacts source code or container images using " +
			"Sigstore tools, Save the Signatures you want to use, and Validate & Control the " +
			"deployments to allow only the known Signatures.",
	}

	// TODO required for keyless mode
	os.Setenv(cosignCLI.ExperimentalEnv, "1")

	cli.AddCommand(initCmd.Command(), list.Command(), add.Command(), sign.Command(), update.Command(), configure.Command())

	if err := cli.Execute(); err != nil {
		return err
	}

	return nil
}
