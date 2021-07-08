package add

import "github.com/spf13/cobra"

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Adds the given sigrun repository to the list of allowed producers in the cluster",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return nil
		},
	}
}
