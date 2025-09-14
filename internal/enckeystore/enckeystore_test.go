package enckeystore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// helper: create a fresh temp directory + file path for a store.
func newTempStorePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "keys.yaml")
}

func TestLoadMissingFile(t *testing.T) {
	path := newTempStorePath(t)
	s := NewEncryptionKeyStore(path)

	if err := s.Load(); err != nil {
		t.Fatalf("Load() on missing file = %v; want nil", err)
	}
	if s.Keys == nil || len(s.Keys) != 0 {
		t.Fatalf("Keys after Load() on missing file = %#v; want empty map", s.Keys)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := newTempStorePath(t)
	s := NewEncryptionKeyStore(path)

	// Ensure default
	def, err := s.EnsureDefaultKey()
	if err != nil {
		t.Fatalf("EnsureDefaultKey() = %v", err)
	}
	if err := def.Validate(); err != nil {
		t.Fatalf("default key Validate() = %v", err)
	}

	// Add a bucket key
	k, err := GenerateEncryptionKey()
	if err != nil {
		t.Fatalf("GenerateEncryptionKey() = %v", err)
	}
	if err := s.AddKey("photos", k); err != nil {
		t.Fatalf("AddKey(photos) = %v", err)
	}

	// Persist
	if err := s.Save(); err != nil {
		t.Fatalf("Save() = %v", err)
	}

	// Load into a fresh store and compare
	s2 := NewEncryptionKeyStore(path)
	if err := s2.Load(); err != nil {
		t.Fatalf("second Load() = %v", err)
	}

	if got := len(s2.Keys); got != len(s.Keys) {
		t.Fatalf("len(Keys) = %d; want %d", got, len(s.Keys))
	}
	if s2.Keys["default"] == "" || s2.Keys["photos"] == "" {
		t.Fatalf("missing keys after reload: %#v", s2.Keys)
	}
}

func TestGetBucketAndFallback(t *testing.T) {
	path := newTempStorePath(t)
	s := NewEncryptionKeyStore(path)

	// default
	def, err := s.AddDefaultKey("")
	if err != nil {
		t.Fatalf("AddDefaultKey(\"\") = %v", err)
	}
	// specific
	bk, err := GenerateEncryptionKey()
	if err != nil {
		t.Fatalf("GenerateEncryptionKey() = %v", err)
	}
	if err := s.AddKey("logs", bk); err != nil {
		t.Fatalf("AddKey(logs) = %v", err)
	}

	// bucket present
	got, err := s.Get("logs")
	if err != nil {
		t.Fatalf("Get(logs) error = %v", err)
	}
	if got != bk {
		t.Fatalf("Get(logs) = %q; want %q", got, bk)
	}

	// bucket missing -> fallback to default
	got2, err := s.Get("missing")
	if err != nil {
		t.Fatalf("Get(missing) error = %v", err)
	}
	if got2 != def {
		t.Fatalf("Get(missing) fallback = %q; want default %q", got2, def)
	}
}

func TestAddKeyNoReplace(t *testing.T) {
	path := newTempStorePath(t)
	s := NewEncryptionKeyStore(path)

	k1, _ := GenerateEncryptionKey()
	if err := s.AddKey("b", k1); err != nil {
		t.Fatalf("AddKey first = %v", err)
	}
	k2, _ := GenerateEncryptionKey()
	if err := s.AddKey("b", k2); err == nil {
		t.Fatalf("AddKey replace = nil; want error")
	}
}

func TestAddDefaultNoReplace(t *testing.T) {
	path := newTempStorePath(t)
	s := NewEncryptionKeyStore(path)

	if _, err := s.AddDefaultKey(""); err != nil {
		t.Fatalf("AddDefaultKey first = %v", err)
	}
	if _, err := s.AddDefaultKey(""); err == nil {
		t.Fatalf("AddDefaultKey replace = nil; want error")
	}
}

func TestEnsureDefaultInvalidExisting(t *testing.T) {
	path := newTempStorePath(t)

	// Write an invalid default key directly to disk (bypass validation)
	type onDisk struct {
		Keys map[string]EncryptionKey `yaml:"keys"`
	}
	bad := onDisk{
		Keys: map[string]EncryptionKey{
			"default": "too-short-or-wrong-format",
		},
	}
	raw, err := yaml.Marshal(&bad)
	if err != nil {
		t.Fatalf("yaml marshal = %v", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write = %v", err)
	}

	s := NewEncryptionKeyStore(path)
	if err := s.Load(); err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if _, err := s.EnsureDefaultKey(); err == nil {
		t.Fatalf("EnsureDefaultKey with invalid existing default = nil; want error")
	} else if !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("EnsureDefaultKey error = %v; want contains 'invalid'", err)
	}
}

func TestListBucketsSorted(t *testing.T) {
	path := newTempStorePath(t)
	s := NewEncryptionKeyStore(path)

	// Insert out of order
	ka, _ := GenerateEncryptionKey()
	kb, _ := GenerateEncryptionKey()
	if _, err := s.AddDefaultKey(""); err != nil {
		t.Fatalf("AddDefaultKey = %v", err)
	}
	if err := s.AddKey("zeta", kb); err != nil {
		t.Fatalf("AddKey(zeta) = %v", err)
	}
	if err := s.AddKey("alpha", ka); err != nil {
		t.Fatalf("AddKey(alpha) = %v", err)
	}

	got := s.ListBuckets()
	want := []string{"alpha", "default", "zeta"}
	if len(got) != len(want) {
		t.Fatalf("ListBuckets len = %d; want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ListBuckets[%d] = %q; want %q (full=%v)", i, got[i], want[i], got)
		}
	}
}

func TestSavePermissions(t *testing.T) {
	path := newTempStorePath(t)
	s := NewEncryptionKeyStore(path)

	if _, err := s.AddDefaultKey(""); err != nil {
		t.Fatalf("AddDefaultKey = %v", err)
	}
	if err := s.Save(); err != nil {
		t.Fatalf("Save() = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat file = %v", err)
	}
	mode := info.Mode().Perm()
	// Expect file to be owner-read/write only (0600). We allow umask-insensitive check:
	if mode&0o077 != 0 {
		t.Fatalf("file perms = %o; want no group/other bits set", mode)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	path := newTempStorePath(t)

	yamlWithUnknown := `
keys:
  default: "abc"
unknownField: "boom"
`
	if err := os.WriteFile(path, []byte(yamlWithUnknown), 0o600); err != nil {
		t.Fatalf("write = %v", err)
	}

	s := NewEncryptionKeyStore(path)
	err := s.Load()
	if err == nil {
		t.Fatalf("Load() = nil; want error due to unknown field")
	}
	if !strings.Contains(err.Error(), "unknown field") && !strings.Contains(err.Error(), "field") {
		t.Fatalf("Load() error = %v; want 'unknown field'", err)
	}
}

func TestGetNoDefaultNoBucket(t *testing.T) {
	path := newTempStorePath(t)
	s := NewEncryptionKeyStore(path)

	_, err := s.Get("missing")
	if err == nil {
		t.Fatalf("Get() without bucket and default = nil; want error")
	}
}

func TestHasKey(t *testing.T) {
	path := newTempStorePath(t)
	s := NewEncryptionKeyStore(path)

	if s.HasKey("x") {
		t.Fatalf("HasKey(x) = true; want false")
	}
	k, _ := GenerateEncryptionKey()
	if err := s.AddKey("x", k); err != nil {
		t.Fatalf("AddKey(x) = %v", err)
	}
	if !s.HasKey("x") {
		t.Fatalf("HasKey(x) = false; want true")
	}
}

func TestAtomicWriteOverwritesFully(t *testing.T) {
	// This test exercises that Save writes complete content,
	// by saving then re-loading and checking both keys exist.
	path := newTempStorePath(t)
	s := NewEncryptionKeyStore(path)

	if _, err := s.AddDefaultKey(""); err != nil {
		t.Fatalf("AddDefaultKey = %v", err)
	}
	k, _ := GenerateEncryptionKey()
	if err := s.AddKey("bucket", k); err != nil {
		t.Fatalf("AddKey(bucket) = %v", err)
	}
	if err := s.Save(); err != nil {
		t.Fatalf("Save() = %v", err)
	}

	// Overwrite with new state that still must be complete
	s2 := NewEncryptionKeyStore(path)
	if err := s2.Load(); err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if _, ok := s2.Keys["default"]; !ok {
		t.Fatalf("after reload, missing 'default' key")
	}
	if _, ok := s2.Keys["bucket"]; !ok {
		t.Fatalf("after reload, missing 'bucket' key")
	}
}
