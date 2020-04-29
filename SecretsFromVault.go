package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type vaultSecret struct {
	Path      string `json:"path,omitempty" yaml:"path,omitempty"`
	Key       string `json:"key,omitempty" yaml:"key,omitempty"`
	SecretKey string `json:"secretKey,omitempty" yaml:"secretKey,omitempty"`
}

type plugin struct {
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Options          *types.GeneratorOptions `json:"options,omitempty" yaml:"options,omitempty"`
	Secrets          []vaultSecret           `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	h                *resmap.PluginHelpers
	VaultClient      *vaultapi.Client
}

//nolint
var KustomizePlugin plugin

func (p *plugin) Config(h *resmap.PluginHelpers, c []byte) error {
	p.h = h
	vaultClient, err := createVaultClient()
	if err != nil {
		return err
	}
	p.VaultClient = vaultClient
	return yaml.Unmarshal(c, p)
}

// The plan here is to convert the plugin's input
// into the format used by the builtin secret generator plugin.
func (p *plugin) Generate() (resmap.ResMap, error) {
	args := types.SecretArgs{}
	args.Name = p.Name
	args.Namespace = p.Namespace

	for _, secret := range p.Secrets {
		value, err := p.getSecretFromVault(secret.Path, secret.Key)
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

func (p *plugin) getSecretFromVault(path string, key string) (value string, err error) {
	secret, err := p.VaultClient.Logical().Read(path)
	if err != nil {
		return "", err
	}
	if secret == nil {
		return "", fmt.Errorf("the path %s was not found", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("malformed secret data: %q", secret.Data["data"])
	}
	if v, ok := data[key].(string); ok {
		return v, nil
	}

	return "", fmt.Errorf("Failed to get secret from Vault: %s:%s", path, key)
}

func createVaultClient() (*vaultapi.Client, error) {
	addr, ok := os.LookupEnv("VAULT_ADDR")
	if !ok {
		return nil, errors.New("missing `VAULT_ADDR` env var: required")
	}
	token, exists := os.LookupEnv("VAULT_TOKEN")
	if !exists {
		tokenPath, exists := os.LookupEnv("VAULT_TOKEN_PATH")
		if !exists {
			return nil, errors.New("No vault token and no vault token path")
		}

		tBytes, err := ioutil.ReadFile(tokenPath)
		if err != nil {
			fmt.Println("Could not read Vault token from $VAULT_TOKEN_PATH")
			return nil, err
		}

		token = strings.TrimSpace(string(tBytes))
		if len(token) == 0 {
			fmt.Println("Vault token file is empty")
			return nil, err
		}
	}
	config := &vaultapi.Config{
		Address: addr,
	}
	client, err := vaultapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)
	return client, nil
}
