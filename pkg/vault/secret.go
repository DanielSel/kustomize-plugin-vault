package vault

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

func RetrieveSecret(client *vaultapi.Client, path string, key string) (value string, err error) {
	secret, err := client.Logical().Read(path)
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

	return "", fmt.Errorf("failed to get secret from Vault: %s:%s", path, key)
}
