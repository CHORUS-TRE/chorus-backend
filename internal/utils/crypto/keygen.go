package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// GeneratePrivateKeyPEM generates a fresh P-256 EC private key and returns
// its PEM encoding.
func GeneratePrivateKeyPEM() (string, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("unable to generate EC private key: %w", err)
	}

	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("unable to marshal EC private key: %w", err)
	}

	block := &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}
	return string(pem.EncodeToMemory(block)), nil
}
