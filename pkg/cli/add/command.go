package add

import (
	"fmt"

	"github.com/devopstoday11/sigrun/pkg/controller"

	"github.com/spf13/cobra"
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

			cont, err := controller.GetController()
			if err != nil {
				return err
			}

			return cont.Add(args...)
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
