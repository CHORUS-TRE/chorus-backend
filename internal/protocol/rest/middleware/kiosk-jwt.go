package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	embed "github.com/CHORUS-TRE/chorus-backend/api"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	jwt_helper "github.com/CHORUS-TRE/chorus-backend/internal/jwt/helper"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	jwt_go "github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

// KioskJWTPath is the URL path served for the kiosk JWT-to-cookie exchange.
//
// A kiosk app pod is started with KIOSK_JWT_URL pointing here. The pod's
// headless Chrome navigates to "<KIOSK_JWT_URL>#jwt=<short-lived-jwt>".
// The embedded HTML page reads the JWT from the URL fragment and POSTs it
// back to this same path, where it is validated and exchanged for the
// regular `jwttoken` HttpOnly session cookie. The kiosk's main browser then
// reuses that cookie (same Chrome user-data-dir) so that any OIDC sign-in
// flow against this backend completes silently for the pre-logged user.
const KioskJWTPath = "/kiosk-jwt"

// AddKioskJWT mounts the kiosk JWT exchange handlers on KioskJWTPath.
//
//   - GET  /kiosk-jwt -> serves the embedded HTML/JS page that extracts the
//     JWT from the URL fragment and POSTs it to this same path.
//   - POST /kiosk-jwt -> validates the JWT and, on success, sets the
//     `jwttoken` HttpOnly session cookie scoped to the configured cookie
//     domain. Returns 204 on success.
func AddKioskJWT(h http.Handler, cfg config.Config, keyFunc jwt_go.Keyfunc, claimsFactory jwt_model.ClaimsFactory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != KioskJWTPath {
			h.ServeHTTP(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			serveKioskJWTPage(w, r)
		case http.MethodPost:
			handleKioskJWTExchange(w, r, cfg, keyFunc, claimsFactory)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func serveKioskJWTPage(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	page, err := embed.KioskJWTEmbed.ReadFile("kiosk-jwt/index.html")
	if err != nil {
		logger.TechLog.Error(r.Context(), "unable to read embedded kiosk-jwt page", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	// The page must never be cached: it consumes a single-use short-lived JWT.
	h.Set("Cache-Control", "no-store")
	h.Set("X-Content-Type-Options", "nosniff")
	// Restrict the page so only its own inline script runs and it can only
	// fetch the same path it was served from.
	h.Set("Content-Security-Policy", "default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; connect-src 'self'")
	h.Set("Referrer-Policy", "no-referrer")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(page)
}

func logRequest(r *http.Request) {
	logger.SecLog.Info(r.Context(), "kiosk-jwt: request received",
		zap.String("method", r.Method),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
	)
}

type kioskJWTRequest struct {
	Token string `json:"token"`
}

func handleKioskJWTExchange(w http.ResponseWriter, r *http.Request, cfg config.Config, keyFunc jwt_go.Keyfunc, claimsFactory jwt_model.ClaimsFactory) {
	ctx := r.Context()

	// Limit body size; a JWT is small.
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	defer r.Body.Close()

	var req kioskJWTRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	token := strings.TrimSpace(req.Token)
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	claims, err := jwt_helper.ParseToken(ctx, token, keyFunc, claimsFactory)
	if err != nil {
		logger.SecLog.Warn(ctx, "kiosk-jwt: invalid token", zap.Error(err))
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	c, ok := claims.(*jwt_model.JWTClaims)
	if !ok {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// The cookie is set with the same attributes used by the regular login
	// flow (see internal/api/v1/authentication-controller.go), so that the
	// OIDC provider's middleware picks it up exactly the same way.
	expiresUnix := c.StandardClaims.ExpiresAt
	if expiresUnix <= 0 {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	expires := time.Unix(expiresUnix, 0).UTC()

	cookie := &http.Cookie{
		Name:     "jwttoken",
		Value:    token,
		Path:     "/",
		Domain:   cfg.Daemon.HTTP.Headers.CookieDomain,
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, cookie)

	logger.SecLog.Info(ctx, "kiosk-jwt: session cookie set",
		zap.Uint64("user_id", c.ID),
		zap.Uint64("tenant_id", c.TenantID),
		zap.String("for_client", c.ForClient),
	)

	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}
