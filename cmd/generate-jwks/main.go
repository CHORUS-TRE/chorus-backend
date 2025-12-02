package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"

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
}
