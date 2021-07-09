package init

import (
	"fmt"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/cli/config"
	cosignCLI "github.com/sigstore/cosign/cmd/cosign/cli"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initializes repository to use sigrun by creating a sigrun-config.yaml file",
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			fmt.Println("Please list all the container registry paths of images that need to be signed by sigrun")
			var imagePathsLine string
			_, err = fmt.Scanf("%s", &imagePathsLine)
			if err != nil {
				return err
			}
			images := strings.Split(imagePathsLine, ",")

			keys, err := cosign.GenerateKeyPair(cosignCLI.GetPass)
			if err != nil {
				return err
			}

			return config.Create(&config.Config{
				PublicKey:  string(keys.PublicBytes),
				PrivateKey: string(keys.PrivateBytes),
				Images:     images,
			})
		},
	}
}
