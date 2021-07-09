package commands

import (
	"github.com/devopstoday11/sigrun/pkg/cli/commands/cluster/add"
	clusterinit "github.com/devopstoday11/sigrun/pkg/cli/commands/cluster/init"
	"github.com/devopstoday11/sigrun/pkg/cli/commands/cluster/update"
	repoinit "github.com/devopstoday11/sigrun/pkg/cli/commands/repo/init"
	"github.com/devopstoday11/sigrun/pkg/cli/commands/repo/sign"
	"github.com/spf13/cobra"
)

func Run() error {
	cli := &cobra.Command{
		Use: "sigrun",
		Short: "Sign your artifacts source code or container images using " +
			"Sigstore tools, Save the Signatures you want to use, and Validate & Control the " +
			"deployments to allow only the known Signatures.",
	}

	cli.AddCommand(RepoCommands(), ClusterCommands())

	if err := cli.Execute(); err != nil {
		return err
	}

	return nil
}

func RepoCommands() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "repo",
		Short: "Commands related to managing a sigrun repo",
	}

	initCmd.AddCommand(repoinit.Command(), sign.Command())

	return initCmd
}
func ClusterCommands() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "cluster",
		Short: "Commands related to managing a sigrun cluster",
	}

	initCmd.AddCommand(clusterinit.Command(), add.Command(), update.Command())

	return initCmd
}
