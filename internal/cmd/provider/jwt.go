package provider

import (
	"sync"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	jwt_go "github.com/golang-jwt/jwt"
)

var claimsFactoryOnce sync.Once
var claimsFactory jwt_model.ClaimsFactory

// ProvideClaimsFactory returns a fresh JWT-claims-factory instance.
func ProvideClaimsFactory() jwt_model.ClaimsFactory {
	claimsFactoryOnce.Do(func() {
		claimsFactory = jwt_model.NewJWTClaimsFactory(logger.SecLog)
	})
	return claimsFactory
}

var keyFuncOnce sync.Once
var keyFunc jwt_go.Keyfunc

// ProvideKeyFunc returns a JWT-key-function that simply returns
// the signing-key upon invocation.
func ProvideKeyFunc(signingKey string) jwt_go.Keyfunc {
	keyFuncOnce.Do(func() {
		keyFunc = func(token *jwt_go.Token) (interface{}, error) {
			return []byte(signingKey), nil
		}
	})
	return keyFunc
}
