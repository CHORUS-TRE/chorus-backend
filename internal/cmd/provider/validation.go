package provider

import (
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/validation"

	val "github.com/go-playground/validator/v10"
)

var validatorOnce sync.Once
var validator *val.Validate

func ProvideValidator() *val.Validate {
	validatorOnce.Do(func() {
		validator = validation.NewValidator()
	})
	return validator
}
