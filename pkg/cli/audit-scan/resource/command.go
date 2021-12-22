package resource

import (
	"context"

	"github.com/devopstoday11/sigrun/pkg/controller"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource",
		Short: "Scans a pod in the given namespace",
	}

	var namespace string
	cmd.Flags().StringVar(&namespace, "namespace", "default", "The namespace in which the pod exists")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

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
		for _, podName := range args {
			pod, err := kclient.CoreV1().Pods(namespace).Get(context.Background(), podName, v1.GetOptions{})
			if err != nil {
				return err
			}
			containers = append(containers, pod.Spec.Containers...)
		}

		return controller.ValidateContainers(configMap, containers)
	}

	return cmd
}
