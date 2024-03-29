# SigRun <img src="/sigrun1.png" width="300" height="300"> 
Sign your artifacts source code or container images using Sigstore chain of tools & Known Container Image Build tools, Save the Signatures you want to use within your Infra, and Validate &amp; Control the deployments to allow only the known Signatures. Shift-left your supply chain security!
> What's with the Name (in case if you are curious)?
> You can think of multiple ways. It has a flexible interpretation, like Signatures for Runtime or Runtime Signatures or Sign Software for Runtime use. Whatever you want to imagine! :smiley: 
#

### Install

##### Dependencies
Before installing the application the following dependencies need to be installed:
1. Kubernetes command line application `kubectl`
2. Golang version greater than 1.16

```
go install cmd/sigrun/sigrun.go
```

### Usage
```
sigrun --help
```
Please refer to [this](./docs/USAGE.md) for information about basic flow.

##### Purpose:
To make it easy to use SigStore chain of tools. Make the Supply Chain Security for Software adoption easy. 
#
##### Usage feasibility:
Local, CI/CD pipelines, K8s Clusters, VMs. 
#
#### Features:
- Using Sigstore tools in your Infra for Air-Gap offline usage via your CI/CD Pipeline
- Sign your artifacts, container images, files, packages, etc. automatically along with their sha256 digest creation & saving into ledger
- Private & Public key-pair generator (Cosign, GPG, and more in future) for signing 
- Keyless signing 
- Save your artifacts signatures to certain ledger storage
- Save your container image signatures to certain ledger storage
- Validate Signatures using Storage location of Signatures
- Control deployments to allow only known Signatures using our Custom Admission Controller or OPA/Kyverno/Gatekeeper
- Vault Integration to save Keys if you prefer to save private key(s) 
- CI/CD Tools integration
- Integration with tools like Buildpacks, Buildah, Source2Image, Kaniko, Skaffold, Docker Build, Podman, etc. 
- OIDC/Dex embeded for Login 
- Vulnerability Scanning of your container images
- Integrate with Non-Profit SigStore public services/tools
- Integrate with Syft for Software Bill of Materials (SBOM) [github.com/anchore/syft]
- Integrate with Package Hunter by Gitlab [gitlab.com/gitlab-org/security-products/package-hunter]
- 


#

## Contributing
See [docs/contributing.md](docs/CONTRIBUTING.md)

## Code of Conduct
See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## CodeOwners & Maintainers list
See [MAINTAINERS.md](MAINTAINERS.md)

