module github.com/danielsel/kustomize-plugin-vault

go 1.14

require (
	github.com/hashicorp/vault/api v1.0.4
	sigs.k8s.io/kustomize/api v0.5.1
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/hashicorp/go-cleanhttp => github.com/hashicorp/go-cleanhttp v0.5.0
replace github.com/pkg/errors => github.com/pkg/errors v0.9.1