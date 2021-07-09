package sign

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/cli/config"

	"github.com/pkg/errors"
	cosignCLI "github.com/sigstore/cosign/cmd/cosign/cli"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {

	// TODO support annotations
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Signs the images related to this repo using cosign",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			conf, err := config.Read()
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
