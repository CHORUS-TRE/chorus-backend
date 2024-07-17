package config

const (
	redacted string = "<redacted>"
)

/*
Sensitive is a container for sensitive strings,
which can be read from yaml config files,
but marshaling them to yaml will redact them.
*/
type Sensitive string

func (s Sensitive) MarshalYAML() (interface{}, error) {
	if len(s) == 0 {
		return "", nil
	}
	return redacted, nil
}

// PlainText returns the plain text of a sensitive string
func (s Sensitive) PlainText() string {
	return string(s)
}
