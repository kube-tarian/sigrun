package audit_scan

import (
	full_cluster "github.com/devopstoday11/sigrun/pkg/cli/audit-scan/full-cluster"
	"github.com/devopstoday11/sigrun/pkg/cli/audit-scan/namespace"
	"github.com/devopstoday11/sigrun/pkg/cli/audit-scan/resource"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit-scan",
		Short: "check if existing resources in the cluster are valid",
	}

	cmd.AddCommand(full_cluster.Command(), namespace.Command(), resource.Command())

	return cmd
}
