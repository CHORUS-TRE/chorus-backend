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
		if cfg.Daemon.PrivateKeyFile != "" {
			daemonEncryptionKey, err = loadEncryptionKey(cfg.Daemon.PrivateKeyFile)
			if err != nil {
				logger.TechLog.Fatal(context.Background(), "unable to load encryption key from: "+cfg.Daemon.PrivateKeyFile, zap.Error(err))
			}
		} else if cfg.Daemon.PrivateKey != "" {

		}

	})
	return daemonEncryptionKey
}

func loadEncryptionKey(filename string) (*crypto.Secret, error) {
	//nolint:gosec
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %v: %w", filename, err)
	}

	return loadEncryptionKeyFromReader(file)
}

func loadEncryptionKeyFromReader(r io.Reader) (*crypto.Secret, error) {
	secretPEM, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key content: %w", err)
	}

	return loadEncryptionKeyFromBytes(secretPEM)
}
func loadEncryptionKeyFromBytes(b []byte) (*crypto.Secret, error) {
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}
	if !strings.Contains(block.Type, "PRIVATE KEY") {
		return nil, fmt.Errorf("unexpected pem block type: %s", block.Type)
	}
	salt := ProvideConfig().Daemon.Salt
	return crypto.NewSecret(crypto.Derive(block.Bytes, []byte(salt)))
}
