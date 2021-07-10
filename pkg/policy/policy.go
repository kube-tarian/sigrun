package policy

import kyvernoV1 "github.com/kyverno/kyverno/pkg/api/kyverno/v1"

func New() *kyvernoV1.ClusterPolicy {
	return &kyvernoV1.ClusterPolicy{}
}
