package holder

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"github.com/AlexeySadkovich/eldberg/config"
	"github.com/AlexeySadkovich/eldberg/internal/utils"
)

type HolderService interface {
	Address() string
	PrivateKey() *ecdsa.PrivateKey
	PrivateKeyString() string
}

/*
Holder represents the user who hold the transaction or block
*/
type Holder struct {
	publicKey  *ecdsa.PublicKey
	privateKey *ecdsa.PrivateKey
	encoder    *base64.Encoding
}

var _ HolderService = (*Holder)(nil)

func New(config *config.Config) (*Holder, error) {
	holder := new(Holder)

	// Check if private key already exists
	// and if it doesn't then create new Holder
	// but if exists then restore Holder from
	// private key
	holderPrivateKey, err := utils.ReadHolderPrivateKey(config.Node.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	if holderPrivateKey == "" {
		holder, err = newHolder()
		if err != nil {
			return nil, fmt.Errorf("create new holder: %w", err)
		}

		privateKey := holder.PrivateKeyString()
		if err := utils.StoreHolderPrivateKey(config.Node.PrivateKeyPath, privateKey); err != nil {
			return nil, fmt.Errorf("store private key: %w", err)
		}
	} else {
		holder, err = restoreHolder(holderPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("restore holder: %w", err)
		}
	}

	return holder, nil
}

func (h *Holder) Address() string {
	pubBytes, err := x509.MarshalPKIXPublicKey(h.publicKey)
	if err != nil {
		return ""
	}

	return h.encoder.EncodeToString(pubBytes)
}

func (h *Holder) PrivateKey() *ecdsa.PrivateKey {
	return h.privateKey
}

func (h *Holder) PrivateKeyString() string {
	privBytes, err := x509.MarshalPKCS8PrivateKey(h.privateKey)
	if err != nil {
		return ""
	}

	return h.encoder.EncodeToString(privBytes)
}

func newHolder() (*Holder, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key failed: %w", err)
	}

	return &Holder{
		publicKey:  &(privateKey).PublicKey,
		privateKey: privateKey,
		encoder:    base64.StdEncoding,
	}, nil
}

func restoreHolder(privateKey string) (*Holder, error) {
	holder := &Holder{encoder: base64.StdEncoding}

	if err := parseCredentials(holder, privateKey); err != nil {
		return nil, fmt.Errorf("restore holder failed: %w", err)
	}

	return holder, nil
}

func parseCredentials(holder *Holder, privateKeyStr string) error {
	privateKeyBytes, err := holder.encoder.DecodeString(privateKeyStr)
	if err != nil {
		return fmt.Errorf("decoding private key failed: %w", err)
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		return fmt.Errorf("parsing private key failed: %w", err)
	}

	privateKey, ok := parsedKey.(*ecdsa.PrivateKey)
	if !ok {
		return fmt.Errorf("malformed private key")
	}

	holder.privateKey = privateKey
	holder.publicKey = &(privateKey).PublicKey

	return nil
}
