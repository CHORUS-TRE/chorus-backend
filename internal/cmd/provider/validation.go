package provider

import (
	"reflect"
	"strings"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/validation"

	val "github.com/go-playground/validator/v10"
)

var validatorOnce sync.Once
var validator *val.Validate

func ProvideValidator() *val.Validate {
	validatorOnce.Do(func() {
		validator = validation.NewValidator()

		// Report the yaml tag name instead of the Go field name
		// (e.g. "daemon.grpc.host" instead of "Daemon.GRPC.Host")
		// Falls back to the Go field name for structs with no yaml tag.
		validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		validator.RegisterStructValidation(validateOpenIDMode, config.Mode{})
	})
	return validator
}

// validateOpenIDMode requires the OpenID fields an "openid"-type mode actually
// depends on.
// These can't be expressed as required_if tags on OpenID's own fields
func validateOpenIDMode(sl val.StructLevel) {
	mode := sl.Current().Interface().(config.Mode)
	if mode.Type != "openid" || !mode.Enabled {
		return
	}

	o := mode.OpenID
	reportIfEmpty := func(value, field, structField string) {
		if value == "" {
			sl.ReportError(value, field, structField, "required_openid", "")
		}
	}

	reportIfEmpty(o.ID, "openid.id", "OpenID.ID")
	reportIfEmpty(o.ClientID, "openid.client_id", "OpenID.ClientID")
	reportIfEmpty(o.ClientSecret.PlainText(), "openid.client_secret", "OpenID.ClientSecret")
	reportIfEmpty(o.AuthorizeURL, "openid.authorize_url", "OpenID.AuthorizeURL")
	reportIfEmpty(o.TokenURL, "openid.token_url", "OpenID.TokenURL")
	reportIfEmpty(o.UserInfoURL, "openid.user_info_url", "OpenID.UserInfoURL")

	if o.EnableFrontendRedirect {
		reportIfEmpty(o.ChorusFrontendRedirectURL, "openid.chorus_frontend_redirect_url", "OpenID.ChorusFrontendRedirectURL")
	} else {
		reportIfEmpty(o.ChorusBackendHost, "openid.chorus_backend_host", "OpenID.ChorusBackendHost")
	}
}
