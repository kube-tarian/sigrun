package update

import "github.com/spf13/cobra"

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Updates the given sigrun repository to its latest configuration in the cluster",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return nil
		},
	}
}
