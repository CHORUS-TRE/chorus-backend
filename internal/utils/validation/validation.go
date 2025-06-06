package validation

import (
	"context"
	"regexp"
	"time"

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
