package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"math"

	"golang.org/x/crypto/pbkdf2"

	"github.com/pkg/errors"
)

type Secret struct {
	EncSecret []byte
	Key       []byte
	Salt      []byte
}

func NewSecret(secret []byte) (*Secret, error) {

	key := make([]byte, 32)
	if _, err := rand.Reader.Read(key); err != nil {
		return nil, errors.Wrap(err, "unable to generate Key: random reader failed")
	}

	salt := make([]byte, 32)
	if _, err := rand.Reader.Read(salt); err != nil {
		return nil, errors.Wrap(err, "unable to generate Salt: random reader failed")
	}
	dk := Derive(key, salt)
	enc, err := Encrypt(secret, dk)
	Zero(secret)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to encrypt secret")
	}

	return &Secret{EncSecret: enc, Key: key, Salt: salt}, nil
}

func (k *Secret) Get() ([]byte, error) {
	dk := Derive(k.Key, k.Salt)

	dec, err := Decrypt(k.EncSecret, dk)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to decrypt seed")
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
