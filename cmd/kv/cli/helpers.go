package cli

import (
	"fmt"
	"strings"

	"github.com/yousysadmin/kv/internal/enckeystore"
	"github.com/yousysadmin/kv/pkg/encrypt"
)

func parseKey(input string) (key string, bucket string) {
	if strings.Contains(input, "@") {
		parts := strings.SplitN(input, "@", 2)
		key = parts[0]
		bucket = parts[1]
		return
	}
	return input, bucketName
}

// loadAllKeys returns all keys in Encryption Key Store
func loadAllKeys(storePath, encryptionKey string) (map[string]string, *enckeystore.EncryptionKeyStore, error) {
	// if encryptionKey is set that validate and return as default
	if encryptionKey != "" {
		if err := encrypt.ValidateAESKey(encryptionKey); err != nil {
			return nil, nil, fmt.Errorf("provided encryption key is invalid: %w", err)
		}
		return map[string]string{"default": encryptionKey}, nil, nil
	}

	// Encryption Key Store init
	ks := enckeystore.NewEncryptionKeyStore(storePath)
	if err := ks.Load(); err != nil {
		return nil, nil, fmt.Errorf("load key store: %w", err)
	}

	// Ensure default, if not then will generate new one and save
	created := !ks.HasKey("default")
	def, err := ks.EnsureDefaultKey()
	if err != nil {
		return nil, nil, fmt.Errorf("ensure default key: %w", err)
	}
	if created {
		if err := ks.Save(); err != nil {
			return nil, nil, fmt.Errorf("save key store: %w", err)
		}
	}

	keys := make(map[string]string, len(ks.Keys))
	for b, k := range ks.Keys {
		keys[b] = string(k)
	}
	// Ensure default is present
	keys["default"] = string(def)

	return keys, ks, nil
}

// selectKey chooses a key for a bucket from the Encryption Keys Store.
func selectKey(keysStore map[string]string, bucket string) (string, error) {
	if k, ok := keysStore[bucket]; ok && k != "" {
		return k, nil
	}
	if def, ok := keysStore["default"]; ok && def != "" {
		return def, nil
	}
	return "", fmt.Errorf("no key for bucket %q and no default key", bucket)
}
