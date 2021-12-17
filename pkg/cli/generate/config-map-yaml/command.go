package config_map_yaml

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config-map-yaml",
		Short: "generates a config map from given sigrun config files",
		RunE: func(cmd *cobra.Command, args []string) error {

			
			
			return nil
		},
	}

	return cmd
}
