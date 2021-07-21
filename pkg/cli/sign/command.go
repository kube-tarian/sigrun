package sign

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/config"

	"github.com/pkg/errors"
	cosignCLI "github.com/sigstore/cosign/cmd/cosign/cli"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Signs all images related to a sigrun repo and pushes the signatures to the corresponding container registry",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			conf, err := config.Get(config.FILE_NAME)
			if err != nil {
				return err
			}

			tempPrivKeyFile, err := ioutil.TempFile("", "priv-key")
			if err != nil {
				return err
			}
			defer os.Remove(tempPrivKeyFile.Name())
			_, err = io.Copy(tempPrivKeyFile, strings.NewReader(conf.PrivateKey))
			if err != nil {
				return err
			}

			so := cosignCLI.SignOpts{
				KeyRef: tempPrivKeyFile.Name(),
				Pf:     cosignCLI.GetPass,
			}
			ctx := context.Background()
			for _, img := range conf.Images {
				if err := cosignCLI.SignCmd(ctx, so, img, true, "", false, false); err != nil {
					return errors.Wrapf(err, "signing %s", img)
				}
			}
			return nil
		},
	}

	return cmd
}
