package init

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initializes repository to use sigrun by creating a sigrun-config.yaml file",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return nil
		},
	}
}
