package enckeystore

import (
	"bufio"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/yousysadmin/kv/pkg/encrypt"
	"gopkg.in/yaml.v3"
)

type EncryptionKey string

// Validate ensures the AES key.
func (k EncryptionKey) Validate() error {
	if err := encrypt.ValidateAESKey(string(k)); err != nil {
		return fmt.Errorf("invalid AES key: %w", err)
	}
	return nil
}

// GenerateEncryptionKey creates a new AES-256 key using your encrypt package.
func GenerateEncryptionKey() (EncryptionKey, error) {
	key, err := encrypt.GenerateRandomAESKey(encrypt.AES256)
	if err != nil {
		return "", fmt.Errorf("failed to generate AES-256 key: %w", err)
	}
	return EncryptionKey(key), nil
}

type EncryptionKeyStore struct {
	path string                   `yaml:"-"`
	Keys map[string]EncryptionKey `yaml:"keys"`

	mu sync.RWMutex
}

// NewEncryptionKeyStore creates an empty store bound to a file path.
func NewEncryptionKeyStore(path string) *EncryptionKeyStore {
	return &EncryptionKeyStore{
		path: path,
		Keys: make(map[string]EncryptionKey),
	}
}

// Load reads keys from disk. Missing file just return is empty store.
func (s *EncryptionKeyStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if s.Keys == nil {
				s.Keys = make(map[string]EncryptionKey)
			}
			return nil
		}
		return err
	}

	type onDisk struct {
		Keys map[string]EncryptionKey `yaml:"keys"`
	}
	var disk onDisk

	dec := yaml.NewDecoder(strings.NewReader(string(b)))
	dec.KnownFields(true)
	if err := dec.Decode(&disk); err != nil {
		return fmt.Errorf("parse %s: %w", s.path, err)
	}

	if disk.Keys == nil {
		disk.Keys = make(map[string]EncryptionKey)
	}
	s.Keys = disk.Keys
	return nil
}

// ReLoad is an alias for Load.
func (s *EncryptionKeyStore) ReLoad() error { return s.Load() }

// Save writes to disk atomically with 0600 perms.
func (s *EncryptionKeyStore) Save() error {
	s.mu.RLock()
	dump := struct {
		Keys map[string]EncryptionKey `yaml:"keys"`
	}{
		Keys: s.copyKeysLocked(),
	}
	s.mu.RUnlock()

	data, err := yaml.Marshal(&dump)
	if err != nil {
		return fmt.Errorf("yaml marshal: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	return atomicWriteFile(s.path, data, 0o600)
}

// atomicWriteFile writes to a temp file, fsyncs file & dir, then renames.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-keys-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp: %w", err)
	}

	w := bufio.NewWriter(tmp)
	if _, err := w.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := w.Flush(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("flush temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("fsync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp -> %s: %w", path, err)
	}

	// fsync the directory to persist the rename on some filesystems
	dirFD, err := os.Open(dir)
	if err == nil {
		_ = dirFD.Sync()
		_ = dirFD.Close()
	}
	return nil
}

// Get returns the key for bucketName if present and valid,
// else returns a valid "default" key if present. Otherwise error.
func (s *EncryptionKeyStore) Get(bucketName string) (EncryptionKey, error) {
	s.mu.RLock()
	key, ok := s.Keys[bucketName]
	def, hasDef := s.Keys["default"]
	s.mu.RUnlock()

	if ok && key != "" && key.Validate() == nil {
		return key, nil
	}
	if hasDef && def != "" && def.Validate() == nil {
		return def, nil
	}
	return "", fmt.Errorf("no valid encryption key for %q and no valid default", bucketName)
}

// AddKey inserts a new key for a bucket after validation.
// If the bucket already has a key, it returns an error (no replacement allowed).
func (s *EncryptionKeyStore) AddKey(bucketName string, key EncryptionKey) error {
	if strings.TrimSpace(bucketName) == "" {
		return fmt.Errorf("bucket name cannot be empty")
	}
	if err := key.Validate(); err != nil {
		return fmt.Errorf("invalid encryption key for bucket %q: %w", bucketName, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Keys == nil {
		s.Keys = make(map[string]EncryptionKey)
	}
	if _, exists := s.Keys[bucketName]; exists {
		return fmt.Errorf("encryption key for bucket %q already exists; remove it from file to replace", bucketName)
	}
	s.Keys[bucketName] = key
	return nil
}

// AddDefaultKey sets the "default" key if and only if it does not yet exist.
// If default already exists, returns an error (no replacement allowed).
// If key == "", a new AES-256 key is generated.
func (s *EncryptionKeyStore) AddDefaultKey(key EncryptionKey) (EncryptionKey, error) {
	if key == "" {
		gen, err := GenerateEncryptionKey()
		if err != nil {
			return "", err
		}
		key = gen
	}
	if err := key.Validate(); err != nil {
		return "", fmt.Errorf("invalid default key: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Keys == nil {
		s.Keys = make(map[string]EncryptionKey)
	}
	if _, exists := s.Keys["default"]; exists {
		return "", fmt.Errorf(`"default" key already exists; remove it from file to replace`)
	}
	s.Keys["default"] = key
	return key, nil
}

// EnsureDefaultKey creates and stores a default key only if missing.
// Never replaces existing default.
func (s *EncryptionKeyStore) EnsureDefaultKey() (EncryptionKey, error) {
	s.mu.RLock()
	def, exists := s.Keys["default"]
	s.mu.RUnlock()
	if exists {
		if err := def.Validate(); err != nil {
			// existing but invalid → user must fix file manually
			return "", fmt.Errorf(`existing "default" key is invalid; remove/fix it in the file: %w`, err)
		}
		return def, nil
	}
	return s.AddDefaultKey("")
}

// HasKey reports true if the store currently holds any key for the bucket (no validation).
func (s *EncryptionKeyStore) HasKey(bucketName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.Keys[bucketName]
	return ok
}

// ListBuckets returns bucket names in sorted order.
func (s *EncryptionKeyStore) ListBuckets() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.Keys))
	for k := range s.Keys {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// copyKeysLocked clones s.Keys under read lock.
func (s *EncryptionKeyStore) copyKeysLocked() map[string]EncryptionKey {
	cp := make(map[string]EncryptionKey, len(s.Keys))
	maps.Copy(cp, s.Keys)
	return cp
}
