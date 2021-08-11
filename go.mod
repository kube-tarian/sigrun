module github.com/devopstoday11/sigrun

go 1.16

require (
	github.com/docker/distribution v2.7.1+incompatible
	github.com/google/go-containerregistry v0.5.1
	github.com/kyverno/kyverno v1.4.2-0.20210710010146-13caaed8b778
	github.com/pkg/errors v0.9.1
	github.com/sassoftware/relic v7.2.1+incompatible
	github.com/sigstore/cosign v1.0.1
	github.com/sigstore/sigstore v0.0.0-20210726180807-7e34e36ecda1
	github.com/spf13/cobra v1.2.1
	github.com/tidwall/pretty v1.2.0
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/cli-runtime v0.21.1
	k8s.io/client-go v0.21.3
)

replace github.com/gorilla/rpc v1.2.0+incompatible => github.com/gorilla/rpc v1.2.0
