package database

import (
	"go.etcd.io/bbolt"
	bboltErr "go.etcd.io/bbolt/errors"
)

// Bolt implements StorageBackend using bbolt.
type Bolt struct {
	db *bbolt.DB
}

// NewBolt creates a new Bolt.
func NewBolt(db *bbolt.DB) *Bolt {
	return &Bolt{db: db}
}

func (b *Bolt) Put(bucket, key, value string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucketObj, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return bucketObj.Put([]byte(key), []byte(value))
	})
}

func (b *Bolt) Get(bucket, key string) (string, error) {
	var value string
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucketObj := tx.Bucket([]byte(bucket))
		if bucketObj == nil {
			return bboltErr.ErrBucketNotFound
		}
		val := bucketObj.Get([]byte(key))

		if val == nil {
			return nil
		}
		value = string(val)
		return nil
	})
	return value, err
}

func (b *Bolt) Delete(bucket, key string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucketObj := tx.Bucket([]byte(bucket))
		if bucketObj == nil {
			return bboltErr.ErrBucketNotFound
		}
		return bucketObj.Delete([]byte(key))
	})
}

func (b *Bolt) CreateBucket(bucket string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})
}

func (b *Bolt) DeleteBucket(bucket string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket([]byte(bucket))
	})
}

func (b *Bolt) ListKeys(bucket string) ([]string, error) {
	var keys []string
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucketObj := tx.Bucket([]byte(bucket))
		if bucketObj == nil {
			return bboltErr.ErrBucketNotFound
		}
		return bucketObj.ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	return keys, err
}

func (b *Bolt) ListBuckets() ([]string, error) {
	var buckets []string
	err := b.db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, bucket *bbolt.Bucket) error {
			buckets = append(buckets, string(name))
			return nil
		})
	})
	return buckets, err
}

func (b *Bolt) BucketExists(bucket string) (bool, error) {
	var exists bool
	err := b.db.View(func(tx *bbolt.Tx) error {
		if tx.Bucket([]byte(bucket)) != nil {
			exists = true
		}
		return nil
	})
	return exists, err
}
