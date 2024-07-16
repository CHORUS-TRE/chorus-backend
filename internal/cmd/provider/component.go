package provider

import (
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/component"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
)

// The following are descriptional parameters that described a
// particular instance of a chorus component.
var (
	componentName = "chorus"
	componentID   = "" // componentID is a unique ID generated at startup.
	version       = "" // version is set by the compiler.
	gitCommit     = "" // gitCommit is set by the compiler.
	goVersion     = "" // goVersion is set by the compiler.
)

type Info struct {
	Name               string `json:"name,omitempty"`
	Version            string `json:"version,omitempty"`
	RuntimeEnvironment string `json:"runtime_environment,omitempty"`
	ComponentID        string `json:"id,omitempty"`
	Commit             string `json:"commit,omitempty"`
	GoVersion          string `json:"-"`
}

var componentInfoOnce sync.Once

// ProvideComponentInfo returns the component Information.
func ProvideComponentInfo() *Info {
	componentInfoOnce.Do(func() {
		// Generate uuid for component.
		componentID = uuid.Next()
	})

	return &Info{
		Name:               componentName,
		Version:            version,
		RuntimeEnvironment: component.RuntimeEnvironment,
		ComponentID:        componentID,
		Commit:             gitCommit,
		GoVersion:          goVersion,
	}
}
