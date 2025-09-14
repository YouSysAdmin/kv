// Package enckeystore provides a tiny, file-backed store for AES encryption keys,
// organized by "bucket" name with a special fallback "default" key.
//
// # Overview
//
// A Store is bound to a single YAML file on disk and exposes a small API:
//
//   - Load / Save: read and write the store (atomic write with fsync; 0600 perms)
//   - AddDefaultKey: set "default" once; does not replace if it already exists
//   - EnsureDefaultKey: create a fresh AES-256 default if missing
//   - AddKey: add a new, per-bucket key (no replacement allowed)
//   - Get: return the key for a bucket or the valid "default" fallback
//   - HasKey / ListBuckets: inspect what’s present
//
// All public methods are safe for concurrent use; the store guards internal
// state with an RWMutex. Disk I/O is explicit—callers decide when to Load and Save.
//
// # Key material and validation
//
// Keys are represented as EncryptionKey (a string). All additions and reads that
// surface keys validate via your encrypt package:
//
//   - EncryptionKey.Validate() -> uses encrypt.ValidateAESKey(string(k))
//   - GenerateEncryptionKey()  -> uses encrypt.GenerateRandomAESKey(encrypt.AES256)
//
// If a provided key is invalid, the operation fails with a wrapped error.
// Get() only returns keys that validate; otherwise it transparently tries
// the "default" key (also validated) before failing.
//
// # Persistence format
//
// The YAML file looks like:
//
//	keys:
//	  default: "<base64-or-hex-aes-key>"
//	  photos:  "<base64-or-hex-aes-key>"
//	  logs:    "<base64-or-hex-aes-key>"
//
// Unknown fields are rejected on Load() (yaml.KnownFields(true)), helping catch
// typos and format drift.
//
// # Atomic writes & permissions
//
// Save() performs an atomic, durable write sequence:
//
//  1. write to a temp file in the same directory
//  2. flush and fsync the temp file
//  3. rename over the target path
//  4. fsync the directory (best effort)
//
// The final file mode is 0600, and its parent directory is ensured (0700).
// This minimizes partial writes and reduces exposure of key material.
//
// # Replacement policy
//
// AddKey() and AddDefaultKey() do not replace existing entries. Replacements
// must be done manually by editing the on-disk YAML (or deleting and re-adding).
// This is intentional to avoid accidental key rotation. EnsureDefaultKey()
// never replaces an existing "default"—it only creates one if missing.
//
// # Example
//
// The following example shows a typical lifecycle: load (or initialize),
// ensure a default key, add a bucket key, read it, and save.
//
//	import (
//		"fmt"
//		"log"
//		"os"
//		"path/filepath"
//
//		"github.com/yousysadmin/kv/pkg/enckeystore"
//	)
//
//	func example() {
//		// Choose a path (for demo, a file in the temp dir).
//		path := filepath.Join(os.TempDir(), "keys.yaml")
//
//		store := enckeystore.NewEncryptionKeyStore(path)
//
//		// Load existing state; missing file is OK (treated as empty store).
//		if err := store.Load(); err != nil {
//			log.Fatalf("load: %v", err)
//		}
//
//		// Ensure there's a valid default (generates AES-256 if absent).
//		defKey, err := store.EnsureDefaultKey()
//		if err != nil {
//			log.Fatalf("ensure default: %v", err)
//		}
//		fmt.Println("default key exists:", defKey != "")
//
//		// Add a per-bucket key (no replacement if one already exists).
//		k, err := enckeystore.GenerateEncryptionKey()
//		if err != nil {
//			log.Fatalf("generate: %v", err)
//		}
//		if err := store.AddKey("photos", k); err != nil {
//			// If it already exists, this will error; that's expected by design.
//			log.Printf("add key (photos): %v", err)
//		}
//
//		// Read the key for a bucket (falls back to "default" if bucket missing).
//		got, err := store.Get("photos")
//		if err != nil {
//			log.Fatalf("get: %v", err)
//		}
//		_ = got // use the key
//
//		// Persist current state atomically with secure perms.
//		if err := store.Save(); err != nil {
//			log.Fatalf("save: %v", err)
//		}
//	}
//
// # Error handling & invariants
//
//   - Load() returns nil on a missing file and resets unknown fields.
//   - Get(bucket) fails if neither bucket nor a valid default are present.
//   - EnsureDefaultKey() returns an error if an existing default is present but invalid,
//     prompting a manual fix in the YAML (to avoid silent replacement).
//   - AddKey/AddDefaultKey validate the provided key first and never replace.
//
// # Security reminders
//
//   - Keep the YAML file on a trusted filesystem; Save() enforces 0600 but your
//     environment and backups still matter.
//   - Consider process memory exposure if logging keys; avoid printing key values.
//   - If you need rotation, design a deliberate process to edit the YAML and
//     roll keys carefully across dependent systems.
package enckeystore
