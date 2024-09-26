package schedule

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/vechain/thor/v2/tx"
)

var bucketName = []byte("transactions")

type Item struct {
	Tx   *tx.Transaction
	Date time.Time
}

type Schedule struct {
	db *bolt.DB
}

func NewSchedule(dbPath string) (*Schedule, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		return nil
	})
	if err != nil {
		db.Close() // Chiudi il database se c'è un errore
		return nil, err
	}

	return &Schedule{db: db}, nil
}

func (s *Schedule) Push(tx *tx.Transaction, date time.Time) error {
	item := Item{Tx: tx, Date: date}
	value, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return s.db.Update(func(btx *bolt.Tx) error {
		b := btx.Bucket(bucketName)
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(date.UnixNano()))
		return b.Put(key, value)
	})
}

func (s *Schedule) Pop() (*Item, error) {
	var item *Item

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()
		k, v := c.First()
		if k == nil {
			return nil // No items in the database
		}

		err := json.Unmarshal(v, &item)
		if err != nil {
			return err
		}

		return b.Delete(k)
	})

	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Schedule) Top() (*Item, error) {
	var item *Item

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()
		k, v := c.First()
		if k == nil {
			return nil // No items in the database
		}

		return json.Unmarshal(v, &item)
	})

	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Schedule) Len() (int, error) {
	var count int

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			count++
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Schedule) Close() error {
	return s.db.Close()
}