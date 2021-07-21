package add

import (
	"context"
	"encoding/base64"
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

type RepoMetaData struct {
	ChainNo   int64
	Path      string
	PublicKey string
}

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Adds a sigrun repo to the policy agent. The config file of the sigrun repo is parsed and the policy agent is update according to the config.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			err = validateAddInput(args...)
			if err != nil {
				return err
			}

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

			pathToConfig, err := config.ReadRepos(args...)
			if err != nil {
				return err
			}

			pathToGUID := make(map[string]string)
			for path := range pathToConfig {
				guid, err := config.GetGUID(path)
				if err != nil {
					return err
				}
				pathToGUID[path] = guid
			}

			// add repos to sigrun-repos annotation
			sigrunReposJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos-metadata"])
			if err != nil {
				return err
			}
			guidToRepoMeta := make(map[string]*RepoMetaData)
			_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)

			pubKToGUID := make(map[string]string)
			for guid, repoMD := range guidToRepoMeta {
				pubKToGUID[repoMD.PublicKey] = guid
			}

			for path, conf := range pathToConfig {
				if guidToRepoMeta[pathToGUID[path]] != nil {
					return fmt.Errorf("sigrun repo at " + path + " with guid " + pathToGUID[path] + " has already been added")
				}

				if guid := pubKToGUID[conf.PublicKey]; guid != "" {
					return fmt.Errorf("sigrun repo at " + path + " has the same public key as a sigrun repo that has already been added with guid " + guid)
				}

				guidToRepoMeta[pathToGUID[path]] = &RepoMetaData{
					ChainNo:   conf.ChainNo,
					Path:      path,
					PublicKey: conf.PublicKey,
				}
			}
			guidToRepoRaw, err := json.Marshal(guidToRepoMeta)
			if err != nil {
				return err
			}
			cpol.Annotations["sigrun-repos-metadata"] = base64.StdEncoding.EncodeToString(guidToRepoRaw)

			// add image verification rule for each image from config for each repo
			verifyImages := cpol.Spec.Rules[0].VerifyImages
			for _, conf := range pathToConfig {
				for _, confImg := range conf.Images {
					verifyImages = append(verifyImages, &kyvernoV1.ImageVerification{
						Image: confImg + "*",
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

func validateAddInput(args ...string) error {
	for i, arg := range args {
		if arg == "" {
			return fmt.Errorf("empty path found at position " + fmt.Sprint(i+1))
		}
	}

	return nil
}
