package configure

import (
	"fmt"

	cosignCLI "github.com/sigstore/cosign/cmd/cosign/cli"

	"github.com/devopstoday11/sigrun/pkg/config"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Appends the current config file in a sigrun repo to the update chain and signs it with the previous config file in the update chain",
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			conf, err := config.Get(config.FILE_NAME)
			if err != nil {
				return err
			}

			oldConf, err := config.GetChainHead()
			if err != nil {
				return err
			}

			isSame, err := config.IsSame(conf, oldConf)
			if err != nil {
				return err
			}

			if isSame {
				return fmt.Errorf("config has not changed")
			}

			conf.ChainNo = oldConf.ChainNo + 1

			password, err := cosignCLI.GetPass(true)
			if err != nil {
				return err
			}
			sig, err := oldConf.Sign(string(password), conf)
			if err != nil {
				return err
			}
			conf.Signature = sig

			err = config.Set(config.FILE_NAME, conf)
			if err != nil {
				return err
			}

			return config.Set(".sigrun/"+fmt.Sprint(conf.ChainNo)+".json", conf)
		},
	}
}
