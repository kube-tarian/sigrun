package remove

import (
	"github.com/devopstoday11/sigrun/pkg/controller"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Removes a sigrun repo from the policy agent. Updates the policy agent by removing all data related to sigrun repo.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			cont, err := controller.GetController()
			if err != nil {
				return nil
			}

			return cont.Remove(args...)
		},
	}
}
