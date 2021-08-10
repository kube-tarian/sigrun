package list

import (
	"encoding/json"
	"fmt"

	"github.com/devopstoday11/sigrun/pkg/controller"

	"github.com/tidwall/pretty"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists metadata about sigrun repos that have been added",
	}

	var controllerF string
	cmd.Flags().StringVar(&controllerF, "controller", "sigrun", "specify the controller you would like to use")
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		cont, err := controller.GetController(controllerF)
		if err != nil {
			return err
		}

		repoInfo, err := cont.List()
		if err != nil {
			return err
		}

		encodedRepoInfo, err := json.Marshal(repoInfo)
		if err != nil {
			return err
		}

		fmt.Println("Sigrun-repos:\n" + string(pretty.Pretty(encodedRepoInfo)))
		return nil
	}

	return cmd
}
