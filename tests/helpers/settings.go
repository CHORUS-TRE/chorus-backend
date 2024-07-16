//go:build unit || integration || acceptance

package helpers

import (
	"fmt"
	"os"
)

// Available environment variables to modify the test settings
const (
	COMPONENT_URL = "COMPONENT_URL"
)

func ComponentURL() string {
	if os.Getenv(COMPONENT_URL) != "" {
		return os.Getenv(COMPONENT_URL)
	}

	return fmt.Sprintf("%s:%s", Conf().Daemon.HTTP.Host, Conf().Daemon.HTTP.Port)
}
