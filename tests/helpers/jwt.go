//go:build unit || integration || acceptance
// +build unit integration acceptance

package helpers

import (
	"time"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"

	jwt_go "github.com/golang-jwt/jwt"
)

func CreateJWTToken(id, tenantId uint64, role string) string {
	claims := &jwt_model.JWTClaims{
		ID:        id,
		TenantID:  tenantId,
		FirstName: "hello",
		LastName:  "moto",
		Roles:     []string{role},
		Username:  "hmoto",
		StandardClaims: jwt_go.StandardClaims{
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Unix(),
			IssuedAt:  jwt_go.TimeFunc().Unix(),
		},
	}
	obj := jwt_go.NewWithClaims(jwt_go.SigningMethodHS256, claims)
	token, err := obj.SignedString([]byte(Conf().Daemon.JWT.Secret))
	if err != nil {
		panic(err)
	}
	return token
}
