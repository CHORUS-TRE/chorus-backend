package validation

import (
	"context"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	val "github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

const (
	generalString = `^[0-9a-zA-ZA-Za-zÀ-ÿ\s,;.:\-_\$£!^'?=()/&%*#\\"\'\+\[\]\{\}@]*$`
	safeString    = `^[0-9a-zA-ZA-Za-zÀ-ÿ\s\-_@.]*$`
)

var (
	generalStringRegexp *regexp.Regexp
	safeStringRegexp    *regexp.Regexp
)

func NewValidator() *val.Validate {
	v := val.New()

	var err error
	ctx := context.Background()

	generalStringRegexp, err = regexp.Compile(generalString)
	if err != nil {
		logger.TechLog.Fatal(ctx, "unable to create regexp", zap.Error(err))
	}

	err = v.RegisterValidation("generalstring", generalstring)
	if err != nil {
		logger.TechLog.Fatal(ctx, "unable to register validator 'generalstring'", zap.Error(err))
	}

	safeStringRegexp, err = regexp.Compile(safeString)
	if err != nil {
		logger.TechLog.Fatal(ctx, "unable to create regexp", zap.Error(err))
	}

	err = v.RegisterValidation("safestring", safestring)
	if err != nil {
		logger.TechLog.Fatal(ctx, "unable to register validator 'safestring'", zap.Error(err))
	}

	err = v.RegisterValidation("ltetomorrowutc", ltetomorrowutc)
	if err != nil {
		logger.TechLog.Fatal(context.Background(), "unable to register validator 'ltetomorrowutc'", zap.Error(err))
	}

	// Report the yaml tag name instead of the Go field name
	// (e.g. "daemon.grpc.host" instead of "Daemon.GRPC.Host")
	// Falls back to the Go field name for structs with no yaml tag.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	v.RegisterStructValidation(validateOpenIDMode, config.Mode{})
	v.RegisterStructValidation(validateHTTPHeaders, config.HTTPHeaders{})

	return v
}

func generalstring(fl val.FieldLevel) bool {
	t := fl.Field().Interface().(string)
	return generalStringRegexp.MatchString(t)
}

func safestring(fl val.FieldLevel) bool {
	t := fl.Field().Interface().(string)
	return safeStringRegexp.MatchString(t)
}

func ltetomorrowutc(fl val.FieldLevel) bool {
	t := fl.Field().Interface().(time.Time)
	d := time.Now().UTC().Add(48 * time.Hour).Truncate(24 * time.Hour)
	return t.Before(d) || t.Equal(d)
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

// validateHTTPHeaders rejects wildcard mode combined with a non-empty
// allowlist - browsers reject a credentialed "*" response, so the
// combination is always dead configuration.
func validateHTTPHeaders(sl val.StructLevel) {
	headers := sl.Current().Interface().(config.HTTPHeaders)
	if len(headers.AccessControlAllowOrigins) > 0 && headers.AccessControlAllowOriginWildcard {
		sl.ReportError(headers.AccessControlAllowOriginWildcard, "access_control_allow_origin_wildcard", "AccessControlAllowOriginWildcard", "wildcard_with_allowlist", "")
	}
}
