package storage_test

import (
	"errors"
	"fmt"
	"github.com/yousysadmin/kv/internal/storage"
	"github.com/yousysadmin/kv/pkg/encrypt"
	"go.etcd.io/bbolt"
	bboltErr "go.etcd.io/bbolt/errors"
	"os"
	"testing"
)

func setupTestDB(t *testing.T) (*bbolt.DB, func()) {
	dbFile := "test.db"
	db, err := bbolt.Open(dbFile, 0600, nil)
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	return db, func() {
		db.Close()
		os.Remove(dbFile)
	}
}

func mustGenKey(t *testing.T) string {
	k, err := encrypt.GenerateRandomAESKey(encrypt.AES256)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	return k
}

func TestAddAndGet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	s := storage.NewEntityStorage(db, mustGenKey(t))
	err := s.Add(storage.DefaultBucket, "foo", "bar")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	val, err := s.Get(storage.DefaultBucket, "foo")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if val != "bar" {
		t.Errorf("Expected 'bar', got '%s'", val)
	}
}

func TestAddEmptyValue(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	s := storage.NewEntityStorage(db, mustGenKey(t))
	err := s.Add(storage.DefaultBucket, "empty", "")
	if err != nil {
		t.Fatalf("Add empty failed: %v", err)
	}

	val, err := s.Get(storage.DefaultBucket, "empty")
	if err != nil || val != "" {
		t.Fatalf("Expected empty string, got: '%s', err: %v", val, err)
	}
}

func TestGetNonExistentKey(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	s := storage.NewEntityStorage(db, mustGenKey(t))
	_, err := s.Get(storage.DefaultBucket, "missing")
	if err == nil {
		t.Fatal("Expected error for missing key")
	}
}

func TestDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	s := storage.NewEntityStorage(db, mustGenKey(t))
	_ = s.Add(storage.DefaultBucket, "deleteMe", "123")

	err := s.Delete(storage.DefaultBucket, "deleteMe")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = s.Get(storage.DefaultBucket, "deleteMe")
	if err == nil {
		t.Fatal("Expected error after delete")
	}
}

func TestListKeysOnly(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	s := storage.NewEntityStorage(db, mustGenKey(t))
	_ = s.Add(storage.DefaultBucket, "k1", "v1")
	_ = s.Add(storage.DefaultBucket, "k2", "v2")

	items, err := s.List(storage.DefaultBucket, false)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(items))
	}
}

func TestListWithValues(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	s := storage.NewEntityStorage(db, mustGenKey(t))
	_ = s.Add(storage.DefaultBucket, "k1", "v1")
	_ = s.Add(storage.DefaultBucket, "k2", "v2")

	items, err := s.List(storage.DefaultBucket, true)
	if err != nil {
		t.Fatalf("List with values failed: %v", err)
	}

	for _, item := range items {
		if item.Value == "" {
			t.Errorf("Expected value for key %s, got empty", item.Key)
		}
	}
}

func TestListBuckets(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	s := storage.NewEntityStorage(db, mustGenKey(t))
	_ = s.Add("bucket1", "a", "1")
	_ = s.Add("bucket2", "b", "2")

	buckets, err := s.ListBuckets()
	if err != nil {
		t.Fatalf("ListBuckets failed: %v", err)
	}

	if len(buckets) < 2 {
		t.Errorf("Expected at least 2 buckets, got %d", len(buckets))
	}
}

func TestDeleteBucket(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	s := storage.NewEntityStorage(db, mustGenKey(t))
	_ = s.Add("toDelete", "foo", "bar")

	err := s.DeleteBucket("toDelete")
	if err != nil {
		t.Fatalf("DeleteBucket failed: %v", err)
	}

	_, err = s.Get("toDelete", "foo")
	if !errors.Is(err, bboltErr.ErrBucketNotFound) {
		t.Errorf("Expected ErrBucketNotFound, got: %v", err)
	}
}

func BenchmarkAdd(b *testing.B) {
	db, cleanup := setupTestDB(&testing.T{})
	defer cleanup()

	s := storage.NewEntityStorage(db, mustGenKey(&testing.T{}))
	for i := 0; i < b.N; i++ {
		_ = s.Add(storage.DefaultBucket, "benchkey", "benchvalue")
	}
}

func BenchmarkGet(b *testing.B) {
	db, cleanup := setupTestDB(&testing.T{})
	defer cleanup()

	s := storage.NewEntityStorage(db, mustGenKey(&testing.T{}))
	_ = s.Add(storage.DefaultBucket, "benchkey", "benchvalue")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Get(storage.DefaultBucket, "benchkey")
	}
}

func BenchmarkList(b *testing.B) {
	db, cleanup := setupTestDB(&testing.T{})
	defer cleanup()

	s := storage.NewEntityStorage(db, mustGenKey(&testing.T{}))
	for i := 0; i < 1000; i++ {
		_ = s.Add(storage.DefaultBucket, fmt.Sprintf("key%d", i), "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.List(storage.DefaultBucket, true)
	}
}
