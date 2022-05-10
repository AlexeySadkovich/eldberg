package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/AlexeySadkovich/eldberg/blockchain/crypto"
	"github.com/AlexeySadkovich/eldberg/blockchain/utils"
	"github.com/AlexeySadkovich/eldberg/holder"
)

type Transaction struct {
	Sender    string
	Receiver  string
	Value     float64
	TimeStamp string
	Hash      []byte
	Signature []byte
}

func NewTransaction(holderAddr string, holderPK *ecdsa.PrivateKey, receiver string, value float64) *Transaction {
	tx := &Transaction{
		Sender:    holderAddr,
		Receiver:  receiver,
		Value:     value,
		TimeStamp: utils.GetCurrentTimestamp(),
	}

	tx.Hash = crypto.Hash(tx.Bytes())
	tx.Signature = holder.Sign(holderPK, tx.Hash)

	return tx
}

func EmptyTransaction() *Transaction {
	return new(Transaction)
}

func (tx *Transaction) Bytes() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	_ = enc.Encode(tx)

	return buf.Bytes()
}

func (tx *Transaction) Serialize() ([]byte, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	return data, nil
}

func (tx *Transaction) Deserialize(data []byte) error {
	err := json.Unmarshal(data, &tx)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return nil
}

func (tx *Transaction) IsValid() bool {
	if len(tx.Sender) == 0 || len(tx.Receiver) == 0 {
		return false
	}
	if tx.Value <= 0 {
		return false
	}

	return true
}
