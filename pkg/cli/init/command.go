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
	}

	var repoPath string
	cmd.Flags().StringVar(&repoPath, "path", "./", "Path to the sigrun repository")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
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

		fmt.Printf("Please enter the mode of operation, default - '%v'\nModes of operation - '%v','%v'\n", config.CONFIG_MODE_KEYLESS, config.CONFIG_MODE_KEYLESS, config.CONFIG_MODE_KEYPAIR)
		var mode string
		_, err = fmt.Scanf("%s", &mode)
		if err != nil {
			return err
		}

		if mode == config.CONFIG_MODE_KEYPAIR {
			keys, err := cosign.GenerateKeyPair(func(b bool) ([]byte, error) {
				return cosignCLI.GetPass(b)
			})
			if err != nil {
				return err
			}

			return config.NewKeypairConfig(name, string(keys.PublicBytes), string(keys.PrivateBytes), images).InitializeRepository(repoPath)
		} else {
			fmt.Println("Please enter the email id's of the maintainers")
			var emailsLine string
			_, err = fmt.Scanf("%s", &emailsLine)
			if err != nil {
				return err
			}
			emails := strings.Split(emailsLine, ",")

			return config.NewKeylessConfig(name, emails, images).InitializeRepository(repoPath)
		}
	}

	cmd.AddCommand(cluster.Command())

	return cmd
}
