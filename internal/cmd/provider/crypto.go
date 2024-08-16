package provider

import (
	"context"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/crypto"
)

var daemonEncryptionKeyOnce sync.Once
var daemonEncryptionKey *crypto.Secret

func ProvideDaemonEncryptionKey() *crypto.Secret {
	daemonEncryptionKeyOnce.Do(func() {
		cfg := ProvideConfig()
		var err error
		daemonEncryptionKey, err = loadEncryptionKey(cfg.Daemon.PrivateKeyFile)
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "unable to load encryption key from: "+cfg.Daemon.PrivateKeyFile, zap.Error(err))
		}
	})
	return daemonEncryptionKey
}

var daemonPrivateKeyOnce sync.Once
var daemonPrivateKey *crypto.Secret

func ProvideDaemonPrivateKey() *crypto.Secret {
	daemonPrivateKeyOnce.Do(func() {
		cfg := ProvideConfig()
		var err error
		daemonPrivateKey, err = loadPrivateKey(cfg.Daemon.PrivateKeyFile)
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "unable to load private key from: "+cfg.Daemon.PrivateKeyFile, zap.Error(err))
		}
	})
	return daemonPrivateKey
}

func loadEncryptionKey(filename string) (*crypto.Secret, error) {
	//nolint:gosec
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %v: %w", filename, err)
	}

	return loadEncryptionKeyFromReader(file, filename)
}

func loadEncryptionKeyFromReader(r io.Reader, filename string) (*crypto.Secret, error) {
	secretPEM, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key content: %w", err)
	}

	block, _ := pem.Decode(secretPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key from file %v", filename)
	}
	if !strings.Contains(block.Type, "PRIVATE KEY") {
		return nil, fmt.Errorf("unexpected pem block type: %s", block.Type)
	}
	salt := ProvideConfig().Daemon.Salt
	return crypto.NewSecret(crypto.Derive(block.Bytes, []byte(salt)))
}

func loadPrivateKey(f string) (*crypto.Secret, error) {
	//nolint:gosec
	secretPEM, err := os.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %v: %w", f, err)
	}

	block, _ := pem.Decode(secretPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key from file %v", f)
	}
	return crypto.NewSecret(block.Bytes)
}
