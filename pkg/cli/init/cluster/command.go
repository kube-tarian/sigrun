package cluster

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "cluster",
		Short: "Initializes cluster to use sigrun by installing a policy agent to verify images",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			fmt.Println("Installing kyverno")

			return exec.Command("kubectl", "create", "-f", "https://raw.githubusercontent.com/kyverno/kyverno/main/definitions/release/install.yaml").Run()
		},
	}
}
