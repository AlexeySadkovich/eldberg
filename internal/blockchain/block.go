package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/AlexeySadkovich/eldberg/internal/blockchain/consensus"
	"github.com/AlexeySadkovich/eldberg/internal/blockchain/crypto"
	"github.com/AlexeySadkovich/eldberg/internal/blockchain/utils"
)

type Block struct {
	Header       *Header
	Transactions []*Transaction

	merkleTree *MerkleTree
}

func NewBlock(holderAddr string, prevHash []byte) *Block {
	header := &Header{
		TimeStamp:    utils.GetCurrentTimestamp(),
		PreviousHash: prevHash,
		Holder:       holderAddr,
	}

	merkleTree := NewMerkleTree([]*Transaction{})

	return &Block{
		Header:       header,
		Transactions: []*Transaction{},
		merkleTree:   merkleTree,
	}
}

func EmptyBlock() *Block {
	return new(Block)
}

func (b *Block) AddTransaction(tx *Transaction) error {
	if !tx.IsValid() {
		return fmt.Errorf("invalid transaction")
	}

	b.Transactions = append(b.Transactions, tx)

	b.merkleTree.AddTransaction(tx)
	b.Header.MerkleRoot = b.merkleTree.CalculateRoot()

	return nil
}

func (b *Block) IsValid() bool {
	if b == nil {
		return false
	}

	// Check that block were mined correctly
	blockHash := crypto.Hash(b.Bytes())
	if !consensus.CheckNonce(blockHash, b.Header.Nonce, b.Header.Difficulty) {
		return false
	}

	// TODO: add more validations

	return true
}

func (b *Block) Fullness() int {
	return len(b.Transactions)
}

func (b *Block) Bytes() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	_ = enc.Encode(b)

	return buf.Bytes()
}

func (b *Block) Serialize() ([]byte, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	return data, nil
}

func (b *Block) Deserialize(data []byte) error {
	err := json.Unmarshal(data, &b)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return nil
}
