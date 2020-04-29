package vault

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

const (
	Timeout             time.Duration = 1 * time.Minute
	EnvAddress          string        = "VAULT_ADDR"
	EnvAuthToken        string        = "VAULT_TOKEN"
	EnvAuthTokenPath    string        = "VAULT_TOKEN_PATH"
	EnvAuthLdapUser     string        = "VAULT_LDAP_USER"
	EnvAuthLdapPassword string        = "VAULT_LDAP_PASSWORD"
)

type ClientOptions struct {
	Address string
}

func NewClientFromEnv(options *ClientOptions) (*vaultapi.Client, error) {
	// Determine Vault Address
	config := &vaultapi.Config{}
	addr, ok := os.LookupEnv("VAULT_ADDR")
	switch {
	case options != nil && options.Address != "":
		config.Address = options.Address
	case ok:
		config.Address = addr
	default:
		return nil, fmt.Errorf("vaultAddress not set and missing `%s` env var: required", EnvAddress)

	}

	client, err := vaultapi.NewClient(config)
	if err != nil {
		return nil, err
	}

	// Attempt logins in order of priority (token -> ldap)
	if token, err := loginTokenEnv(); err == nil {
		return client, loginToken(client, token)
	}
	if tokenPath, err := loginTokenPathEnv(); err == nil {
		return client, loginTokenPath(client, tokenPath)
	}
	if user, password, err := loginLdapEnv(); err == nil {
		return client, loginLdap(client, user, password)
	}

	// Error if all logins fail
	return nil, fmt.Errorf(`missing credentials. one of the following environment variable combinations must be set:
	 (%s) or (%s) or (%s, %s)`,
		EnvAuthToken, EnvAuthTokenPath, EnvAuthLdapUser, EnvAuthLdapPassword)
}

func loginToken(client *vaultapi.Client, token string) error {
	client.SetToken(token)
	return nil
}

func loginTokenEnv() (string, error) {
	token, ok := os.LookupEnv(EnvAuthToken)
	if !ok {
		return "", fmt.Errorf("missing `%s` env var: required", EnvAuthToken)
	}
	return token, nil
}

func loginTokenPath(client *vaultapi.Client, tokenPath string) error {
	tBytes, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return fmt.Errorf("Could not read Vault token from ${%s}: %w", EnvAuthTokenPath, err)
	}

	token := strings.TrimSpace(string(tBytes))
	if len(token) == 0 {
		return fmt.Errorf("Vault token file is empty: %w", err)
	}

	client.SetToken(token)
	return nil
}

func loginTokenPathEnv() (string, error) {
	token, ok := os.LookupEnv(EnvAuthTokenPath)
	if !ok {
		return "", fmt.Errorf("missing `%s` env var: required", EnvAuthTokenPath)
	}
	return token, nil
}

func loginLdap(client *vaultapi.Client, user, password string) error {
	ctx, cnFnc := context.WithTimeout(context.Background(), Timeout)
	defer cnFnc()

	// Build and send authentication request
	authReq := client.NewRequest("POST", fmt.Sprintf("/v1/auth/ldap/login/%s", user))
	err := authReq.SetJSONBody(map[string]interface{}{"password": password})
	if err != nil {
		return fmt.Errorf("invalid auth request: %w", err)
	}
	authResp, err := client.RawRequestWithContext(ctx, authReq)
	if err != nil {
		return fmt.Errorf("ldap authentication error: %w", err)
	}

	// Extract token
	decodedAuthResp := &struct {
		Auth *struct {
			ClientToken string `json:"client_token,omitempty"`
		} `json:"auth,omitempty"`
	}{}
	if err := authResp.DecodeJSON(decodedAuthResp); err != nil {
		return fmt.Errorf("error parsing ldap authentication response: %w", err)
	}
	if decodedAuthResp.Auth == nil {
		return fmt.Errorf("invalid ldap authentication response: auth is nil")
	}
	if decodedAuthResp.Auth.ClientToken == "" {
		return fmt.Errorf("invalid ldap authentication response: auth.client_token is empty")
	}

	client.SetToken(decodedAuthResp.Auth.ClientToken)
	return nil
}

func loginLdapEnv() (string, string, error) {
	user, ok := os.LookupEnv(EnvAuthLdapUser)
	if !ok {
		return "", "", fmt.Errorf("missing `%s` env var: required", EnvAuthLdapUser)
	}
	password, ok := os.LookupEnv(EnvAuthLdapPassword)
	if !ok {
		return "", "", fmt.Errorf("missing `%s` env var: required", EnvAuthLdapPassword)
	}
	return user, password, nil
}
