package holder

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSign(t *testing.T) {
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, err)

	data := []byte("some data to sign")

	s := Sign(pk, data)
	require.NotNil(t, s)

	ok := Verify(&pk.PublicKey, s, data)
	require.True(t, ok)
}

func TestSignInvalid(t *testing.T) {
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, err)

	data := []byte("some data to sign")

	s := Sign(pk, data)
	require.NotNil(t, s)

	invalidData := []byte("some invalid data to verify")

	ok := Verify(&pk.PublicKey, s, invalidData)
	require.False(t, ok)
}
