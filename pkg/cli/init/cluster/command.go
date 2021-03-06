package cluster

import (
	"github.com/devopstoday11/sigrun/pkg/controller"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Initializes a kubernetes cluster to be a sigrun consumer",
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		return controller.New().Init()
	}

	return cmd
}
