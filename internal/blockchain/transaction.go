package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/json"

	"github.com/AlexeySadkovich/eldberg/internal/blockchain/crypto"
	"github.com/AlexeySadkovich/eldberg/internal/blockchain/utils"
	"github.com/AlexeySadkovich/eldberg/internal/holder"
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

func (tx *Transaction) Bytes() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	_ = enc.Encode(tx)

	return buf.Bytes()
}

func (tx *Transaction) Serialize() string {
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return ""
	}

	return string(txBytes)
}

func (tx *Transaction) Deserialize(data string) error {
	return json.Unmarshal([]byte(data), &tx)
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
