package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/pkg/errors"
)

func EncryptToString(plaintext []byte, key []byte) (string, error) {
	e, err := Encrypt(plaintext, key)
	if err != nil {
		return "", errors.Wrapf(err, "unable to encrypt plaintext")
	}
	return base64.StdEncoding.EncodeToString(e), nil
}

func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func DecryptFromString(ciphertext string, key []byte) ([]byte, error) {
	c, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to decode base64 encoded ciphertext")
	}
	return Decrypt(c, key)
}

func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize+gcm.Overhead() {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func Zero(b []byte) {
	if b == nil {
		return
	}
	lenb := len(b)
	for i := 0; i < lenb; i++ {
		b[i] = 0
	}
}
