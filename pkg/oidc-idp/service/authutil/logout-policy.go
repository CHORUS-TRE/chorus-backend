package authutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"go.uber.org/zap"
)

var (
	errLogoutCancelled error = errors.New("logout cancelled by the user")
)

type LogoutPage struct {
	BaseURL     string
	CallbackID  string
	IsLoggedOut bool
	Session     map[string]any
}

func LogoutPolicy(cfg config.Config) goidc.LogoutPolicy {
	return goidc.NewLogoutPolicy(
		"main",
		func(r *http.Request, ls *goidc.LogoutSession) bool {
			return true
		},
		func(w http.ResponseWriter, r *http.Request, ls *goidc.LogoutSession) (goidc.Status, error) {
			logout := r.URL.Query().Get("logout")

			if logout == "" {
				logger.TechLog.Debug(context.Background(), "rendering logout page", zap.String("callback_id", ls.CallbackID))

				callbackURL, err := callbackURL(cfg, "logout", ls.CallbackID, "logout", nil)
				if err != nil {
					return goidc.StatusFailure, fmt.Errorf("unable to build callback URL: %w", err)
				}

				err = redirect(cfg, w, "logout", callbackURL, nil)
				if err != nil {
					return goidc.StatusFailure, fmt.Errorf("unable to redirect to logout page: %w", err)
				}

				return goidc.StatusInProgress, nil
			}

			if !isTrue(logout) {
				logger.TechLog.Debug(context.Background(), "user cancelled logout", zap.String("logout", logout))
				v := url.Values{}
				v.Set("error", errLogoutCancelled.Error())
				err := redirect(cfg, w, "error", "", v)
				if err != nil {
					return goidc.StatusFailure, fmt.Errorf("unable to redirect to error page: %w", err)
				}

				return goidc.StatusFailure, errLogoutCancelled
			}

			cookie, err := r.Cookie(cookieUserSessionID)
			if err != nil {
				logger.TechLog.Debug(context.Background(), "the session cookie was not found", zap.Error(err))
				return goidc.StatusSuccess, nil
			}

			delete(userSessionStore, cookie.Value)

			// r.Header.Set("Set-Cookie", "jwttoken=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax")
			logger.TechLog.Debug(context.Background(), "removing jwt token cookie upon logout")
			r.AddCookie(&http.Cookie{
				Name:  "jwttoken",
				Value: "",

				Path:     "/",
				MaxAge:   0,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})

			return goidc.StatusSuccess, nil
		},
	)
}

func HandleLogout(cfg config.Config) goidc.HandleDefaultPostLogoutFunc {
	return func(w http.ResponseWriter, r *http.Request, ls *goidc.LogoutSession) error {
		v := url.Values{}
		v.Set("is_logged_out", "true")
		err := redirect(cfg, w, "logout", "", v)
		if err != nil {
			return fmt.Errorf("unable to redirect to logout page: %w", err)
		}

		return nil
	}
}
