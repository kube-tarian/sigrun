package remove

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/config"

	"github.com/devopstoday11/sigrun/pkg/policy"
	kyvernoV1 "github.com/kyverno/kyverno/pkg/api/kyverno/v1"
	kyvernoclient "github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Removes a sigrun repo from the policy agent. Updates the policy agent by removing all data related to sigrun repo.",
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

			sigrunReposJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos-metadata"])
			if err != nil {
				return err
			}
			guidToRepoMeta := make(map[string]*policy.RepoMetaData)
			_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)
			verifyImages := cpol.Spec.Rules[0].VerifyImages
			for _, path := range args {
				guid, err := config.GetGUID(path)
				if err != nil {
					return err
				}

				if guidToRepoMeta[guid] == nil {
					return fmt.Errorf("sigrun repo at " + path + " with guid " + guid + " does not exist ")
				}

				var buf []*kyvernoV1.ImageVerification
				for _, vi := range verifyImages {
					if vi.Key != guidToRepoMeta[guid].PublicKey {
						buf = append(buf, vi)
					}
				}
				verifyImages = buf

				delete(guidToRepoMeta, guid)
			}
			guidToRepoRaw, err := json.Marshal(guidToRepoMeta)
			if err != nil {
				return err
			}
			cpol.Annotations["sigrun-repos-metadata"] = base64.StdEncoding.EncodeToString(guidToRepoRaw)
			cpol.Spec.Rules[0].VerifyImages = verifyImages

			_, err = kClient.KyvernoV1().ClusterPolicies().Update(ctx, cpol, v1.UpdateOptions{})
			if err != nil {
				return err
			}

			return nil
		},
	}
}
