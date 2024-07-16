package crypto

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/stretchr/testify/require"
)

func TestCreateTotpSecret(t *testing.T) {

	daemonEncryptionKey, err := NewSecret(Derive([]byte("dummysecret"), []byte("salt")))
	require.NoError(t, err)

	s, err := CreateTotpSecret("toto", daemonEncryptionKey)
	require.Nil(t, err)
	fmt.Printf("encrypted secret: %v\n", s)

	dec, err := DecryptTotpSecret(s, daemonEncryptionKey)
	require.Nil(t, err)
	fmt.Printf("decrypted secret: %v\n", string(dec))
}

func TestCreateTotpRecoveryCodes(t *testing.T) {

	daemonEncryptionKey, err := NewSecret(Derive([]byte("dummysecret"), []byte("salt")))
	require.NoError(t, err)

	encryptionKey, err := daemonEncryptionKey.Get()
	require.NoError(t, err)

	codes, err := CreateTotpRecoveryCodes(5, daemonEncryptionKey)
	require.Nil(t, err)
	require.Equal(t, 5, len(codes))
	fmt.Printf("codes: %v \n", codes)

	for _, c := range codes {
		dec, err := DecryptFromString(c, encryptionKey)
		require.Nil(t, err)
		fmt.Printf("code: %v \n", base64.StdEncoding.EncodeToString(dec))
	}
}

func TestDecryptTotpRecoveryCodes(t *testing.T) {

	daemonEncryptionKey, err := NewSecret(Derive([]byte("dummysecret"), []byte("salt")))
	require.NoError(t, err)

	encryptionKey, err := daemonEncryptionKey.Get()
	require.NoError(t, err)

	codes, err := CreateTotpRecoveryCodes(5, daemonEncryptionKey)
	require.Nil(t, err)
	require.Equal(t, 5, len(codes))
	fmt.Printf("codes: %v \n", codes)

	decCodes, err := DecryptTotpRecoveryCodes(codes, daemonEncryptionKey)
	require.Nil(t, err)
	require.Equal(t, 5, len(decCodes))

	for _, c := range codes {
		dec, err := DecryptFromString(c, encryptionKey)
		require.Nil(t, err)
		require.Contains(t, decCodes, string(dec))
	}

}

func TestVerifyTotpRecoveryCodes(t *testing.T) {

	daemonEncryptionKey, err := NewSecret(Derive([]byte("dummysecret"), []byte("salt")))
	require.NoError(t, err)

	encryptionKey, err := daemonEncryptionKey.Get()
	require.NoError(t, err)

	codes, err := CreateTotpRecoveryCodes(5, daemonEncryptionKey)
	require.Nil(t, err)
	require.Equal(t, 5, len(codes))
	fmt.Printf("codes: %v \n", codes)

	wrappedCodes := []*model.TotpRecoveryCode{}
	id := uint64(0)
	for _, c := range codes {
		tempCode := model.TotpRecoveryCode{
			ID:       id,
			TenantID: 88888,
			UserID:   100,
			Code:     c,
		}
		wrappedCodes = append(wrappedCodes, &tempCode)
		id += 1
	}

	res, err := VerifyTotpRecoveryCode(context.Background(), "fakecode", wrappedCodes, daemonEncryptionKey)
	require.NoError(t, err)
	require.Nil(t, res)

	totpRecoveryCode, err := DecryptFromString(codes[0], encryptionKey)
	require.Nil(t, err)
	res, err = VerifyTotpRecoveryCode(context.Background(), string(totpRecoveryCode), wrappedCodes, daemonEncryptionKey)
	require.NoError(t, err)
	require.Equal(t, codes[0], res.Code)

}

func TestEncryptDecryptTotpData(t *testing.T) {
	data := "data"

	daemonEncryptionKey, err := NewSecret(Derive([]byte("dummysecret"), []byte("salt")))
	require.NoError(t, err)

	encryptionKey, err := daemonEncryptionKey.Get()
	require.NoError(t, err)

	enc, err := EncryptToString([]byte(data), encryptionKey)
	require.Nil(t, err)
	fmt.Printf("enc: %v\n", enc)

	dec, err := DecryptFromString(enc, encryptionKey)
	require.Nil(t, err)
	fmt.Printf("dec: %v\n", string(dec))
	require.Equal(t, data, string(dec))
}
