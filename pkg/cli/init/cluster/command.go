package cluster

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/policy"

	"k8s.io/client-go/rest"

	kyvernoclient "github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "cluster",
		Short: "Initializes a kubernetes cluster to be a sigrun consumer",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			fmt.Println("Attempting to install default policy agent(kyverno)...")
			kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
			if err != nil {
				return err
			}

			return initKyverno(kRestConf)
		},
	}
}

func initKyverno(kRestConf *rest.Config) error {
	kClient, err := kyvernoclient.NewForConfig(kRestConf)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cpol, err := kClient.KyvernoV1().ClusterPolicies().Get(ctx, policy.NAME, v1.GetOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
	} else {
		if cpol.Name == policy.NAME {
			fmt.Println("Cluster has already been initialized")
		}
	}

	fmt.Println("Installing default policy agent(kyverno)...")
	err = exec.Command("kubectl", "create", "-f", "https://raw.githubusercontent.com/kyverno/kyverno/main/definitions/install.yaml").Run()
	if err != nil {
		return err
	}

	_, err = kClient.KyvernoV1().ClusterPolicies().Create(ctx, policy.New(), v1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}
