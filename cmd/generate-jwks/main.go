package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/go-jose/go-jose/v4"
)

func main() {
	printPublicKey := flag.Bool("public-key", false, "also print the public key as a one-line PEM body (for registering with an external Keycloak instance)")
	flag.Parse()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("unable to generate RSA key: %v", err)
	}

	jwk := jose.JSONWebKey{
		Key:       privateKey,
		KeyID:     "chorus-backend-jwk-1",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}

	jwks := struct {
		Keys []jose.JSONWebKey `json:"keys"`
	}{Keys: []jose.JSONWebKey{jwk}}

	b, err := json.MarshalIndent(jwks, "", "  ")
	if err != nil {
		log.Fatalf("unable to marshal JWKS: %v", err)
	}
	fmt.Println(string(b))

	if *printPublicKey {
		der, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			log.Fatalf("unable to marshal public key: %v", err)
		}
		block := &pem.Block{Type: "PUBLIC KEY", Bytes: der}
		fmt.Println("\nPUBLIC KEY (Keycloak one-liner body):")
		fmt.Println(pemBodyOneLine(block))
	}
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
