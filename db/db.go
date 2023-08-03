package db

import (
	"fmt"
	"log"
	"sync"

	bolt "go.etcd.io/bbolt"
)

const (
	bucket = "patches"
)

type BoltDB struct {
	db *bolt.DB
}

var instantiated *BoltDB

var once sync.Once

func NewBoltDB() *BoltDB {
	once.Do(func() {
		db, err := bolt.Open("my.db", 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		instantiated = &BoltDB{db}
	})

	return instantiated
}

func (b *BoltDB) Path() string {
	return b.db.Path()
}

func (b *BoltDB) Close() {
	b.db.Close()
}

func (b *BoltDB) Set(key string, val []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bu, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		if err := bu.Put([]byte(key), val); err != nil {
			return fmt.Errorf("put in patches : %s", err)
		}

		return nil
	})
}

func (b *BoltDB) Get(key string) []byte {
	var value []byte

	b.db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket([]byte(bucket))

		value = bu.Get([]byte(key))
		return nil
	})

	return value
}
