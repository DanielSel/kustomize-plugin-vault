module github.com/danielsel/kustomize-plugin-vault

go 1.14

require (
	github.com/hashicorp/vault/api v1.0.4
	sigs.k8s.io/kustomize/api v0.3.2
	sigs.k8s.io/yaml v1.1.0
)

exclude (
	github.com/russross/blackfriday v2.0.0+incompatible
	sigs.k8s.io/kustomize/api v0.2.0
)
