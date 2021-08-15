package crypto

import (
	"crypto/sha256"
)

func Hash(data []byte) []byte {
	hash := sha256.Sum256(data)

	return hash[:]
}
