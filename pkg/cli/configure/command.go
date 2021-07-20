package configure

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Appends the current config file in a sigrun repo to the update chain and signs it with the previous config file in the update chain",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return nil
		},
	}
}
