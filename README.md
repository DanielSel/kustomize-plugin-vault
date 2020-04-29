# kustomize-plugin-vault
Kustomize (v3) Secret Generator Plugin for HashiCorp Vault

# How to use
The plugin generates Kubernetes Secrets from KV Secrets in HashiCorp Vault. 
The target Vault server needs to be specified in the **VAULT_ADDR** environment variable

Authentication to Vault can be done using either of following environment variables:
* **VAULT_TOKEN**: Directly specify the token used to access vault
* *(Coming Soon)* **VAULT_ROLE_ID** and **VAULT_SECRET_ID**: Authenticate using Vault AppRole authentication
* *(Coming Soon)* **VAULT_AD_USER** and **VAULT_AD_PASSWORD**: Hard-coded credentials from a Service Account in LDAP/AD

Example for secret generator resource:
```
apiVersion: kustomize.rohde-schwarz.com/v1alpha1
kind: SecretsFromVault
metadata:
  name: secret-one
secrets:
  - path: path/to/secret/one
    key: key_one_in_vault_secret_one
    secretKey: KEY_ONE_IN_GENERATED_SECRET_ONE
  - path: path/to/secret/two
    key: key_seven_in_vault_secret_two
    secretKey: KEY_TWO_IN_GENERATED_SECRET_ONE
```
*Note: Every SecretsFromVault resource generates exactly one Kubernetes secret from **n** Vault secrets and **m** secrets per key*


How to reference it in your kustomization.yaml:
```
generators:
- vault-cloud-azure.yaml
```