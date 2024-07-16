package crypto

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeriveEncryptDecrypt(t *testing.T) {

	key := []byte("Key")
	salt := []byte("salt_salt")
	plain := "toto_toto_toto_toto_toto_toto_toto"

	k := Derive(key, salt)
	fmt.Printf("Key: %v\n", k)

	enc, err := Encrypt([]byte(plain), k)
	assert.Nil(t, err)

	fmt.Printf("enc: %v\n", enc)

	dec, err := Decrypt(enc, k)
	assert.Nil(t, err)
	fmt.Printf("dec: %v\n", string(dec))

	assert.Equal(t, plain, string(dec))

}

func TestNewSecret(t *testing.T) {
	start := time.Now()
	s, err1 := NewSecret([]byte("my secret"))
	sec, err2 := s.Get()
	fmt.Printf("took: %v ns\n", time.Since(start).Nanoseconds())
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	fmt.Printf("res: %v\n", string(sec))
}
