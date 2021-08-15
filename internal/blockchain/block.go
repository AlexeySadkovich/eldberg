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

func (b *Block) Serialize() string {
	blockStr, err := json.Marshal(b)
	if err != nil {
		return ""
	}

	return string(blockStr)
}

func (b *Block) Deserialize(data string) error {
	return json.Unmarshal([]byte(data), &b)
}
