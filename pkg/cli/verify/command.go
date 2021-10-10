package verify

import (
	transperencylog "github.com/devopstoday11/sigrun/pkg/cli/verify/transperency-log"
	"github.com/devopstoday11/sigrun/pkg/config"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verifies the signature of the given images",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			conf, err := config.ReadRepositoryConfig()
			if err != nil {
				return err
			}

			for _, img := range args {
				err = conf.VerifyImage(img)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.AddCommand(transperencylog.Command())

	return cmd
}
