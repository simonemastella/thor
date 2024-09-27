package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/schedule" // Sostituisci "yourproject" con il nome del tuo modulo
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/tx"
)

func newTx(chainTag byte, clauses []*tx.Clause, gas uint64, blockRef tx.BlockRef, expiration uint32, dependsOn *thor.Bytes32, features tx.Features, from genesis.DevAccount) *tx.Transaction {
	builder := new(tx.Builder).ChainTag(chainTag)
	for _, c := range clauses {
		builder.Clause(c)
	}

	tx := builder.BlockRef(blockRef).
		Expiration(expiration).
		Nonce(rand.Uint64()). // nolint:gosec
		DependsOn(dependsOn).
		Features(features).
		Gas(gas).Build()

	return signTx(tx, from)
}

func signTx(tx *tx.Transaction, acc genesis.DevAccount) *tx.Transaction {
	sig, _ := crypto.Sign(tx.SigningHash().Bytes(), acc.PrivateKey)
	return tx.WithSignature(sig)
}
func main() {
	acc := genesis.DevAccounts()[0]
	dbPath := filepath.Join(".", "data", "schedule.db")
	dbDir := filepath.Dir(dbPath)
	err := os.MkdirAll(dbDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	s, err := schedule.NewSchedule(dbPath)
	if err != nil {
		log.Fatalf("Failed to create schedule: %v", err)
	}
	if s == nil {
		log.Fatal("Schedule is nil after creation")
	}
	defer s.Close()

	// Inserisci alcuni elementi
	tx := newTx(0, nil, 21000, tx.BlockRef{}, 100, nil, tx.Features(0), acc)
	s.Push(tx, time.Now().Add(-2*time.Hour))
	s.Push(tx, time.Now().Add(-1*time.Hour))
	s.Push(tx, time.Now().Add(-4*time.Hour))
	s.Push(tx, time.Now().Add(-8*time.Hour))
	s.Push(tx, time.Now())

	fmt.Printf("LEN IS %v", s.Len())

	// Estrai gli elementi
	for {
		item, _ := s.Pop()
		if item == nil {
			break
		}
		fmt.Printf("Valore: %s, Data: %v\n", item.Tx.ID(), item.Date)
	}
}
