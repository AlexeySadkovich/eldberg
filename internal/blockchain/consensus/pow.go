package consensus

import (
	"math"
	"math/big"
	"math/rand"

	"github.com/AlexeySadkovich/eldberg/internal/blockchain/crypto"
)

/*
StartPoW is implementation of ProofOfWork consensus algorithm

==== Draft ====
*/
func StartPoW(hash []byte, difficulty uint8) uint8 {
	nonce := uint8(rand.Intn(math.MaxUint8))

	target := big.NewInt(int64(difficulty))
	current := big.NewInt(1)

	for {
		hash = crypto.Hash(append(hash, nonce))
		current.SetBytes(hash)

		if current.Cmp(target) == -1 {
			return nonce
		}

		nonce++
	}
}

func CheckNonce(hash []byte, nonce, difficulty uint8) bool {
	d := big.NewInt(int64(difficulty))
	h := crypto.Hash(append(hash, nonce))

	proof := big.NewInt(1)
	proof.SetBytes(h)

	return proof.Cmp(d) == -1
}
