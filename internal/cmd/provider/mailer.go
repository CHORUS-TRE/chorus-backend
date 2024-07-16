package provider

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	mailerService "github.com/CHORUS-TRE/chorus-backend/internal/mailer"
	"github.com/CHORUS-TRE/chorus-backend/internal/mailer/middleware"
	"go.uber.org/zap"
)

var mailerOnce sync.Once
var mailer mailerService.Mailer

func ProvideMailer() mailerService.Mailer {
	mailerOnce.Do(func() {

		cfg := ProvideConfig()

		var rootCas *x509.CertPool
		certificatesRepo := cfg.Services.MailerService.SMTP.CertificatesRepo
		if certificatesRepo != "" {
			var files []string
			err := filepath.Walk(certificatesRepo, func(path string, info fs.FileInfo, err error) error {
				if filepath.Ext(path) == ".pem" {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				logger.TechLog.Warn(context.Background(), "unable to walk through repo "+certificatesRepo, zap.Error(err))
			}

			rootCas = x509.NewCertPool()
			for _, f := range files {
				//nolint:gosec
				certif, err := os.ReadFile(f)
				if err != nil {
					logger.TechLog.Fatal(context.Background(), "unable to read file "+f, zap.Error(err))
				}
				ok := rootCas.AppendCertsFromPEM(certif)
				if !ok {
					logger.TechLog.Fatal(context.Background(), "unable to add certificate "+f, zap.Error(err))
				}
			}
		}
		tlsConfig := &tls.Config{
			RootCAs:    rootCas,
			ServerName: cfg.Services.MailerService.SMTP.ServerName,
			//nolint:gosec
			InsecureSkipVerify: cfg.Services.MailerService.SMTP.InsecureMode,
		}
		var err error
		mailer, err = mailerService.NewMailerService(
			cfg.Services.MailerService.SMTP.User,
			cfg.Services.MailerService.SMTP.Password.PlainText(),
			cfg.Services.MailerService.SMTP.Authentication,
			ProvideFroms(),
			fmt.Sprintf("%s:%s", cfg.Services.MailerService.SMTP.Host, cfg.Services.MailerService.SMTP.Port),
			make(map[uint64]map[string]string),
			tlsConfig,
		)
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "unable to provide mailer service", zap.Error(err))
		}

		mailer = middleware.Logging(logger.TechLog)(mailer)
	})
	return mailer
}

func ProvideFroms() map[uint64]string {

	froms := make(map[uint64]string)
	cfg := ProvideConfig()

	for id, tenant := range cfg.Tenants {
		froms[id] = fmt.Sprintf(`"%s" <%s>`, tenant.Mailing.Sender.FromName, tenant.Mailing.Sender.FromEmail)
	}
	return froms
}
