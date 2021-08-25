package sign

import (
	"fmt"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/config"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Signs all images related to a sigrun repo and pushes the signatures to the corresponding container registry",
	}

	var annotationsRaw string
	cmd.Flags().StringVar(&annotationsRaw, "annotations", "", "specify annotations if any, example format - name=jon,org=sigrun")
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		conf, err := config.ReadRepositoryConfig()
		if err != nil {
			return err
		}

		annotations, err := parseAnnotations(annotationsRaw)
		if err != nil {
			return err
		}

		return conf.SignImages(annotations)
	}

	return cmd
}

func parseAnnotations(annotationsRaw string) (map[string]string, error) {
	if annotationsRaw == "" {
		return make(map[string]string), nil
	}

	annotationsKV := strings.Split(annotationsRaw, ",")
	annotations := make(map[string]string)
	for _, annotation := range annotationsKV {
		annotationSlc := strings.Split(annotation, "=")
		if len(annotationSlc) != 2 {
			return nil, fmt.Errorf("could not parse annotation - " + annotation)
		}

		annotations[annotationSlc[0]] = annotationSlc[1]
	}

	return annotations, nil
}
