package sign

import (
	"github.com/devopstoday11/sigrun/pkg/config"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Signs all images related to a sigrun repo and pushes the signatures to the corresponding container registry",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			conf, err := config.ReadRepositoryConfig()
			if err != nil {
				return err
			}

			return conf.SignImages()
		},
	}

	return cmd
}
