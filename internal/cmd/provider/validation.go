package provider

import (
	"reflect"
	"strings"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/validation"

	val "github.com/go-playground/validator/v10"
)

var validatorOnce sync.Once
var validator *val.Validate

func ProvideValidator() *val.Validate {
	validatorOnce.Do(func() {
		validator = validation.NewValidator()

		// Report the yaml tag name instead of the Go field name (e.g.
		// "daemon.grpc.host" instead of "Daemon.GRPC.Host") -- matches the
		// dotted path used everywhere else (--set, CHORUS_* env vars).
		// Falls back to the Go field name for structs with no yaml tag.
		validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	})
	return validator
}
