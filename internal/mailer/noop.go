package mailer

import (
	"context"
	"html/template"
)

// NoOpMailer is a Mailer that discards every message, used when the mailer
// service is disabled.
type NoOpMailer struct{}

func NewNoOpMailer() *NoOpMailer {
	return &NoOpMailer{}
}

func (m *NoOpMailer) SendMessage(ctx context.Context, tenantID uint64, to []string, subject, title, message string) error {
	return nil
}

func (m *NoOpMailer) Send(ctx context.Context, tenantID uint64, to []string, subject string, tmpl *template.Template, data interface{}) error {
	return nil
}

func (m *NoOpMailer) GetSubject(ctx context.Context, tenantID uint64, emailKey string) string {
	return ""
}

func (m *NoOpMailer) GetTemplate(ctx context.Context, tenantID uint64, tmplKey TemplateKey) *template.Template {
	return nil
}
