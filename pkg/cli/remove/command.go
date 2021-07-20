package remove

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Removes a sigrun repo from the policy agent. Updates the policy agent by removing all data related to sigrun repo.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return nil
		},
	}
}
