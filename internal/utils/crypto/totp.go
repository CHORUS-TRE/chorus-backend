package crypto

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
)

const (
	totpIssuerName         = "chorus"
	totpRecoveryCodeLength = 5
)

// CreateTotpSecret generates a new time-based one-time password (compatible with
// google-authenticator) for the given user with the following properties:
//
//	validity period: 30 seconds
//	secret size: 20 bytes
//	algorithm: HMAC-SHA1.
func CreateTotpSecret(username string, daemonEncryptionKey *Secret) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuerName,
		AccountName: username,
	})
	if err != nil {
		return "", err
	}

	encryptionKey, err := daemonEncryptionKey.Get()
	if err != nil {
		return "", err
	}
	defer Zero(encryptionKey)

	return EncryptToString([]byte(key.Secret()), encryptionKey)
}

func DecryptTotpSecret(encryptedSecret string, daemonEncryptionKey *Secret) (string, error) {
	encryptionKey, err := daemonEncryptionKey.Get()
	if err != nil {
		return "", err
	}
	defer Zero(encryptionKey)

	decryptedSecret, err := DecryptFromString(encryptedSecret, encryptionKey)
	if err != nil {
		return "", err
	}

	return string(decryptedSecret), nil
}

// CreateTotpRecovery generates and returns num TOTP recovery code strings.
func CreateTotpRecoveryCodes(num int, daemonEncryptionKey *Secret) ([]string, error) {
	encryptionKey, err := daemonEncryptionKey.Get()
	if err != nil {
		return nil, err
	}
	defer Zero(encryptionKey)

	codes := make([]string, num)
	for i := 0; i < num; i++ {
		b := make([]byte, totpRecoveryCodeLength)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		bString := hex.EncodeToString(b)
		encCode, err := EncryptToString([]byte(bString), encryptionKey)
		if err != nil {
			return nil, err
		}
		codes[i] = encCode
	}
	return codes, nil
}

// DecryptTotpRecoveryCodes decrypts a given array defined with encrypted recovery codes.
func DecryptTotpRecoveryCodes(encRecoveryCodes []string, daemonEncryptionKey *Secret) ([]string, error) {

	encryptionKey, err := daemonEncryptionKey.Get()
	if err != nil {
		return nil, err
	}
	defer Zero(encryptionKey)

	decryptedRecoveryCodes := []string{}

	for _, c := range encRecoveryCodes {
		code, err := DecryptFromString(c, encryptionKey)
		if err != nil {
			return nil, err
		}
		decryptedRecoveryCodes = append(decryptedRecoveryCodes, string(code))
	}

	return decryptedRecoveryCodes, nil
}

// VerifyTotp checks whether the provided totpCode is a valid TOTP.
func VerifyTotp(totpCode, encSecret string, daemonEncryptionKey *Secret) (bool, error) {
	encryptionKey, err := daemonEncryptionKey.Get()
	if err != nil {
		return false, err
	}
	defer Zero(encryptionKey)

	secret, err := DecryptFromString(encSecret, encryptionKey)
	if err != nil {
		return false, nil
	}
	return totp.Validate(totpCode, string(secret)), nil
}

// VerifyTotpRecoveryCode checks whether a provided TOTP recovery code in a list of codes
// If there is a match the respective recovery code is returned.
func VerifyTotpRecoveryCode(ctx context.Context, totpRecoveryCode string, encCodes []*model.TotpRecoveryCode, daemonEncryptionKey *Secret) (*model.TotpRecoveryCode, error) {
	encryptionKey, err := daemonEncryptionKey.Get()
	if err != nil {
		return nil, err
	}
	defer Zero(encryptionKey)

	for _, c := range encCodes {

		code, err := DecryptFromString(c.Code, encryptionKey)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to decrypt recovery code", zap.Error(err))
			return nil, err
		}
		if totpRecoveryCode == string(code) {
			return c, nil
		}
	}
	return nil, nil
}
