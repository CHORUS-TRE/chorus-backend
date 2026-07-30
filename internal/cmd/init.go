package cmd

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/crypto"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/jwks"

	val "github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// initConfigCmd writes a minimal config file
var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "create a minimal config file with the mandatory fields filled in",
	Long: `writes --config (default configs/config.yaml) containing only the fields check-config requires
against a stock default configuration. Refuses to overwrite an existing file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInitConfig()
	},
}

func init() {
	rootCmd.AddCommand(initConfigCmd)
}

// placeholder's literal value is also baked into internal/config/config.go's
// `validate:"ne=CHANGEME"` tags, keep these in sync.
const placeholder = "CHANGEME"

type mandatoryField struct {
	Path    string
	InitTag string
}

func runInitConfig() error {
	path := "configs/config.yaml"
	if len(configFilenames) > 1 {
		return fmt.Errorf("init-config takes at most one --config file, got %d", len(configFilenames))
	}
	if len(configFilenames) == 1 {
		path = configFilenames[0]
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists -- remove it first or point --config at a new path", path)
	}

	fields, err := mandatoryFields()
	if err != nil {
		return fmt.Errorf("unable to determine mandatory fields: %w", err)
	}

	tree := map[string]interface{}{}
	var placeholders []string
	for _, f := range fields {
		// Part of a required_without pair with daemon.private_key, which we
		// always generate below -- writing an inline key satisfies both.
		if f.Path == "daemon.private_key_file" {
			continue
		}

		value, err := generateValue(f.InitTag)
		if err != nil {
			return fmt.Errorf("unable to generate a value for %q: %w", f.Path, err)
		}
		if value == placeholder {
			placeholders = append(placeholders, f.Path)
		}
		setNestedValue(tree, f.Path, value)
	}

	out, err := yaml.Marshal(tree)
	if err != nil {
		return fmt.Errorf("unable to marshal config: %w", err)
	}
	out = annotateHarborClient(out)

	if err := os.WriteFile(path, out, 0o600); err != nil {
		return fmt.Errorf("unable to write %s: %w", path, err)
	}

	fmt.Printf("wrote %s\n", path)
	if len(placeholders) > 0 {
		fmt.Println("the following fields need a real value before their feature will work:")
		for _, p := range placeholders {
			fmt.Printf("  - %s\n", p)
		}
	}
	return nil
}

// mandatoryFields validates a code-level-defaults-only configuration and
// returns every failing field's dotted yaml path plus its own `init` struct
// tag
func mandatoryFields() ([]mandatoryField, error) {
	cfg := provider.ProvideDefaultConfig()
	err := provider.ProvideValidator().Struct(cfg)
	if err == nil {
		return nil, nil
	}

	var validationErrs val.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return nil, err
	}

	seen := map[string]bool{}
	var fields []mandatoryField
	for _, fe := range validationErrs {
		_, path, _ := strings.Cut(fe.Namespace(), ".")
		path = bracketToDotted(path)
		if seen[path] {
			continue
		}
		seen[path] = true

		initTag, _ := provider.FieldTag(cfg, fe, "init")
		fields = append(fields, mandatoryField{Path: path, InitTag: initTag})
	}
	return fields, nil
}

// bracketToDotted turns a FieldError namespace's map-indexed segments (e.g.
// "storage.datastores[audit].password") into a plain dotted path
// ("storage.datastores.audit.password"), matching real config.yaml nesting.
func bracketToDotted(path string) string {
	path = strings.ReplaceAll(path, "[", ".")
	return strings.ReplaceAll(path, "]", "")
}

// generateValue returns the value to write for a mandatory field, driven by
// its own `init` struct tag in internal/config/config.go:
//   - "random": a generated secret
//   - "privatekey": a generated EC private key PEM
//   - "jwks": a generated JWKS document
//   - "localdev=<value>": a fixed credential matching `make deps`
//   - "placeholder", or no tag at all (an as-yet-untagged mandatory field):
//     an explicit "CHANGEME" rather than an empty value
func generateValue(initTag string) (string, error) {
	tag, param, _ := strings.Cut(initTag, "=")
	switch tag {
	case "random":
		return randomToken()
	case "privatekey":
		return crypto.GeneratePrivateKeyPEM()
	case "jwks":
		jwksJSON, _, err := jwks.Generate()
		return jwksJSON, err
	case "localdev":
		return param, nil
	default:
		return placeholder, nil
	}
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("unable to generate random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// setNestedValue writes value at the given dotted path inside tree,
// creating intermediate maps as needed.
func setNestedValue(tree map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	m := tree
	for i, p := range parts {
		if i == len(parts)-1 {
			m[p] = value
			return
		}
		next, ok := m[p].(map[string]interface{})
		if !ok {
			next = map[string]interface{}{}
			m[p] = next
		}
		m = next
	}
}

// annotateHarborClient inserts a comment above the harbor_client key
// reminding the reader to set username/password for a private registry.
func annotateHarborClient(out []byte) []byte {
	const marker = "  harbor_client:\n"
	const note = "  # If this registry requires auth, also set username/password below.\n" + marker
	return bytes.Replace(out, []byte(marker), []byte(note), 1)
}
