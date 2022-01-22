package namespace

import (
	"context"
	"fmt"

	"github.com/devopstoday11/sigrun/pkg/controller"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace",
		Short: "Scans every pod in the namespace and checks if they are valid",
		RunE: func(cmd *cobra.Command, namespaces []string) error {

			kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
			if err != nil {
				return err
			}

			kclient, err := kubernetes.NewForConfig(kRestConf)
			if err != nil {
				return err
			}

			configMap, err := kclient.CoreV1().ConfigMaps(controller.SIGRUN_CONTROLLER_NAMESPACE).Get(context.Background(), controller.SIGRUN_CONTROLLER_CONFIG, v1.GetOptions{})
			if err != nil {
				return err
			}

			var containers []corev1.Container
			for _, namespace := range namespaces {
				podList, err := kclient.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{})
				if err != nil {
					return err
				}

				for _, pod := range podList.Items {
					containers = append(containers, pod.Spec.Containers...)
					fmt.Println("scanning pod " + pod.Name + "....")
				}
			}

			return controller.ValidateContainers(configMap, containers)
		},
	}

	return cmd
}
