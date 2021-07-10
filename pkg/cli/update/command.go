package update

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
		Use:   "update",
		Short: "Updates the version of the registered sigrun repos in the cluster",
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

			var repos []string
			err = json.NewDecoder(strings.NewReader(cpol.Annotations["sigrun-repos"])).Decode(&repos)
			if err != nil {
				return err
			}

			reposToCommit := make(map[string]string)
			for _, repoS := range repos {
				parts := strings.Split(repoS, "@")
				reposToCommit[parts[0]] = parts[1]
			}
			for _, repoS := range args {
				parts := strings.Split(repoS, "@")
				err = removeRepoImageVerification(parts[0]+"@"+reposToCommit[parts[0]], cpol)
				if err != nil {
					return err
				}
				delete(reposToCommit, parts[0])
				err = addRepoImageVerification(repoS, cpol)
				if err != nil {
					return err
				}
				reposToCommit[parts[0]] = parts[1]
			}

			var updatedRepos []string
			for repo, commit := range reposToCommit {
				updatedRepos = append(updatedRepos, repo+"@"+commit)
			}

			reposRaw, err := json.Marshal(updatedRepos)
			if err != nil {
				return err
			}
			cpol.Annotations["sigrun-repos"] = string(reposRaw)

			_, err = kClient.KyvernoV1().ClusterPolicies().Update(ctx, cpol, v1.UpdateOptions{})
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func removeRepoImageVerification(repo string, cpol *kyvernoV1.ClusterPolicy) error {
	repoToConfig, err := config.ReadRepos(repo)
	if err != nil {
		return err
	}
	verifyImages := cpol.Spec.Rules[0].VerifyImages
	var pubkeyToImages = make(map[string][]string)
	for _, verifyImage := range verifyImages {
		pubkeyToImages[verifyImage.Key] = append(pubkeyToImages[verifyImage.Key], verifyImage.Image)
	}

	for _, conf := range repoToConfig {
		delete(pubkeyToImages, conf.PublicKey)
	}

	verifyImages = []*kyvernoV1.ImageVerification{}
	for key, images := range pubkeyToImages {
		for _, image := range images {
			verifyImages = append(verifyImages, &kyvernoV1.ImageVerification{
				Image: image,
				Key:   key,
			})
		}
	}

	cpol.Spec.Rules[0].VerifyImages = verifyImages

	return nil
}

func addRepoImageVerification(repo string, cpol *kyvernoV1.ClusterPolicy) error {
	repoToConfig, err := config.ReadRepos(repo)
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

	return nil
}
