package main

import (
	"fmt"
	"math/rand"
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
    schedule := schedule.NewSchedule()

	// Inserisci alcuni elementi
	tx := newTx(0, nil, 21000, tx.BlockRef{}, 100, nil, tx.Features(0), acc)
	schedule.Push(tx, time.Now().Add(-2*time.Hour))
	schedule.Push(tx, time.Now().Add(-1*time.Hour))
	schedule.Push(tx, time.Now().Add(-4*time.Hour))
	schedule.Push(tx, time.Now().Add(-8*time.Hour))
	schedule.Push(tx, time.Now())

	item := schedule.Top()
	fmt.Printf("TOP Valore: %s, Data: %v\n", item.Tx.ID(), item.Date)

	// Estrai gli elementi
	for {
		item := schedule.Pop()
		if item == nil {
			break
		}
		fmt.Printf("Valore: %s, Data: %v\n", item.Tx.ID(), item.Date)
	}
}
