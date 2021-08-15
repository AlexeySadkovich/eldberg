package blockchain

import (
	"fmt"

	"github.com/AlexeySadkovich/eldberg/internal/blockchain/crypto"
	"github.com/AlexeySadkovich/eldberg/internal/holder"
	"github.com/AlexeySadkovich/eldberg/internal/storage"
)

type ChainService interface {
	AddBlock(block *Block) error

	GetHeight() int
	GetLastHash() ([]byte, error)
}

type Chain struct {
	CurrentBlock *Block
	Height       int
	storage      storage.StorageService
}

var _ ChainService = (*Chain)(nil)

func New(storage storage.StorageService, holder holder.HolderService) (*Chain, error) {
	chain := &Chain{
		storage: storage,
	}
	chain.Height = chain.GetHeight()

	// Create genesis block if chain is empty
	if chain.Height == 0 {
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
	c.Height++

	return nil
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
