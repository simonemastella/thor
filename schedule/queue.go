package schedule

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/boltdb/bolt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vechain/thor/v2/tx"
)

var bucketName = []byte("transactions")

type Item struct {
	Tx             *tx.Transaction
	Date           time.Time
	InsertionOrder uint64
}
type SerializableItem struct {
	TxBytes        []byte
	Date           time.Time
	InsertionOrder uint64
}

func (i Item) MarshalJSON() ([]byte, error) {
	txBytes, err := rlp.EncodeToBytes(i.Tx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(SerializableItem{
		TxBytes:        txBytes,
		Date:           i.Date,
		InsertionOrder: i.InsertionOrder,
	})
}

func (i *Item) UnmarshalJSON(data []byte) error {
    var si SerializableItem
    if err := json.Unmarshal(data, &si); err != nil {
        return err
    }
    tx := new(tx.Transaction)
    if err := rlp.DecodeBytes(si.TxBytes, tx); err != nil {
        return err
    }
    i.Tx = tx
    i.Date = si.Date
    i.InsertionOrder = si.InsertionOrder
    return nil
}

type Schedule struct {
	db               *bolt.DB
	insertionCounter uint64
	itemCount        int64 // Nuovo campo per il conteggio
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
		db.Close()
		return nil, err
	}

	return &Schedule{db: db, insertionCounter: 0}, nil
}

func (s *Schedule) Push(tx *tx.Transaction, date time.Time) error {
	insertionOrder := atomic.AddUint64(&s.insertionCounter, 1)
	item := Item{Tx: tx, Date: date, InsertionOrder: insertionOrder}
	value, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return s.db.Update(func(btx *bolt.Tx) error {
		b := btx.Bucket(bucketName)
		key := make([]byte, 16)
		binary.BigEndian.PutUint64(key[:8], uint64(date.UnixNano()))
		binary.BigEndian.PutUint64(key[8:], insertionOrder)
		err := b.Put(key, value)
		if err == nil {
			atomic.AddInt64(&s.itemCount, 1) // Incrementa il conteggio
		}
		return err
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

		err = b.Delete(k)
		if err == nil {
			atomic.AddInt64(&s.itemCount, -1) // Decrementa il conteggio
		}
		return err
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

func (s *Schedule) Len() int {
	return int(atomic.LoadInt64(&s.itemCount))
}

func (s *Schedule) Close() error {
	return s.db.Close()
}
