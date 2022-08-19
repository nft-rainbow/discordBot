package database

import (
	"fmt"
	"github.com/boltdb/bolt"
)


var (
	db *bolt.DB
)

var EasyMintBucket = []byte("easy-mint-bucket")
var CustomMintBucket = []byte("custom-mint-bucket")

func ConnectDB(){
	var err error
	db, err = bolt.Open("./bolt.db", 0644, nil)
	if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error{
		_, err = tx.CreateBucketIfNotExists(EasyMintBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(CustomMintBucket)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func InsertDB(address string, val, bucketName []byte) error {
	key := []byte(address)

	err := db.Update(func(tx *bolt.Tx) error{
		bucket, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}

		err = bucket.Put(key, val)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func GetCount(address string, bucketName []byte) (string, error) {
	key := []byte(address)
	var val string

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf("Bucket %s not found!", bucketName)
		}
		val = string(bucket.Get(key))
		return nil
	})

	if err != nil {
		return "", err
	}
	return val, nil
}