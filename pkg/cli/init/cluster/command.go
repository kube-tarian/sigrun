package cluster

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "cluster",
		Short: "Initializes a kubernetes cluster to be a sigrun consumer",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			fmt.Println("Installing kyverno...")

			return exec.Command("kubectl", "create", "-f", "https://raw.githubusercontent.com/kyverno/kyverno/main/definitions/install.yaml").Run()
		},
	}
}
