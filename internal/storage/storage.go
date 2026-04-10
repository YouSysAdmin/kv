package storage

import (
	"errors"
	"fmt"

	"github.com/yousysadmin/kv/internal/models"
	"github.com/yousysadmin/kv/pkg/encrypt"
)

const DefaultBucket = "default"

var (
	ErrValueIsEmpty = errors.New("key not found or value is empty")
)

// StorageBackend defines the interface for underlying database operations.
type StorageBackend interface {
	Put(bucket, key, value string) error
	Get(bucket, key string) (string, error)
	Delete(bucket, key string) error
	CreateBucket(bucket string) error
	DeleteBucket(bucket string) error
	ListKeys(bucket string) ([]string, error)
	ListBuckets() ([]string, error)
	BucketExists(bucket string) (bool, error)
}

// EntityStorage persists Entity data in the database using a backend.
type EntityStorage struct {
	backend       StorageBackend
	encryptionKey string
}

// NewEntityStorage creates a new EntityStorage with the provided backend.
func NewEntityStorage(backend StorageBackend, encryptionKey string) *EntityStorage {
	return &EntityStorage{backend: backend, encryptionKey: encryptionKey}
}

// Add inserts and encrypts a key-value pair into the specified bucket.
func (d *EntityStorage) Add(bucket string, key string, value string) error {
	aes := encrypt.NewAES(d.encryptionKey, value)
	encValue, err := aes.Encrypt()
	if err != nil {
		return fmt.Errorf("failed to encrypt value: %w", err)
	}

	return d.backend.Put(bucket, key, encValue)
}

// Get retrieves and decrypts the value associated with the given key in the specified bucket.
func (d *EntityStorage) Get(bucket string, key string) (string, error) {
	encValue, err := d.backend.Get(bucket, key)
	if err != nil {
		return "", err
	}
	if encValue == "" {
		return "", ErrValueIsEmpty
	}

	aes := encrypt.NewAES(d.encryptionKey, encValue)
	decValue, err := aes.Decrypt()
	if err != nil {
		return "", err
	}
	return decValue, nil
}

// Delete removes the key-value pair from the specified bucket.
func (d *EntityStorage) Delete(bucket string, key string) error {
	return d.backend.Delete(bucket, key)
}

// List returns all keys in the specified bucket, optionally including decrypted values.
func (d *EntityStorage) List(bucket string, withValues bool) ([]models.Entity, error) {
	keys, err := d.backend.ListKeys(bucket)
	if err != nil {
		return nil, err
	}

	var entries []models.Entity
	for _, key := range keys {
		if withValues {
			val, err := d.Get(bucket, key)
			if err != nil {
				return nil, err
			}
			entries = append(entries, models.Entity{Key: key, Value: val})
		} else {
			entries = append(entries, models.Entity{Key: key, Value: ""})
		}
	}
	return entries, nil
}

// AddBucket add new bucket.
func (d *EntityStorage) AddBucket(name string) error {
	return d.backend.CreateBucket(name)
}

// ListBuckets returns the names of all buckets in the database.
func (d *EntityStorage) ListBuckets() ([]string, error) {
	return d.backend.ListBuckets()
}

// BucketExist check bucket existing.
func (d *EntityStorage) BucketExist(name string) (bool, error) {
	return d.backend.BucketExists(name)
}

// DeleteBucket removes the specified bucket from the database.
func (d *EntityStorage) DeleteBucket(bucket string) error {
	return d.backend.DeleteBucket(bucket)
}
