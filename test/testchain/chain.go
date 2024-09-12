// Copyright (c) 2024 The VeChainThor developers

// Distributed under the GNU Lesser General Public License v3.0 software license, see the accompanying
// file LICENSE or <https://www.gnu.org/licenses/lgpl-3.0.html>

package testchain

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/vechain/thor/v2/bft"
	"github.com/vechain/thor/v2/block"
	"github.com/vechain/thor/v2/chain"
	"github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/muxdb"
	"github.com/vechain/thor/v2/packer"
	"github.com/vechain/thor/v2/state"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/tx"
)

// Chain represents the blockchain structure.
// It includes database (db), genesis information (genesis), consensus engine (engine),
// repository for blocks and state (repo), state manager (stater), and the genesis block (genesisBlock).
type Chain struct {
	db           *muxdb.MuxDB
	genesis      *genesis.Genesis
	engine       bft.Committer
	repo         *chain.Repository
	stater       *state.Stater
	genesisBlock *block.Block
}

// Repo returns the blockchain's repository, which stores blocks and other data.
func (c *Chain) Repo() *chain.Repository {
	return c.repo
}

// Stater returns the current state manager of the chain, which is responsible for managing the state of accounts and other elements.
func (c *Chain) Stater() *state.Stater {
	return c.stater
}

// Engine returns the consensus engine responsible for the blockchain's consensus mechanism.
func (c *Chain) Engine() bft.Committer {
	return c.engine
}

// GenesisBlock returns the genesis block of the chain, which is the first block in the blockchain.
func (c *Chain) GenesisBlock() *block.Block {
	return c.genesisBlock
}

// MintTransactions creates a block with the provided transactions and adds it to the blockchain.
// It wraps the transactions with receipts and passes them to MintTransactionsWithReceiptFunc.
func (c *Chain) MintTransactions(account genesis.DevAccount, transactions ...*tx.Transaction) error {
	var txAndRcpts []*TxAndRcpt
	for _, transaction := range transactions {
		txAndRcpts = append(txAndRcpts, &TxAndRcpt{Transaction: transaction})
	}
	return c.MintTransactionsWithReceiptFunc(account, txAndRcpts...)
}

// MintTransactionsWithReceiptFunc mints a block by accepting transactions and their associated receipt functions.
// It calls MintBlock to finalize the process.
func (c *Chain) MintTransactionsWithReceiptFunc(account genesis.DevAccount, txAndRcpts ...*TxAndRcpt) error {
	return c.MintBlock(account, txAndRcpts...)
}

// MintBlock creates and finalizes a new block with the given transactions and receipts.
// It schedules a new block, adopts transactions, packs them into a block, and commits it to the chain.
func (c *Chain) MintBlock(account genesis.DevAccount, txAndRcpts ...*TxAndRcpt) error {
	// Create a new block packer with the current chain state and account information.
	blkPacker := packer.New(c.Repo(), c.Stater(), account.Address, &genesis.DevAccounts()[0].Address, thor.NoFork)

	// Schedule a new block to be packed at the current time.
	blkFlow, err := blkPacker.Schedule(c.Repo().BestBlockSummary(), uint64(time.Now().Unix()))
	if err != nil {
		return fmt.Errorf("unable to schedule a new block: %w", err)
	}

	// Adopt the provided transactions into the block.
	for _, txAndRcpt := range txAndRcpts {
		if err = blkFlow.Adopt(txAndRcpt.Transaction); err != nil {
			return fmt.Errorf("unable to adopt tx into block: %w", err)
		}
	}

	// Pack the adopted transactions into a block.
	newBlk, stage, receipts, err := blkFlow.Pack(account.PrivateKey, 0, false)
	if err != nil {
		return fmt.Errorf("unable to pack tx: %w", err)
	}

	// Execute any receipt modification functions if provided.
	for _, txAndRcpt := range txAndRcpts {
		if txAndRcpt.ReceiptFunc != nil {
			txAndRcpt.ReceiptFunc(receipts)
		}
	}

	// Commit the new block to the chain's state.
	if _, err := stage.Commit(); err != nil {
		return fmt.Errorf("unable to commit tx: %w", err)
	}

	// Add the block to the repository.
	if err := c.Repo().AddBlock(newBlk, receipts, 0); err != nil {
		return fmt.Errorf("unable to add tx to repo: %w", err)
	}

	// Set the new block as the best (latest) block in the repository.
	if err := c.Repo().SetBestBlockID(newBlk.Header().ID()); err != nil {
		return fmt.Errorf("unable to set best block: %w", err)
	}

	return nil
}

// GetAllBlocks retrieves all blocks from the blockchain, starting from the best block and moving backward to the genesis block.
// It limits the retrieval time to 5 seconds to avoid excessive delays.
func (c *Chain) GetAllBlocks() ([]*block.Block, error) {
	bestBlkSummary := c.Repo().BestBlockSummary()
	var blks []*block.Block
	currBlockID := bestBlkSummary.Header.ID()
	startTime := time.Now()

	// Traverse the chain backwards until the genesis block is reached or timeout occurs.
	for {
		blk, err := c.repo.GetBlock(currBlockID)
		if err != nil {
			return nil, err
		}
		blks = append(blks, blk)

		// Stop when the genesis block is reached and reverse the slice to have genesis at position 0.
		if blk.Header().Number() == c.genesisBlock.Header().Number() {
			slices.Reverse(blks) // make sure genesis is at position 0
			return blks, err
		}
		currBlockID = blk.Header().ParentID()

		// Check if the retrieval process is taking too long (more than 5 seconds).
		if time.Since(startTime) > 5*time.Second {
			return nil, errors.New("taking more than 5 seconds to retrieve all blocks")
		}
	}
}

// BestBlock returns the current best (latest) block in the chain.
func (c *Chain) BestBlock() (*block.Block, error) {
	return c.Repo().GetBlock(c.Repo().BestBlockSummary().Header.ID())
}

// GetForkConfig returns the current fork configuration based on the ID of the genesis block.
func (c *Chain) GetForkConfig() thor.ForkConfig {
	return thor.GetForkConfig(c.GenesisBlock().Header().ID())
}
