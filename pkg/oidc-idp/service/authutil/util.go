// Package authutil contains utilities to set up example authorization server
// using goidc.
package authutil

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"go.uber.org/zap"
)

var (
	Claims = []string{
		goidc.ClaimEmail,
	}
	ACRs          = []goidc.ACR{goidc.ACRMaceIncommonIAPBronze, goidc.ACRMaceIncommonIAPSilver}
	DisplayValues = []goidc.DisplayValue{goidc.DisplayValuePage, goidc.DisplayValuePopUp}
)

func callbackURL(cfg config.Config, policyBase, callbackID, authorizePage string, v url.Values) (string, error) {
	u, err := url.Parse(cfg.Services.OpenIDConnectProvider.IssuerURL)
	if err != nil {
		return "", fmt.Errorf("unable to parse request URI: %w", err)
	}
	u.Path = path.Join(u.Path, policyBase, callbackID, authorizePage)
	if v != nil {
		u.RawQuery = v.Encode()
	}
	return u.String(), nil
}

func redirect(cfg config.Config, w http.ResponseWriter, interactionPage, callbackURL string, v url.Values) error {
	cu, err := url.Parse(cfg.Services.OpenIDConnectProvider.FrontendInteractionsURL)
	if err != nil {
		return fmt.Errorf("unable to parse interaction base URL: %w", err)
	}

	if v == nil {
		v = url.Values{}
	}
	v.Set("callback_url", callbackURL)
	v.Set("page", interactionPage)
	cu.RawQuery = v.Encode()

	w.Header().Set("Location", cu.String())
	w.WriteHeader(http.StatusFound)

	return nil
}

func PrivateJWKSFunc(cfg config.Config) goidc.JWKSFunc {
	var jwks goidc.JSONWebKeySet
	err := json.Unmarshal([]byte(cfg.Services.OpenIDConnectProvider.JWKS.PlainText()), &jwks)
	if err != nil {
		logger.TechLog.Fatal(context.Background(), "unable to unmarshal private test server key", zap.Error(err))
	}

	return func(ctx context.Context) (goidc.JSONWebKeySet, error) {
		return jwks, nil
	}
}

func TokenOptionsFunc(alg goidc.SignatureAlgorithm) goidc.TokenOptionsFunc {
	return func(grantInfo goidc.GrantInfo, _ *goidc.Client) goidc.TokenOptions {
		opts := goidc.NewJWTTokenOptions(alg, 600)
		return opts
	}
}

func IssueRefreshToken(client *goidc.Client, grantInfo goidc.GrantInfo) bool {
	return slices.Contains(client.GrantTypes, goidc.GrantRefreshToken)
}

func HTTPClient(_ context.Context) *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}
}

func ErrorLoggingFunc(ctx context.Context, err error) {
	logger.TechLog.Error(ctx, "OIDC provider error", zap.Error(err))
}

func RenderError(cfg config.Config) goidc.RenderErrorFunc {
	return func(w http.ResponseWriter, r *http.Request, errToDisplay error) error {
		v := url.Values{}
		v.Set("error", errToDisplay.Error())

		err := redirect(cfg, w, "error", "", v)
		if err != nil {
			return fmt.Errorf("unable to redirect to error page: %w", err)
		}

		return nil
	}
}

func CheckJTIFunc() goidc.CheckJTIFunc {
	jtiStore := make(map[string]struct{})
	return func(ctx context.Context, jti string) error {
		if _, ok := jtiStore[jti]; ok {
			return errors.New("jti already used")
		}

		jtiStore[jti] = struct{}{}
		return nil
	}
}

func isTrue(s string) bool {
	return s == "true"
}

func TimestampNow() int {
	return int(Now().Unix())
}

func Now() time.Time {
	return time.Now().UTC()
}
