package pow

import (
	"math"
	"math/big"
	"math/rand"

	"github.com/AlexeySadkovich/eldberg/internal/blockchain/crypto"
	"github.com/AlexeySadkovich/eldberg/internal/consensus"
)

type pow struct {
	hash   []byte
	target *big.Int
	nonce  uint8

	done chan struct{}
}

// Check Algorithm interface
var _ consensus.Algorithm = (*pow)(nil)

func New(hash []byte, difficulty uint8) consensus.Algorithm {
	nonce := uint8(rand.Intn(math.MaxUint8))
	target := big.NewInt(int64(difficulty))

	pow := &pow{
		hash:   hash,
		target: target,
		nonce:  nonce,
		done:   make(chan struct{}),
	}

	return pow
}

func (pow *pow) Done() <-chan struct{} {
	return pow.done
}

func (pow *pow) Run() {
	current := big.NewInt(1)
	hash := pow.hash

	for {
		hash = crypto.Hash(append(hash, pow.nonce))
		current.SetBytes(hash)

		if current.Cmp(pow.target) == -1 {
			pow.Stop()
			return
		}

		pow.nonce++
	}
}

func (pow *pow) Stop() {
	close(pow.done)
}

func CheckNonce(hash []byte, nonce, difficulty uint8) bool {
	d := big.NewInt(int64(difficulty))
	h := crypto.Hash(append(hash, nonce))

	proof := big.NewInt(1)
	proof.SetBytes(h)

	return proof.Cmp(d) == -1
}
