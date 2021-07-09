package init

import "github.com/spf13/cobra"

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initializes the cluster to use sigrun",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return nil
		},
	}
}
