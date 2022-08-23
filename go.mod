module github.com/devopstoday11/sigrun

go 1.16

require (
	github.com/docker/distribution v2.8.1+incompatible
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-git/go-git/v5 v5.2.0
	github.com/google/go-containerregistry v0.11.0
	github.com/pkg/errors v0.9.1
	github.com/sigstore/cosign v1.10.1
	github.com/sigstore/rekor v0.4.1-0.20220114213500-23f583409af3
	github.com/sigstore/sigstore v1.2.1-0.20220614141825-9c0e2e247545
	github.com/spf13/cobra v1.5.0
	github.com/tidwall/pretty v1.2.0
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/cli-runtime v0.22.0
	k8s.io/client-go v0.23.5
)

replace github.com/gorilla/rpc v1.2.0+incompatible => github.com/gorilla/rpc v1.2.0
