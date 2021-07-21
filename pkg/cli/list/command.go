package list

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/devopstoday11/sigrun/pkg/policy"
	"github.com/tidwall/pretty"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"

	kyvernoclient "github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists metadata about sigrun repos that have been added",
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
			if err != nil {
				return err
			}

			kClient, err := kyvernoclient.NewForConfig(kRestConf)
			if err != nil {
				return err
			}

			ctx := context.Background()
			cpol, err := kClient.KyvernoV1().ClusterPolicies().Get(ctx, policy.NAME, v1.GetOptions{})
			if err != nil {
				return err
			}
			sigrunRepos, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos-metadata"])
			if err != nil {
				return err
			}

			fmt.Println("Sigrun-repos:\n" + string(pretty.Pretty(sigrunRepos)))
			return nil
		},
	}
}
