package update

import (
	"github.com/devopstoday11/sigrun/pkg/controller"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Updates all the sigrun repos present in the policy agent. Checks the sigrun repos for updates, verifies the updates and updates the policy agent to handle the updates",
	}
	var controllerF string
	cmd.Flags().StringVar(&controllerF, "controller", "sigrun", "specify the controller you would like to use")
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		cont, err := controller.GetController(controllerF)
		if err != nil {
			return err
		}

		return cont.Update()
	}

	return cmd
}
