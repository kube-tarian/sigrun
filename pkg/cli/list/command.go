package list

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/devopstoday11/sigrun/pkg/policy"

	kyvernoV1 "github.com/kyverno/kyverno/pkg/api/kyverno/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"

	kyvernoclient "github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List existing repos using image verification",
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
			var cpol *kyvernoV1.ClusterPolicy
			cpol, err = kClient.KyvernoV1().ClusterPolicies().Get(ctx, "sigrun-verify", v1.GetOptions{})
			if err != nil {
				fmt.Println("Could not find policy, creating new policy...")
				cpol = policy.New()
				cpol, err = kClient.KyvernoV1().ClusterPolicies().Create(ctx, cpol, v1.CreateOptions{})
				if err != nil {
					return err
				}
			}
			sigrunRepos, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos"])
			if err != nil {
				return err
			}
			fmt.Println("Sigrun-repos - " + string(sigrunRepos))
			return nil
		},
	}
}
