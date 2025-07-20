package cmd

import (
	"errors"
	"fmt"
	"github.com/yousysadmin/kv/internal/storage"
	"github.com/yousysadmin/kv/pkg/encrypt"
	"os"
	"strings"
)

func parseKey(input string) (key string, bucket string) {
	if strings.Contains(input, "@") {
		parts := strings.SplitN(input, "@", 2)
		return parts[0], parts[1]
	}
	return input, storage.DefaultBucket
}

func getEncryptKey(encryptKey string, encryptKeyFile string) (string, error) {
	if encryptKey != "" {
		if err := encrypt.ValidateAESKey(encryptKey); err != nil {
			return "", fmt.Errorf("provided encryption key is invalid: %w", err)
		}
		return encryptKey, nil
	}

	data, err := os.ReadFile(encryptKeyFile)
	if err == nil && len(strings.TrimSpace(string(data))) > 0 {
		key := strings.TrimSpace(string(data))
		if err := encrypt.ValidateAESKey(key); err == nil {
			return key, nil
		}
	}

	if !errors.Is(err, os.ErrNotExist) && err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	key, err := encrypt.GenerateRandomAESKey(encrypt.AES256)
	if err != nil {
		return "", fmt.Errorf("failed to generate AES key: %w", err)
	}

	if err := os.WriteFile(encryptKeyFile, []byte(key), 0600); err != nil {
		return "", fmt.Errorf("failed to write key file: %w", err)
	}

	return key, nil
}
