package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/jordan-wright/email"
	"github.com/pkg/errors"
)

type Mailer interface {
	SendMessage(ctx context.Context, tenantID uint64, to []string, subject, title, message string) error
	Send(ctx context.Context, tenantID uint64, to []string, subject string, tmpl *template.Template, data interface{}) error
	GetSubject(ctx context.Context, tenantID uint64, emailKey string) string
	GetTemplate(ctx context.Context, tenantID uint64, tmplKey TemplateKey) *template.Template
}

type MailerService struct {
	hostPort      string
	from          map[uint64]string
	auth          smtp.Auth
	emailSubjects map[uint64]map[string]string

	templates map[string]*template.Template
	tlsConfig *tls.Config
}

// NewMailerService returns a new MailerService.
func NewMailerService(user, password, authentication string, from map[uint64]string, hostPort string, emailSubjects map[uint64]map[string]string, tlsconfig *tls.Config) (*MailerService, error) {

	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse host:port: %v", hostPort)
	}

	if strings.ToLower(authentication) != "none" {
		if user == "" {
			return &MailerService{}, fmt.Errorf("user cannot be empty when using authentication")
		}
		if password == "" {
			return &MailerService{}, fmt.Errorf("password cannot be empty when using authentication")
		}
	}

	var auth smtp.Auth
	switch strings.ToLower(authentication) {

	case "none":
		auth = nil
	case "plain":
		auth = smtp.PlainAuth("", user, password, host)
	case "login":
		auth = NewLoginAuth(user, password)
	}

	// Setup the mailer.
	var m = &MailerService{
		hostPort:      fmt.Sprintf("%v:%v", host, port),
		from:          from,
		auth:          auth,
		emailSubjects: emailSubjects,
		templates:     make(map[string]*template.Template),
		tlsConfig:     tlsconfig,
	}

	// Prepare all the templates required.
	for name, mailTemplate := range mailTemplates {
		if err := m.prepare(name.String(), mailTemplate); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// prepare is a helper to prepare a template for future use.
func (m *MailerService) prepare(name, mailTemplate string) error {
	var tmpl, err = template.New(name).Parse(mailTemplate)
	if err != nil {
		return errors.Wrapf(err, "could not prepare template %q", name)
	}
	m.templates[name] = tmpl
	return nil
}

func (m *MailerService) SendMessage(ctx context.Context, tenantID uint64, to []string, subject, title, message string) error {
	return m.Send(ctx, tenantID, to, subject, m.GetTemplate(ctx, tenantID, "titleText"), TitleText{Title: title, Text: message})
}

func (m *MailerService) Send(ctx context.Context, tenantID uint64, to []string, subject string, tmpl *template.Template, data interface{}) error {
	if tmpl == nil {
		return errors.New("template should not be nil")
	}
	// Execute template.
	var b = new(bytes.Buffer)
	{
		var err = tmpl.Execute(b, data)
		if err != nil {
			return err
		}
	}

	from := "CHORUS <no-reply@chorus-tre.ch>"
	if strings.Contains(m.from[tenantID], "@") {
		from = m.from[tenantID]
	}

	// Send mail
	var mail = &email.Email{
		To:      to,
		From:    from,
		Subject: subject,
		HTML:    b.Bytes(),
		Headers: textproto.MIMEHeader{},
	}

	var err error
	if m.tlsConfig.InsecureSkipVerify || m.tlsConfig.RootCAs != nil {
		err = mail.SendWithStartTLS(m.hostPort, m.auth, m.tlsConfig)
	} else {
		err = mail.Send(m.hostPort, m.auth)
	}
	if err != nil {
		return errors.Wrapf(err, "could not send mail via %v: from: %v, to: %v, subject: %v", m.hostPort, from, to, subject)
	}

	return nil
}

// GetSubject returns a custom email subject if it exists. The emailKey is case insensitive.
func (m *MailerService) GetSubject(ctx context.Context, tenantID uint64, emailKey string) string {
	if m.emailSubjects[tenantID] == nil {
		return ""
	}
	return m.emailSubjects[tenantID][strings.ToLower(emailKey)]
}

// GetTemplate returns the email template if it exists.
func (m *MailerService) GetTemplate(ctx context.Context, tenantID uint64, tmplKey TemplateKey) *template.Template {
	return m.templates[tmplKey.String()]
}
