package policy

import (
	kyvernoV1 "github.com/kyverno/kyverno/pkg/api/kyverno/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const NAME = "sigrun-verify"

func New() *kyvernoV1.ClusterPolicy {
	background := false
	return &kyvernoV1.ClusterPolicy{
		TypeMeta: v1.TypeMeta{
			Kind:       "ClusterPolicy",
			APIVersion: "kyverno.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: NAME,
			Annotations: map[string]string{
				"sigrun-keys":  "",
				"sigrun-repos": "",
			},
		},
		Spec: kyvernoV1.Spec{
			Rules: []kyvernoV1.Rule{
				{
					Name: "sigrun",
					MatchResources: kyvernoV1.MatchResources{
						ResourceDescription: kyvernoV1.ResourceDescription{
							Kinds: []string{"Pod"},
						},
					},
				},
			},
			ValidationFailureAction: "enforce",
			Background:              &background,
		},
		Status: kyvernoV1.PolicyStatus{},
	}
}
