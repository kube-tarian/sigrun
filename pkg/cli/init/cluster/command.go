package cluster

import (
	"fmt"

	"github.com/devopstoday11/sigrun/pkg/controller"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Initializes a kubernetes cluster to be a sigrun consumer",
	}

	var controllerF string
	cmd.Flags().StringVar(&controllerF, "controller", "sigrun", "specify the controller you would like to initialize the cluster with")
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {

		cont, err := controller.GetController()
		if err == nil {
			return fmt.Errorf("cluster has already been intialized with controller of type - " + cont.Type())
		}

		if controllerF == "kyverno" {
			return controller.NewKyvernoController().Init()
		} else {
			return controller.NewSigrunController().Init()
		}
	}

	return cmd
}
