# SigRun
Sign your artifacts source code or container images using Sigstore chain of tools, Save the Signatures you want to use within your Infra, and Validate &amp; Control the deployments to allow only the known Signatures.
#
##### Purpose:
To make it easy to use SigStore chain of tools. Make the Supply Chain Security for Software adoption easy. 
#
##### Usage feasibility:
Local, CI/CD pipelines, K8s Clusters, VMs. 
#
#### Features:
- Sign your artifacts
- Private & Public key-pair
- Keyless
- Save your artifacts signatures to certain storage
- Save your container image signatures to certain storage
- Validate Signatures using Storage location of Signatures
- Control deployments to allow only known Signatures using our Custom Admission Controller or OPA/Kyverno/Gatekeeper
- Vault Integration to save Keys
- CI/CD Tools integration
- OIDC
- Vulnerability Scanning of your container images
- RBAC -> Enterprise
- SSO -> Enterprise

#
