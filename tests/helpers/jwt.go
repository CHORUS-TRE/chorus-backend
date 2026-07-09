//go:build unit || integration || acceptance
// +build unit integration acceptance

package helpers

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"time"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"

	jwt_go "github.com/golang-jwt/jwt"
)

func CreateJWTToken(id, tenantId uint64, roleName string, roleContext map[string]string) string {
	// TODO: discuss adding function in the auth package to create a token with roles, so we don't have to do this here
	roles := []jwt_model.Role{{Name: roleName, Context: roleContext}}

	rolesJSON, err := json.Marshal(roles)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(rolesJSON); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	r := base64.StdEncoding.EncodeToString(buf.Bytes())

	claims := &jwt_model.JWTClaims{
		ID:        id,
		TenantID:  tenantId,
		FirstName: "hello",
		LastName:  "moto",
		R:         r,
		Username:  "hmoto",
		StandardClaims: jwt_go.StandardClaims{
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Unix(),
			IssuedAt:  jwt_go.TimeFunc().Unix(),
		},
	}
	obj := jwt_go.NewWithClaims(jwt_go.SigningMethodHS256, claims)
	token, err := obj.SignedString([]byte(Conf().Daemon.JWT.Secret.PlainText()))
	if err != nil {
		panic(err)
	}
	return token
}
