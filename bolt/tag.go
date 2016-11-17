package bolt

import (
	"bytes"
	"log"

	"github.com/boltdb/bolt"
)

var tagBucket = []byte("tags")

type TagSearcher struct {
	Driver *Driver
}

func (s *TagSearcher) Index(tag string) error {
	return s.Driver.store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(tagBucket)
		return bucket.Put([]byte(tag), []byte(tag))
	})
}

func (s *TagSearcher) Search(prefix string) ([]string, error) {
	tags := make([]string, 0)

	err := s.Driver.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(tagBucket)
		c := bucket.Cursor()

		if len(prefix) == 0 {
			for tag, _ := c.First(); tag != nil; tag, _ = c.Next() {
				tags = append(tags, string(tag))
			}
		} else {
			for tag, _ := c.Seek([]byte(prefix)); bytes.HasPrefix(tag, []byte(prefix)); tag, _ = c.Next() {
				tags = append(tags, string(tag))
				log.Println(tag)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tags, nil
}
