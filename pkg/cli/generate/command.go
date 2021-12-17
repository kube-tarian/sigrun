package generate

import (
	config_map_yaml "github.com/devopstoday11/sigrun/pkg/cli/generate/config-map-yaml"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "commands that generate some artifacts",
	}

	cmd.AddCommand(config_map_yaml.Command())

	return cmd
}
