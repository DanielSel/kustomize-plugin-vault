package main

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"

	"github.com/danielsel/kustomize-plugin-vault/pkg/vault"
)

type vaultSecret struct {
	Path      string `json:"path,omitempty" yaml:"path,omitempty"`
	Key       string `json:"key,omitempty" yaml:"key,omitempty"`
	SecretKey string `json:"secretKey,omitempty" yaml:"secretKey,omitempty"`
}

type plugin struct {
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Options      *types.GeneratorOptions `json:"options,omitempty" yaml:"options,omitempty"`
	Type         string                  `json:"type,omitempty" yaml:"type,omitempty"`
	Secrets      []vaultSecret           `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	VaultAddress string                  `json:"vaultAddress,omitempty" yaml:"vaultAddress,omitempty"`

	h           *resmap.PluginHelpers
	VaultClient *vaultapi.Client
}

//nolint
var KustomizePlugin plugin

func (p *plugin) Config(h *resmap.PluginHelpers, c []byte) error {
	p.h = h
	if err := yaml.Unmarshal(c, p); err != nil {
		return fmt.Errorf("error parsing secret generator yaml: %w", err)
	}
	vaultClient, err := vault.NewClientFromEnv(&vault.ClientOptions{Address: p.VaultAddress})
	if err != nil {
		return fmt.Errorf("vault secret '%s': error creating vault client: %w", p.Name, err)
	}
	p.VaultClient = vaultClient
	return nil
}

// The plan here is to convert the plugin's input
// into the format used by the builtin secret generator plugin.
func (p *plugin) Generate() (resmap.ResMap, error) {
	args := types.SecretArgs{}
	args.Name = p.Name
	args.Namespace = p.Namespace
	args.Type = p.Type

	for _, secret := range p.Secrets {
		value, err := vault.RetrieveSecret(p.VaultClient, secret.Path, secret.Key)
		if err != nil {
			return nil, err
		}

		var key string
		if secret.SecretKey != "" {
			key = secret.SecretKey
		} else {
			key = secret.Key
		}

		entry := fmt.Sprintf("%s=%s", key, value)
		args.LiteralSources = append(args.LiteralSources, entry)
	}
	return p.h.ResmapFactory().FromSecretArgs(
		kv.NewLoader(p.h.Loader(), p.h.Validator()), p.Options, args)
}
