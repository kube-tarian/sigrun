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

type Repo struct {
	ChainNo int64
	Path    string
}

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Adds a sigrun repo to the policy agent. The config file of the sigrun repo is parsed and the policy agent is update according to the config.",
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
			var newPolicy bool
			cpol, err = kClient.KyvernoV1().ClusterPolicies().Get(ctx, policy.NAME, v1.GetOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					fmt.Println("Could not find policy, creating new policy...")
					cpol = policy.New()
					newPolicy = true
				} else {
					return err
				}
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
			sigrunReposJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos"])
			if err != nil {
				return err
			}
			guidToRepo := make(map[string]*Repo)
			_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepo)
			for path, conf := range pathToConfig {
				guidToRepo[pathToGUID[path]] = &Repo{
					ChainNo: conf.ChainNo,
					Path:    path,
				}
			}
			guidToRepoRaw, err := json.Marshal(guidToRepo)
			if err != nil {
				return err
			}
			cpol.Annotations["sigrun-repos"] = base64.StdEncoding.EncodeToString(guidToRepoRaw)

			sigrunKeysJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-keys"])
			if err != nil {
				return err
			}
			guidToKeys := make(map[string]string)
			_ = json.NewDecoder(strings.NewReader(string(sigrunKeysJSON))).Decode(&guidToKeys)
			for path, conf := range pathToConfig {
				guidToKeys[pathToGUID[path]] = conf.PublicKey
			}
			guidToKeysRaw, err := json.Marshal(guidToKeys)
			if err != nil {
				return err
			}
			cpol.Annotations["sigrun-keys"] = base64.StdEncoding.EncodeToString(guidToKeysRaw)

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

			if newPolicy {
				_, err = kClient.KyvernoV1().ClusterPolicies().Create(ctx, cpol, v1.CreateOptions{})
				if err != nil {
					return err
				}
			} else {
				_, err = kClient.KyvernoV1().ClusterPolicies().Update(ctx, cpol, v1.UpdateOptions{})
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}
