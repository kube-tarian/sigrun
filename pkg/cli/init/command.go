package init

import (
	"fmt"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/cli/init/cluster"

	"github.com/devopstoday11/sigrun/pkg/config"
	cosignCLI "github.com/sigstore/cosign/cmd/cosign/cli"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Provides an interactive interface to create a sigrun repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Please enter the name of this sigrun repo")
			var moniker string
			_, err := fmt.Scanf("%s", &moniker)
			if err != nil {
				return err
			}

			fmt.Println("Please list all the container registry paths of images that need to be signed by sigrun")
			var imagePathsLine string
			_, err = fmt.Scanf("%s", &imagePathsLine)
			if err != nil {
				return err
			}
			images := strings.Split(imagePathsLine, ",")

			keys, err := cosign.GenerateKeyPair(func(b bool) ([]byte, error) {
				return cosignCLI.GetPass(b)
			})
			if err != nil {
				return err
			}

			return config.Create(&config.Config{
				Moniker:    moniker,
				PublicKey:  string(keys.PublicBytes),
				PrivateKey: string(keys.PrivateBytes),
				Images:     images,
			})
		},
	}

	cmd.AddCommand(cluster.Command())

	return cmd
}
