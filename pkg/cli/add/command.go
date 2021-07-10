package add

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/config"

	"github.com/devopstoday11/sigrun/pkg/policy"

	kyvernoV1 "github.com/kyverno/kyverno/pkg/api/kyverno/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"

	kyvernoclient "github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Adds the given sigrun repository to the list of allowed producers in the cluster",
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
			}

			// add repos to sigrun-repos annotation
			var repos []string
			err = json.NewDecoder(strings.NewReader(cpol.Annotations["sigrun-repos"])).Decode(&repos)
			if err != nil {
				return err
			}
			repos = append(repos, args...)
			reposRaw, err := json.Marshal(repos)
			if err != nil {
				return err
			}
			cpol.Annotations["sigrun-repos"] = string(reposRaw)

			// add image verification rule for each image from config for each repo
			repoToConfig, err := config.ReadRepos(args...)
			if err != nil {
				return err
			}
			verifyImages := cpol.Spec.Rules[0].VerifyImages
			for _, conf := range repoToConfig {
				for _, confImg := range conf.Images {
					verifyImages = append(verifyImages, &kyvernoV1.ImageVerification{
						Image: confImg,
						Key:   conf.PublicKey,
					})
				}
			}
			cpol.Spec.Rules[0].VerifyImages = verifyImages

			_, err = kClient.KyvernoV1().ClusterPolicies().Update(ctx, cpol, v1.UpdateOptions{})
			if err != nil {
				return err
			}

			return nil
		},
	}
}
