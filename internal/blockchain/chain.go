package blockchain

import (
	"fmt"

	"github.com/AlexeySadkovich/eldberg/internal/blockchain/crypto"
	"github.com/AlexeySadkovich/eldberg/internal/holder"
	"github.com/AlexeySadkovich/eldberg/internal/storage"
)

type ChainService interface {
	AddBlock(block *Block) error

	SetCurrentBlock(block *Block)
	GetCurrentBlock() *Block

	GetHeight() int
	GetLastHash() ([]byte, error)
}

type Chain struct {
	currentBlock *Block
	height       int
	storage      storage.StorageService
}

var _ ChainService = (*Chain)(nil)

func New(storage storage.StorageService, holder holder.HolderService) (ChainService, error) {
	chain := &Chain{
		storage: storage,
	}
	chain.height = chain.GetHeight()

	// Create genesis block if chain is empty
	if chain.height == 0 {
		genesis := NewBlock(holder.Address(), []byte{})

		tx := NewTransaction(
			holder.Address(),
			holder.PrivateKey(),
			holder.Address(),
			50,
		)

		err := genesis.AddTransaction(tx)
		if err != nil {
			return nil, fmt.Errorf("creating genesis block failed: %w", err)
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
