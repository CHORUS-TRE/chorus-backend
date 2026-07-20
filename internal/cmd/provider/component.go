package provider

import (
	"runtime"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/component"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
)

// The following are descriptional parameters that described a
// particular instance of a chorus component. version and gitCommit
// are set by the compiler; componentName is a static constant.
var (
	componentName = "chorus"
	version       = "" // version is set by the compiler.
	gitCommit     = "" // gitCommit is set by the compiler.
)

type Info struct {
	Name               string `json:"name,omitempty"`
	Version            string `json:"version,omitempty"`
	RuntimeEnvironment string `json:"runtime_environment,omitempty"`
	ComponentID        string `json:"id,omitempty"`
	Commit             string `json:"commit,omitempty"`
	GoVersion          string `json:"-"`
}

// componentID is generated once at runtime (not by the compiler), the
// first time ProvideComponentInfo is called.
var (
	componentID       string
	componentInfoOnce sync.Once
)

// ProvideComponentInfo returns the component Information.
func ProvideComponentInfo() *Info {
	componentInfoOnce.Do(func() {
		componentID = uuid.Next()
	})

	return &Info{
		Name:               componentName,
		Version:            version,
		RuntimeEnvironment: component.RuntimeEnvironment,
		ComponentID:        componentID,
		Commit:             gitCommit,
		GoVersion:          runtime.Version(),
	}
}
