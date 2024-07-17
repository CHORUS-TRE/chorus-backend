package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestDecodeSecret(t *testing.T) {
	var val struct {
		MySecret Sensitive `yaml:"my_secret"`
		MyString string    `yaml:"my_string"`
	}
	encoded := `
my_secret: 'password'
my_string: 'string'
`
	err := yaml.Unmarshal([]byte(encoded), &val)
	require.NoError(t, err)
	require.Equal(t, Sensitive("password"), val.MySecret)
	require.Equal(t, "string", val.MyString)
}

func TestEncodeSecret(t *testing.T) {
	val := struct {
		SomeString      string    `yaml:"some_string"`
		EmptyString     string    `yaml:"empty_string"`
		SomeStringOmit  string    `yaml:"some_string_omit,omitempty"`
		EmptyStringOmit string    `yaml:"empty_string_omit,omitempty"`
		SomeSecret      Sensitive `yaml:"some_secret"`
		EmptySecret     Sensitive `yaml:"empty_secret"`
		SomeSecretOmit  Sensitive `yaml:"some_secret_omit,omitempty"`
		EmptySecretOmit Sensitive `yaml:"empty_secret_omit,omitempty"`
	}{
		SomeString:     "string1",
		SomeStringOmit: "string2",
		SomeSecret:     "password1",
		SomeSecretOmit: "password2",
	}
	encoded, err := yaml.Marshal(val)
	require.NoError(t, err)
	require.Equal(t,
		`some_string: string1
empty_string: ""
some_string_omit: string2
some_secret: <redacted>
empty_secret: ""
some_secret_omit: <redacted>
`, string(encoded))
}
