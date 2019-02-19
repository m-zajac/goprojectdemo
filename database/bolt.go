package database

import (
	"github.com/etcd-io/bbolt"
	"github.com/pkg/errors"
)

// BoltKVStore provides simple kv store interface based on boltdb.
type BoltKVStore struct {
	db         *bbolt.DB
	bucketName []byte
}

// NewBoltKVStore creates new BoltKVStore instance.
func NewBoltKVStore(dbPath string, bucketName string) (*BoltKVStore, error) {
	db, err := bbolt.Open(dbPath, 0666, nil)
	if err != nil {
		return nil, errors.Wrap(err, "opening database")
	}

	if err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "creating database bucket")
	}

	return &BoltKVStore{
		db:         db,
		bucketName: []byte(bucketName),
	}, nil
}

// ReadKey returns data saved for given key. Returns null if there's no data stored.
func (s *BoltKVStore) ReadKey(key []byte) ([]byte, error) {
	var data []byte
	if err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		data = b.Get(key)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "reading from db")
	}

	return data, nil
}

// UpdateKey stores given data under given key.
func (s *BoltKVStore) UpdateKey(key []byte, data []byte) error {
	if err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		return b.Put(key, data)
	}); err != nil {
		return errors.Wrap(err, "writing to db")
	}

	return nil
}

// Close closes database.
func (s *BoltKVStore) Close() error {
	return s.db.Close()
}
