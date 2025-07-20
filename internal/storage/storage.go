package storage

import (
	"errors"
	"github.com/yousysadmin/kv/pkg/encrypt"
	"go.etcd.io/bbolt"
	bboltErr "go.etcd.io/bbolt/errors"
)

const DefaultBucket = "default"

type Entity struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

var (
	ErrValueIsEmpty = errors.New("key not found or value is empty")
)

// EntityStorage persists Entity data in the database.
type EntityStorage struct {
	db            *bbolt.DB
	encryptionKey string
}

// NewEntityStorage creates a new EntityStorage.
func NewEntityStorage(db *bbolt.DB, encryptionKey string) *EntityStorage {
	return &EntityStorage{db: db, encryptionKey: encryptionKey}
}

// Add inserts and encrypts a key-value pair into the specified bucket.
func (d *EntityStorage) Add(bucket string, key string, value string) error {
	aes := encrypt.NewAES(d.encryptionKey, value)
	encValue, err := aes.Encrypt()
	if err != nil {
		return err
	}

	return d.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), []byte(encValue))
	})
}

// Get retrieves and decrypts the value associated with the given key in the specified bucket.
func (d *EntityStorage) Get(bucket string, key string) (string, error) {
	var value string
	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bboltErr.ErrBucketNotFound
		}
		encValue := b.Get([]byte(key))
		if encValue == nil {
			return ErrValueIsEmpty
		}
		aes := encrypt.NewAES(d.encryptionKey, string(encValue))
		decValue, err := aes.Decrypt()
		if err != nil {
			return err
		}
		value = decValue
		return nil
	})
	return value, err
}

// Delete removes the key-value pair from the specified bucket.
func (d *EntityStorage) Delete(bucket string, key string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bboltErr.ErrBucketNotFound
		}
		return b.Delete([]byte(key))
	})
}

// List returns all keys in the specified bucket, optionally including decrypted values.
func (d *EntityStorage) List(bucket string, withValues bool) ([]Entity, error) {
	var entries []Entity

	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bboltErr.ErrBucketNotFound
		}
		b.ForEach(func(k, v []byte) error {
			if withValues {
				aes := encrypt.NewAES(d.encryptionKey, string(v))
				decValue, err := aes.Decrypt()
				if err != nil {
					return err
				}
				entries = append(entries, Entity{string(k), decValue})
			} else {
				entries = append(entries, Entity{string(k), ""})
			}

			return nil
		})
		return nil
	})
	return entries, err
}

// ListBuckets returns the names of all buckets in the database.
func (d *EntityStorage) ListBuckets() ([]string, error) {
	var buckets []string
	err := d.db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(n []byte, b *bbolt.Bucket) error {
			buckets = append(buckets, string(n))
			return nil
		})
	})
	return buckets, err
}

// DeleteBucket removes the specified bucket from the database.
func (d *EntityStorage) DeleteBucket(bucket string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket([]byte(bucket))
	})
}
