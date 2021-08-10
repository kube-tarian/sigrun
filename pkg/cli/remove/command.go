package remove

import (
	"github.com/devopstoday11/sigrun/pkg/controller"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Removes a sigrun repo from the policy agent. Updates the policy agent by removing all data related to sigrun repo.",
	}

	var controllerF string
	cmd.Flags().StringVar(&controllerF, "controller", "sigrun", "specify the controller you would like to use")
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		cont, err := controller.GetController(controllerF)
		if err != nil {
			return nil
		}

		return cont.Remove(args...)
	}

	return cmd
}
