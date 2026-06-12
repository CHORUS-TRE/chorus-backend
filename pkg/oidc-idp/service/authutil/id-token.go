package authutil

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/luikyv/go-oidc/pkg/goidc"
)

// idTokenClaims mirrors the shape of an OIDC ID token, with the most
// commonly used profile claims. Anything left empty is omitted from the
// serialized JWT so consumers see only the fields we explicitly populate.
type idTokenClaims struct {
	jwt.Claims

	AuthTime          int64  `json:"auth_time,omitempty"`
	AZP               string `json:"azp,omitempty"`
	Email             string `json:"email,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Name              string `json:"name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
}

// IssueIDToken signs an OIDC ID token for the given client and user using
// the OIDC provider's private JWKS. The signing algorithm is taken from the
// JWK itself (its `alg` field), so the token always matches whatever the
// provider publishes at its JWKS endpoint — no algorithm is hard-coded here.
// The resulting token has the same shape as one produced by the
// authorization-code flow, so downstream consumers can verify it against the
// provider's published JWKS at `<issuer>/openid-connect/.well-known/jwks.json`.
//
// We can't reuse go-oidc's internal token.MakeIDToken: it lives under
// internal/token, takes an internal/oidc.Context, and there is no public
// Provider.MakeIDToken wrapper. The closest public API, Provider.MakeToken,
// only mints access tokens. So we drive go-jose directly, but pick the alg
// from the JWK to stay aligned with whatever the operator configured.
func IssueIDToken(ctx context.Context, cfg config.Config, clientID string, user *user_model.User, ttl time.Duration) (string, error) {
	if user == nil {
		return "", fmt.Errorf("user must not be nil")
	}
	if clientID == "" {
		return "", fmt.Errorf("clientID must not be empty")
	}

	jwks, err := PrivateJWKSFunc(cfg)(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to load private JWKS: %w", err)
	}

	signingKey, err := pickSigningKey(jwks)
	if err != nil {
		return "", err
	}

	alg := jose.SignatureAlgorithm(signingKey.Algorithm)
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: alg, Key: signingKey},
		(&jose.SignerOptions{}).
			WithType("JWT").
			WithHeader(jose.HeaderKey("kid"), signingKey.KeyID),
	)
	if err != nil {
		return "", fmt.Errorf("unable to build JWT signer (alg=%s): %w", alg, err)
	}

	now := time.Now().UTC()
	claims := idTokenClaims{
		Claims: jwt.Claims{
			Issuer:    cfg.Services.OpenIDConnectProvider.IssuerURL,
			Subject:   strconv.FormatUint(user.ID, 10),
			Audience:  jwt.Audience{clientID},
			Expiry:    jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		AuthTime:          now.Unix(),
		AZP:               clientID,
		Email:             user.Email,
		PreferredUsername: user.Username,
		Name:              joinName(user.FirstName, user.LastName),
		GivenName:         user.FirstName,
		FamilyName:        user.LastName,
	}

	token, err := jwt.Signed(signer).Claims(claims).Serialize()
	if err != nil {
		return "", fmt.Errorf("unable to sign ID token: %w", err)
	}

	return token, nil
}

// pickSigningKey returns the first JWK in jwks usable as a signing key: it
// must be flagged for signing (Use == "sig" or unset), declare a concrete
// algorithm (not "none"), and hold private key material. This mirrors how
// go-oidc itself selects the signing key, so the algorithm follows whatever
// the operator put in the JWKS without us having to hard-code RS256/ES256/…
func pickSigningKey(jwks goidc.JSONWebKeySet) (goidc.JSONWebKey, error) {
	for _, key := range jwks.Keys {
		if key.Use != "" && key.Use != "sig" {
			continue
		}
		if key.Algorithm == "" || key.Algorithm == string(goidc.None) {
			continue
		}
		if !key.Valid() || key.IsPublic() {
			continue
		}
		return key, nil
	}
	return goidc.JSONWebKey{}, fmt.Errorf("no usable private signing JWK found in private JWKS")
}

// joinName joins first and last name with a single space, skipping empty
// segments so we don't produce a stray leading/trailing space.
func joinName(first, last string) string {
	switch {
	case first == "" && last == "":
		return ""
	case first == "":
		return last
	case last == "":
		return first
	default:
		return first + " " + last
	}
}
