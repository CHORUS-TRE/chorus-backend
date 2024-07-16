package helper

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/trustelem/zxcvbn"
	"golang.org/x/crypto/bcrypt"
)

const lowerCase = "abcdefghijklmnopqrstuvwxyz"
const upperCase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numbers = "0123456789"
const SpecialChars = `%&'"#?!@$%^&*-.+/()= _^`

func IsStrongPassword(pwd string) bool {
	switch {
	case len(pwd) < 14,
		!strings.ContainsAny(pwd, lowerCase),
		!strings.ContainsAny(pwd, upperCase),
		!strings.ContainsAny(pwd, numbers),
		!strings.ContainsAny(pwd, SpecialChars),
		zxcvbn.PasswordStrength(pwd, nil).Score < 3:
		return false
	}
	return true
}

func HashPass(pass string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "unable to hash password")
	}
	return string(hash), nil
}

//nolint:gosimple
func CheckPassHash(hash, pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	if err != nil {
		return false
	}
	return true
}

func GeneratePassword(length int) (string, error) {
	var dictionary = lowerCase + upperCase + numbers + SpecialChars
	for {
		var bytes = make([]byte, length)
		n, err := rand.Read(bytes)
		if err != nil {
			return "", err
		}
		if n != length {
			return "", fmt.Errorf("got only %v of %v random bytes", n, length)
		}
		for k, v := range bytes {
			bytes[k] = dictionary[v%byte(len(dictionary))]
		}
		if IsStrongPassword(string(bytes)) {
			return string(bytes), nil
		}
	}
}
