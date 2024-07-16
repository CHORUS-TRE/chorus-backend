package middleware

import (
	"context"
	"html/template"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/mailer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type mailerServiceLogging struct {
	logger *logger.ContextLogger
	next   mailer.Mailer
}

func (c mailerServiceLogging) SendMessage(ctx context.Context, tenantID uint64, to []string, subject, title, message string) error {
	now := time.Now()
	err := c.next.SendMessage(ctx, tenantID, to, subject, title, message)
	if err != nil {
		c.logger.Error(ctx, "sending message has failed",
			zap.Error(err),
			zap.Strings("to", to),
			zap.String("subject", subject),
			zap.String("title", title),
			zap.String("message", message),
			zap.Float64("elapsed_ms", float64(time.Since(now).Nanoseconds())/1000000.0),
			zap.Uint64("tenant_id", tenantID),
		)
		return errors.Wrapf(err, "unable to send message")
	}

	c.logger.Info(ctx, "message sent",
		zap.Strings("to", to),
		zap.String("subject", subject),
		zap.String("title", title),
		zap.String("message", message),
		zap.Float64("elapsed_ms", float64(time.Since(now).Nanoseconds())/1000000.0),
		zap.Uint64("tenant_id", tenantID),
	)
	return nil
}

func (c mailerServiceLogging) Send(ctx context.Context, tenantID uint64, to []string, subject string, tmpl *template.Template, data interface{}) error {
	now := time.Now()
	err := c.next.Send(ctx, tenantID, to, subject, tmpl, data)
	if err != nil {
		c.logger.Error(ctx, "sending message has failed",
			zap.Error(err),
			zap.Strings("to", to),
			zap.String("subject", subject),
			zap.Float64("elapsed_ms", float64(time.Since(now).Nanoseconds())/1000000.0),
			zap.Uint64("tenant_id", tenantID),
		)
		return errors.Wrapf(err, "unable to send message")
	}

	c.logger.Info(ctx, "message sent",
		zap.Strings("to", to),
		zap.String("subject", subject),
		zap.Float64("elapsed_ms", float64(time.Since(now).Nanoseconds())/1000000.0),
		zap.Uint64("tenant_id", tenantID),
	)
	return nil
}

func (c mailerServiceLogging) GetSubject(ctx context.Context, tenantID uint64, emailKey string) string {
	return c.next.GetSubject(ctx, tenantID, emailKey)
}

func (c mailerServiceLogging) GetTemplate(ctx context.Context, tenantID uint64, tmplKey mailer.TemplateKey) *template.Template {
	return c.next.GetTemplate(ctx, tenantID, tmplKey)
}

func Logging(logger *logger.ContextLogger) func(mailer.Mailer) mailer.Mailer {
	return func(next mailer.Mailer) mailer.Mailer {
		return &mailerServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}
