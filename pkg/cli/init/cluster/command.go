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

	var controllerF string
	cmd.Flags().StringVar(&controllerF, "controller", "sigrun", "specify the controller you would like to use")
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		cont, err := controller.GetController(controllerF)
		if err != nil {
			return err
		}

		return cont.Init()
	}

	return cmd
}
