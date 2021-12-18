package config_map_yaml

import (
	"fmt"

	"github.com/devopstoday11/sigrun/pkg/config"
	"github.com/devopstoday11/sigrun/pkg/controller"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config-map-yaml",
		Short: "generates a config map from given sigrun config files",
		RunE: func(cmd *cobra.Command, args []string) error {
			pathToConfig, err := config.ReadReposFromPath(args...)
			if err != nil {
				return err
			}

			cont := controller.NewSigrunController()
			configMap := &corev1.ConfigMap{
				TypeMeta:   metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{Name: controller.SIGRUN_CONTROLLER_CONFIG},
				Data: map[string]string{
					"guid_to_repo_info": "",
					"image_to_guids":    "",
				},
			}

			for path, conf := range pathToConfig {
				guid, err := config.GetGUIDFromConfigFile(path)
				if err != nil {
					return err
				}

				configMap, err = cont.AddRepo(configMap, guid, path, conf)
				if err != nil {
					return err
				}
			}

			configMapRaw, err := yaml.Marshal(configMap)
			if err != nil {
				return err
			}

			fmt.Println(string(configMapRaw))

			return nil
		},
	}

	return cmd
}
