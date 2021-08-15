package holder

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"math/big"
)

func Sign(privateKey *ecdsa.PrivateKey, hash []byte) []byte {
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return nil
	}

	data := &signature{r, s}

	signature, err := data.serialize()
	if err != nil {
		return nil
	}

	return signature
}

func Verify(publicKey *ecdsa.PublicKey, sign, hash []byte) bool {
	data := &signature{}
	if err := data.deserialize(sign); err != nil {
		return false
	}

	return ecdsa.Verify(publicKey, hash, data.R, data.S)
}

type signature struct {
	R *big.Int
	S *big.Int
}

func (s *signature) serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(s); err != nil {
		return nil, fmt.Errorf("encoding sinature failed: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *signature) deserialize(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)

	return dec.Decode(s)
}
