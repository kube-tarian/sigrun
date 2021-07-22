package add

import (
	"context"
	"fmt"

	"github.com/devopstoday11/sigrun/pkg/config"

	"github.com/devopstoday11/sigrun/pkg/policy"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"

	kyvernoclient "github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

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

			for path, conf := range pathToConfig {
				cpol, err = policy.AddRepo(cpol, pathToGUID[path], path, conf)
				if err != nil {
					return err
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

func validateAddInput(args ...string) error {
	for i, arg := range args {
		if arg == "" {
			return fmt.Errorf("empty path found at position " + fmt.Sprint(i+1))
		}
	}

	return nil
}
