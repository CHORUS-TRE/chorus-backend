package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math"

	"golang.org/x/crypto/pbkdf2"
)

type Secret struct {
	EncSecret []byte
	Key       []byte
	Salt      []byte
}

func NewSecret(secret []byte) (*Secret, error) {

	key := make([]byte, 32)
	if _, err := rand.Reader.Read(key); err != nil {
		return nil, fmt.Errorf("unable to generate Key: random reader failed: %w", err)
	}

	salt := make([]byte, 32)
	if _, err := rand.Reader.Read(salt); err != nil {
		return nil, fmt.Errorf("unable to generate Salt: random reader failed: %w", err)
	}
	dk := Derive(key, salt)
	enc, err := Encrypt(secret, dk)
	Zero(secret)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt secret: %w", err)
	}

	return &Secret{EncSecret: enc, Key: key, Salt: salt}, nil
}

func (k *Secret) Get() ([]byte, error) {
	dk := Derive(k.Key, k.Salt)

	dec, err := Decrypt(k.EncSecret, dk)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt seed: %w", err)
	}
	return dec, nil
}

func (k *Secret) Cleanup() {
	if k != nil {
		Zero(k.Key)
		Zero(k.EncSecret)
		Zero(k.Salt)
	}
}

// super mega secret obfuscation ;-)
func Derive(key, salt []byte) []byte {
	return pbkdf2.Key(key, salt, cost(), 32, sha256.New)
}

func cost() int {
	return int(math.Sqrt(25)) - 2
}
