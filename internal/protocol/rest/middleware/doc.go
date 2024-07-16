package middleware

import (
	"io/fs"
	"net/http"
	"strings"

	embed "github.com/CHORUS-TRE/chorus-backend/api" // For static assets.
)

func AddDoc(h http.Handler) http.Handler {
	apiFS := http.FS(embed.APIEmbed)
	subFS, _ := fs.Sub(embed.UIEmbed, "openapiv2/ui")
	uiFS := http.FS(subFS)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if strings.HasPrefix(r.RequestURI, "/openapi") {
			handler := http.StripPrefix("/openapi", http.FileServer(apiFS))
			handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if strings.HasPrefix(r.RequestURI, "/doc") {
			if r.RequestURI == "/doc" {
				handler := http.StripPrefix("/doc", http.RedirectHandler("/doc/", http.StatusFound))
				handler.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			handler := http.StripPrefix("/doc", http.FileServer(uiFS))
			handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// If not "/doc/", passing to the next middleware
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
