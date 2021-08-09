package update

import (
	"github.com/devopstoday11/sigrun/pkg/controller"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Updates all the sigrun repos present in the policy agent. Checks the sigrun repos for updates, verifies the updates and updates the policy agent to handle the updates",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cont, err := controller.GetController()
			if err != nil {
				return err
			}

			return cont.Update()
		},
	}
}
