package blockchain

import (
	"fmt"

	"github.com/AlexeySadkovich/eldberg/config"
	"github.com/AlexeySadkovich/eldberg/internal/blockchain/crypto"
	"github.com/AlexeySadkovich/eldberg/internal/holder"
	"github.com/AlexeySadkovich/eldberg/internal/storage"
)

type Chain struct {
	currentBlock *Block
	height       int
	storage      storage.Storage
	holder       *holder.Holder
}

func New(storage storage.Storage, holder *holder.Holder, config config.Config) (*Chain, error) {
	chain := &Chain{
		storage: storage,
		holder:  holder,
	}
	chain.height = chain.GetHeight()

	// Create genesis block if chain is empty
	if chain.height == 0 {
		value := config.GetChainConfig().Genesis.Value
		if err := chain.CreateGenesis(value); err != nil {
			return nil, fmt.Errorf("create genesis block: %w", err)
		}
	}

	return chain, nil
}

func (c *Chain) AddBlock(block *Block) error {
	data := block.Bytes()
	hash := crypto.Hash(data)

	if err := c.storage.Chain().Save(hash, data); err != nil {
		return fmt.Errorf("add block: %w", err)
	}

	// Increase chain height on block addition
	// to avoid iterating through all chain
	// when need to get height
	c.height++

	return nil
}

func (c *Chain) SetCurrentBlock(block *Block) {
	c.currentBlock = block
}

func (c *Chain) GetCurrentBlock() *Block {
	return c.currentBlock
}

func (c *Chain) GetHeight() int {
	return c.storage.Chain().GetHeight()
}

func (c *Chain) GetLastHash() ([]byte, error) {
	hash, err := c.storage.Chain().GetLastHash()
	if err != nil {
		return nil, fmt.Errorf("get last hash: %w", err)
	}

	return hash, nil
}

func (c *Chain) CreateGenesis(value float64) error {
	genesis := NewBlock(c.holder.Address(), []byte{})

	tx := NewTransaction(
		c.holder.Address(),
		c.holder.PrivateKey(),
		c.holder.Address(),
		value,
	)

	if err := genesis.AddTransaction(tx); err != nil {
		return fmt.Errorf("add transaction: %w", err)
	}

	return nil
}
