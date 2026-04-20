package crypto

import "fmt"

func EncryptField(plaintext string, daemonKey *Secret) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if daemonKey == nil {
		return "", fmt.Errorf("daemon encryption key is nil")
	}
	key, err := daemonKey.Get()
	if err != nil {
		return "", fmt.Errorf("unable to get encryption key: %w", err)
	}
	defer Zero(key)

	return EncryptToString([]byte(plaintext), key)
}

func DecryptField(ciphertext string, daemonKey *Secret) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	if daemonKey == nil {
		return "", fmt.Errorf("daemon encryption key is nil")
	}
	key, err := daemonKey.Get()
	if err != nil {
		return "", fmt.Errorf("unable to get encryption key: %w", err)
	}
	defer Zero(key)

	decrypted, err := DecryptFromString(ciphertext, key)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
