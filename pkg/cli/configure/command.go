package configure

import (
	"github.com/devopstoday11/sigrun/pkg/config"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "commit",
		Short: "Appends the current config file in a sigrun repo to the update chain and signs it with the previous config file in the update chain",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			newConf, err := config.ReadRepositoryConfig()
			if err != nil {
				return err
			}

			return newConf.CommitRepositoryUpdate()
		},
	}
}
