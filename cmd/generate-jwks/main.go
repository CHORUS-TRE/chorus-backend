package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/go-jose/go-jose/v4"
)

func main() {
	// Generate RSA keypair
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	jwk := jose.JSONWebKey{
		Key:       privateKey,
		KeyID:     "chorus-backend-jwk-1",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}

	b, _ := json.MarshalIndent(jwk, "", "  ")
	fmt.Println(string(b))

	pub := privateKey.PublicKey
	der, err := x509.MarshalPKIXPublicKey(&pub) // SubjectPublicKeyInfo
	if err != nil {
		log.Fatal(err)
	}

	block := &pem.Block{Type: "PUBLIC KEY", Bytes: der}
	fmt.Println("\nPUBLIC KEY (Keycloak one-liner body):")
	fmt.Println(pemBodyOneLine(block))
}

func pemBodyOneLine(block *pem.Block) string {
	p := pem.EncodeToMemory(block)
	// p includes headers + line breaks; strip them
	out := make([]byte, 0, len(p))
	lines := bytesSplitLines(p)
	for _, ln := range lines {
		if len(ln) == 0 {
			continue
		}
		if ln[0] == '-' { // BEGIN/END
			continue
		}
		out = append(out, ln...)
	}
	return string(out)
}

func bytesSplitLines(b []byte) [][]byte {
	var res [][]byte
	start := 0
	for i := 0; i < len(b); i++ {
		if b[i] == '\n' {
			res = append(res, b[start:i])
			start = i + 1
		}
	}
	if start < len(b) {
		res = append(res, b[start:])
	}
	return res
}
