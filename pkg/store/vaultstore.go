package store

import (
	"encoding/base64"
	"fmt"
	"path"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

const maxRecurseDepth = 10

type VaultStore struct {
	vaultClient *vaultapi.Client
	vaultPath   string
}

const dataField = "data"

var errKeyNotExists = fmt.Errorf("vault key not exists")
var errInvalidValue = fmt.Errorf("vault value is malformed")

func createVaultStoreOrDie(vaultPath string, roleId string, secretId string) Store {
	logCtx := log.Fields{
		"vaultPath": vaultPath,
		"roleId":    roleId,
		"secretId":  "<redacted>",
	}

	// create vault client
	vaultConfig := *vaultapi.DefaultConfig()
	vaultConfig.ReadEnvironment()

	log.WithFields(logCtx).Debug("Creating vault client")
	client, err := vaultapi.NewClient(&vaultConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create vault client: %v", err))
	}

	// log in with role id
	creds := make(map[string]interface{})
	creds["role_id"] = roleId
	creds["secret_id"] = secretId
	log.WithFields(logCtx).Info("Authenticating with role ID")
	secret, err := client.Logical().Write("auth/approle/login", creds)
	if err != nil {
		log.WithFields(logCtx).Errorf("Authentication with role ID failed: %v", err)
		panic(fmt.Errorf("Authentication failed with Vault: %v", err))
	}
	log.WithFields(logCtx).Debug("Authentication succeeded, client token set")
	client.SetToken(secret.Auth.ClientToken)

	return &VaultStore{
		vaultClient: client,
		vaultPath:   vaultPath,
	}
}

func (store *VaultStore) makeAbsolutePath(relPath string) string {
	return fmt.Sprintf("secret/%s/%s", store.vaultPath, relPath)
}

func (store *VaultStore) makeError(action string, key string, err error) error {
	return fmt.Errorf("failed to %s key '%s' from %s: %v", action, key, store, err)
}

func (store *VaultStore) List(key string) ([]string, error) {
	store.log(fmt.Sprintf("List key %s", key))

	keys, err := store.list(key)
	if err != nil {
		return nil, store.makeError("list", key, err)
	}

	// remove all trailing slashes
	sanitisedKeys := make([]string, len(keys))
	for i, k := range keys {
		sanitisedKeys[i] = strings.TrimSuffix(k, "/")
	}

	return sanitisedKeys, nil
}

func (store *VaultStore) Exists(key string) (bool, error) {
	store.log(fmt.Sprintf("Head key %s", key))

	absPath := store.makeAbsolutePath(key)

	secret, err := store.vaultClient.Logical().Read(absPath)
	if err != nil {
		return false, store.makeError("head", key, err)
	}
	return secret != nil, nil
}

func (store *VaultStore) Get(key string) ([]byte, error) {
	store.log(fmt.Sprintf("Get key %s", key))

	absPath := store.makeAbsolutePath(key)

	secret, err := store.vaultClient.Logical().Read(absPath)
	if err != nil {
		return nil, store.makeError("get", key, err)
	}
	if secret == nil {
		return nil, store.makeError("get", key, errKeyNotExists)
	}
	if secret.Data[dataField] == nil {
		return nil, store.makeError("get", key, errInvalidValue)
	}

	encodedData := secret.Data[dataField].(string)
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %v", err)
	}

	return decodedData, nil
}

func (store *VaultStore) Set(key string, data []byte) error {
	encodedData := base64.StdEncoding.EncodeToString(data)
	store.log(fmt.Sprintf("Set key %s (len: %d bytes)", key, len(encodedData)))

	absPath := store.makeAbsolutePath(key)

	secretData := make(map[string]interface{})
	secretData[dataField] = encodedData
	_, err := store.vaultClient.Logical().Write(absPath, secretData)
	if err != nil {
		return store.makeError("set", key, err)
	}

	return nil
}

func (store *VaultStore) Delete(key string) error {
	store.log(fmt.Sprintf("Delete key %s", key))

	absPath := store.makeAbsolutePath(key)

	_, err := store.vaultClient.Logical().Delete(absPath)
	if err != nil {
		return store.makeError("delete", key, err)
	}

	return nil
}

func (store *VaultStore) DeleteAll(key string) error {
	store.log(fmt.Sprintf("DeleteAll key %s", key))

	// list all keys to delete
	keysToDel, err := store.listRecurse(key, 0)
	if err != nil {
		return store.makeError("deleteAll", key, err)
	}
	log.Debugf("DeleteAll: deleting %d keys", len(keysToDel))

	// delete all listed keys
	for _, k := range keysToDel {
		if err := store.Delete(k); err != nil {
			// warning only if one key failed to delete
			log.Warnf("failed to delete %d: %v", k, err)
		}
	}

	return nil
}

func (store *VaultStore) String() string {
	return fmt.Sprintf("vault://%s", store.vaultPath)
}

func (store *VaultStore) listRecurse(key string, depth int) ([]string, error) {
	// sanity check (should not exceed 10 levels)
	if depth > maxRecurseDepth {
		return nil, store.makeError("listRecurse", key, fmt.Errorf("maximum recurse depth reached"))
	}

	// list all keys under the current key
	keys, err := store.list(key)
	if err != nil {
		return nil, store.makeError("listRecurse", key, err)
	}

	// create a new string array to hold flattened keys
	flatKeys := make([]string, 0)

	// for each key
	for _, k := range keys {
		fullKey := path.Join(key, k)

		if !isDirectory(k) {
			// if not a directory, append to the flatKeys directly
			flatKeys = append(flatKeys, fullKey)
		} else {
			// otherwise, call listRecurse on it and append all returned keys to flatKeys
			subKeys, err := store.listRecurse(fullKey, depth+1)
			if err != nil {
				return nil, store.makeError("listRecurse", key, err)
			}
			flatKeys = append(flatKeys, subKeys...)
		}
	}

	// return flat list of keys
	return flatKeys, nil
}

func (store *VaultStore) list(key string) ([]string, error) {
	absPath := store.makeAbsolutePath(key)

	secret, err := store.vaultClient.Logical().List(absPath)
	if err != nil {
		return nil, store.makeError("list", key, err)
	}
	if secret == nil {
		return make([]string, 0), nil
	}
	if secret.Data["keys"] == nil {
		return make([]string, 0), nil
	}

	data := secret.Data["keys"]
	return dataAsList(data)
}

func (store *VaultStore) log(msg string) {
	log.WithField("vaultPath", store.vaultPath).Debugf("VAULT_STORE: %s", msg)
}

func isDirectory(key string) bool {
	return strings.HasSuffix(key, "/")
}

func dataAsList(data interface{}) ([]string, error) {
	if list, ok := data.([]interface{}); ok {
		keys := make([]string, 0)
		for _, k := range list {
			keys = append(keys, k.(string))
		}
		return keys, nil
	}

	return nil, fmt.Errorf("data is not a list")
}
