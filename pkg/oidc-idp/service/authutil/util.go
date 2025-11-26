// Package authutil contains utilities to set up example authorization server
// using goidc.
package authutil

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/oidc-idp/service/ui"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"go.uber.org/zap"
)

const (
	Issuer string = "http://localhost:5000/openid-connect"
	// MTLSHost                 string = "https://matls-auth.localhost"
	// headerClientCert         string = "X-Client-Cert"
	// headerXFAPIInteractionID string = "X-Fapi-Interaction-Id"
)

var (
	Scopes = []goidc.Scope{
		goidc.ScopeOpenID, goidc.ScopeOfflineAccess, goidc.ScopeProfile,
		goidc.ScopeEmail, goidc.ScopeAddress, goidc.ScopePhone,
	}
	Claims = []string{
		goidc.ClaimEmail, goidc.ClaimEmailVerified, goidc.ClaimPhoneNumber,
		goidc.ClaimPhoneNumberVerified, goidc.ClaimAddress,
	}
	ACRs          = []goidc.ACR{goidc.ACRMaceIncommonIAPBronze, goidc.ACRMaceIncommonIAPSilver}
	DisplayValues = []goidc.DisplayValue{goidc.DisplayValuePage, goidc.DisplayValuePopUp}
)

var (
	errLogoutCancelled error = errors.New("logout cancelled by the user")
)

func PrivateJWKSFunc(cfg config.Config) goidc.JWKSFunc {
	jwks := goidc.JSONWebKeySet{
		Keys: []goidc.JSONWebKey{
			{
				Algorithm: "HS256",
				KeyID:     "chorus-backend-key",
				Key:       []byte(cfg.Daemon.JWT.Secret.PlainText()),
			},
		},
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

func RenderError() goidc.RenderErrorFunc {
	tmpl := template.Must(template.ParseFS(ui.FS, "*.html"))
	return func(w http.ResponseWriter, r *http.Request, err error) error {
		w.WriteHeader(http.StatusOK)
		_ = tmpl.ExecuteTemplate(w, "error.html", authnPage{
			Error: err.Error(),
		})
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

type LogoutPage struct {
	BaseURL     string
	CallbackID  string
	IsLoggedOut bool
	Session     map[string]any
}

func LogoutPolicy() goidc.LogoutPolicy {

	tmpl := template.Must(template.ParseFS(ui.FS, "logout.html"))
	return goidc.NewLogoutPolicy(
		"main",
		func(r *http.Request, ls *goidc.LogoutSession) bool {
			return true
		},
		func(w http.ResponseWriter, r *http.Request, ls *goidc.LogoutSession) (goidc.Status, error) {
			logout := r.PostFormValue("logout")
			if logout == "" {
				logger.TechLog.Debug(context.Background(), "rendering logout page", zap.String("callback_id", ls.CallbackID))
				sess, err := mapify(ls)
				if err != nil {
					logger.TechLog.Error(context.Background(), "unable to mapify logout session for rendering", zap.Error(err))
					return goidc.StatusFailure, fmt.Errorf("unable to render logout page: %w", err)
				}
				if err := tmpl.ExecuteTemplate(w, "logout.html", LogoutPage{
					BaseURL:    Issuer,
					CallbackID: ls.CallbackID,
					Session:    sess,
				}); err != nil {
					return goidc.StatusFailure, err
				}
				return goidc.StatusInProgress, nil
			}

			if !isTrue(logout) {
				logger.TechLog.Debug(context.Background(), "user cancelled logout", zap.String("logout", logout))
				return goidc.StatusFailure, errLogoutCancelled
			}

			cookie, err := r.Cookie(cookieUserSessionID)
			if err != nil {
				logger.TechLog.Debug(context.Background(), "the session cookie was not found", zap.Error(err))
				return goidc.StatusSuccess, nil
			}

			delete(userSessionStore, cookie.Value)
			return goidc.StatusSuccess, nil
		},
	)
}

func HandleLogout() goidc.HandleDefaultPostLogoutFunc {
	tmpl := template.Must(template.ParseFS(ui.FS, "logout.html"))
	return func(w http.ResponseWriter, r *http.Request, ls *goidc.LogoutSession) error {
		if err := tmpl.ExecuteTemplate(w, "logout.html", LogoutPage{IsLoggedOut: true}); err != nil {
			return fmt.Errorf("could not execute logout template: %w", err)
		}
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
