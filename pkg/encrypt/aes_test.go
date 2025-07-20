package encrypt_test

import (
	"github.com/yousysadmin/kv/pkg/encrypt"
	"strings"
	"testing"
)

func TestGenerateRandomAESKey(t *testing.T) {
	for _, size := range []encrypt.AESKeySize{encrypt.AES128, encrypt.AES192, encrypt.AES256} {
		key, err := encrypt.GenerateRandomAESKey(size)
		if err != nil {
			t.Errorf("GenerateRandomAESKey failed for size %d: %v", size, err)
			continue
		}
		if len(key) != int(size) {
			t.Errorf("Expected key length %d, got %d", size, len(key))
		}
	}
}

func TestValidateAESKey(t *testing.T) {
	valid := []string{
		strings.Repeat("a", 16),
		strings.Repeat("b", 24),
		strings.Repeat("c", 32),
	}
	invalid := []string{
		"short", strings.Repeat("z", 20),
	}

	for _, key := range valid {
		if err := encrypt.ValidateAESKey(key); err != nil {
			t.Errorf("expected valid key, got error: %v", err)
		}
	}

	for _, key := range invalid {
		if err := encrypt.ValidateAESKey(key); err == nil {
			t.Errorf("expected error for key %q, got nil", key)
		}
	}
}

func TestEncryptDecryptAES(t *testing.T) {
	key, _ := encrypt.GenerateRandomAESKey(encrypt.AES256)
	plain := "Secret text to encrypt"

	a := encrypt.NewAES(key, plain)
	ciphertext, err := a.Encrypt()
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if !strings.HasPrefix(ciphertext, encrypt.PrefixAES256) {
		t.Errorf("expected prefix %q, got %q", encrypt.PrefixAES256, ciphertext[:8])
	}

	b := encrypt.NewAES(key, ciphertext)
	decrypted, err := b.Decrypt()
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != plain {
		t.Errorf("expected %q, got %q", plain, decrypted)
	}
}

func TestDecryptWithWrongPrefix(t *testing.T) {
	key256 := strings.Repeat("a", 32)
	wrong := strings.Repeat("b", 24)

	a := encrypt.NewAES(key256, encrypt.PrefixAES192+"xyzdata")
	_, err := a.Decrypt()
	if err == nil {
		t.Error("expected prefix mismatch error, got nil")
	}

	b := encrypt.NewAES(wrong, encrypt.PrefixAES192+"xyzdata")
	_, err = b.Decrypt()
	if err == nil {
		t.Error("expected key length error, got nil")
	}
}

func BenchmarkEncryptAES(b *testing.B) {
	key, _ := encrypt.GenerateRandomAESKey(encrypt.AES256)
	data := strings.Repeat("data1234", 100) // ~800 bytes

	for i := 0; i < b.N; i++ {
		a := encrypt.NewAES(key, data)
		if _, err := a.Encrypt(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecryptAES(b *testing.B) {
	key, _ := encrypt.GenerateRandomAESKey(encrypt.AES256)
	data := strings.Repeat("data1234", 100)

	a := encrypt.NewAES(key, data)
	ciphertext, err := a.Encrypt()
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		d := encrypt.NewAES(key, ciphertext)
		if _, err := d.Decrypt(); err != nil {
			b.Fatal(err)
		}
	}
}
