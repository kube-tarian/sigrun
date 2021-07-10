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
				// TODO handle error here properly
				fmt.Println("Could not find policy, creating new policy...")
				cpol = policy.New()
				cpol, err = kClient.KyvernoV1().ClusterPolicies().Create(ctx, cpol, v1.CreateOptions{})
				if err != nil {
					return err
				}
			}

			//TODO add validation - should check if already exists or if pubkey already exists
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
			err = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepo)
			if err != nil {
				return err
			}
			for path, conf := range pathToConfig {
				guidToRepo[pathToGUID[path]] = &Repo{
					ChainNo: conf.ChainNo,
					Path:    path,
				}
			}
			guideToRepoRaw, err := json.Marshal(guidToRepo)
			if err != nil {
				return err
			}
			cpol.Annotations["sigrun-repos"] = string(guideToRepoRaw)

			sigrunKeysJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-keys"])
			if err != nil {
				return err
			}
			guidToKeys := make(map[string]string)
			err = json.NewDecoder(strings.NewReader(string(sigrunKeysJSON))).Decode(&guidToKeys)
			if err != nil {
				return err
			}
			for path, conf := range pathToConfig {
				guidToKeys[pathToGUID[path]] = conf.PublicKey
			}
			guidToKeysRaw, err := json.Marshal(guidToKeys)
			if err != nil {
				return err
			}
			cpol.Annotations["sigrun-keys"] = string(guidToKeysRaw)

			// add image verification rule for each image from config for each repo
			verifyImages := cpol.Spec.Rules[0].VerifyImages
			for _, conf := range pathToConfig {
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
