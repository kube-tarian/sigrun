package cli

import (
	"os"

	"github.com/devopstoday11/sigrun/pkg/cli/generate"
	"github.com/devopstoday11/sigrun/pkg/cli/remove"

	"github.com/devopstoday11/sigrun/pkg/cli/verify"

	"github.com/devopstoday11/sigrun/pkg/cli/add"
	audit_scan "github.com/devopstoday11/sigrun/pkg/cli/audit-scan"
	initCmd "github.com/devopstoday11/sigrun/pkg/cli/init"
	"github.com/devopstoday11/sigrun/pkg/cli/list"
	"github.com/devopstoday11/sigrun/pkg/cli/sign"
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

	cli.AddCommand(initCmd.Command(), list.Command(), add.Command(), sign.Command(), remove.Command(), verify.Command(), generate.Command(), audit_scan.Command())

	if err := cli.Execute(); err != nil {
		return err
	}

	return nil
}
