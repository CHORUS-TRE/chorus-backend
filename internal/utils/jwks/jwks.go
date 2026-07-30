package jwks

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/go-jose/go-jose/v4"
)

// Generate creates a fresh RSA keypair wrapped in a JWKS document,
// plus the PEM public key body
func Generate() (jwksJSON string, publicKeyPEMBody string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("unable to generate RSA key: %w", err)
	}

	jwk := jose.JSONWebKey{
		Key:       privateKey,
		KeyID:     "chorus-backend-jwk-1",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}

	document := struct {
		Keys []jose.JSONWebKey `json:"keys"`
	}{Keys: []jose.JSONWebKey{jwk}}

	b, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("unable to marshal JWKS: %w", err)
	}

	der, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("unable to marshal public key: %w", err)
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: der}

	return string(b), pemBodyOneLine(block), nil
}

// pemBodyOneLine strips a PEM block's BEGIN/END header/footer and newlines,
// leaving just the base64 body as a single line.
func pemBodyOneLine(block *pem.Block) string {
	var sb strings.Builder
	for line := range strings.SplitSeq(string(pem.EncodeToMemory(block)), "\n") {
		if line == "" || strings.HasPrefix(line, "-----") {
			continue
		}
		sb.WriteString(line)
	}
	return sb.String()
}
