package update

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/config"
	"github.com/devopstoday11/sigrun/pkg/policy"
	kyvernoclient "github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Updates all the sigrun repos present in the policy agent. Checks the sigrun repos for updates, verifies the updates and updates the policy agent to handle the updates",
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

			// add repos to sigrun-repos annotation
			sigrunReposJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos-metadata"])
			if err != nil {
				return err
			}
			guidToRepoMeta := make(map[string]*policy.RepoMetaData)
			_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)
			for guid, md := range guidToRepoMeta {
				confMap, err := config.ReadRepos(md.Path)
				if err != nil {
					return err
				}
				conf := confMap[md.Path]

				if conf.ChainNo > md.ChainNo {
					fmt.Println("verifying sigrun repo with guid " + guid + " and name " + md.Name + " from chain no " + fmt.Sprint(md.ChainNo) + " to " + fmt.Sprint(conf.ChainNo))
					err = config.VerifyChain(md.PublicKey, md.Path, md.ChainNo, conf)
					if err != nil {
						return err
					}

					fmt.Println("updating sigrun repo with guid " + guid + " and name " + md.Name + " from chain no " + fmt.Sprint(md.ChainNo) + " to " + fmt.Sprint(conf.ChainNo))
					cpol, err = policy.RemoveRepo(cpol, guid)
					if err != nil {
						return err
					}

					cpol, err = policy.AddRepo(cpol, guid, md.Path, conf)
					if err != nil {
						return err
					}
				} else {
					fmt.Println("sigrun repo with guid " + guid + " and name " + md.Name + " is already upto date")
				}
			}

			_, err = kClient.KyvernoV1().ClusterPolicies().Update(ctx, cpol, v1.UpdateOptions{})
			if err != nil {
				return err
			}

			return nil
		},
	}
}
