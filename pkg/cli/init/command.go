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
			var name string
			_, err := fmt.Scanf("%s", &name)
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

			fmt.Println("Please enter the mode of operation\nModes of operation - 'keyless','default'")
			var mode string
			_, err = fmt.Scanf("%s", &mode)
			if err != nil {
				return err
			}

			if mode != "keyless" {
				keys, err := cosign.GenerateKeyPair(func(b bool) ([]byte, error) {
					return cosignCLI.GetPass(b)
				})
				if err != nil {
					return err
				}

				return config.NewDefaultConfig(name, string(keys.PublicBytes), string(keys.PrivateBytes), images).InitializeRepository()
			} else {
				fmt.Println("Please enter the email id's of the maintainers")
				var emailsLine string
				_, err = fmt.Scanf("%s", &emailsLine)
				if err != nil {
					return err
				}
				emails := strings.Split(emailsLine, ",")

				return config.NewKeylessConfig(name, emails, images).InitializeRepository()
			}
		},
	}

	cmd.AddCommand(cluster.Command())

	return cmd
}
