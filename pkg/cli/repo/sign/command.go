package sign

import "github.com/spf13/cobra"

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "sign",
		Short: "Signs the images related to this repo using cosign",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return nil
		},
	}
}
