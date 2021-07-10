module github.com/devopstoday11/sigrun

go 1.16

require (
	github.com/kyverno/kyverno v1.4.2-0.20210710010146-13caaed8b778
	github.com/pkg/errors v0.9.1
	github.com/sigstore/cosign v0.5.0
	github.com/spf13/cobra v1.2.1
	k8s.io/apimachinery v0.21.1
	k8s.io/cli-runtime v0.21.1
)

replace github.com/gorilla/rpc v1.2.0+incompatible => github.com/gorilla/rpc v1.2.0
