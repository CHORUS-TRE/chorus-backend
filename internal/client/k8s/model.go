package k8s

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	workbenchUserIDOffset uint64 = 1001
	defaultImageTag              = "latest"
	appInstanceNamePrefix        = "app-instance-"
	maxAppInstanceNameLen        = 15
	defaultAppName               = "unknown"
)

var appInstanceNameRegex = regexp.MustCompile("[^a-z0-9]+")
var workbenchUsernameRegex = regexp.MustCompile("[^a-z0-9_]")

// ----------------------------------------------------------------
// Models and related methods
// ----------------------------------------------------------------

type Workbench struct {
	CurrentGeneration       int64
	ObservedGeneration      int64
	Namespace               string
	TenantID                uint64
	Username                string
	UserID                  uint64
	Name                    string
	InitialResolutionWidth  uint32
	InitialResolutionHeight uint32
	Clipboard               string
	Status                  string
	ServerPodStatus         string
	ServerPodMessage        string
	Apps                    []AppInstance
}

func (w Workbench) UID() string {
	return fmt.Sprintf("workbench-%v", w.TenantID)
}

func (w Workbench) SanitizedUsername() string {
	name := strings.ToLower(w.Username)
	name = strings.ReplaceAll(name, " ", "_")
	name = workbenchUsernameRegex.ReplaceAllString(name, "")

	return name
}

type AppInstance struct {
	ID      uint64
	AppName string

	AppRegistry string
	AppImage    string
	AppTag      string

	K8sState   string
	K8sStatus  string
	K8sMessage string

	ShmSize             string
	KioskConfigURL      string
	KioskConfigJWTURL   string
	KioskConfigJWTToken string
	MaxCPU              string
	MinCPU              string
	MaxMemory           string
	MinMemory           string
	MaxEphemeralStorage string
	MinEphemeralStorage string
}

func (a AppInstance) UID() string {
	return fmt.Sprintf("%s%v", appInstanceNamePrefix, a.ID)
}

func (a AppInstance) SanitizedAppName() string {
	name := strings.ToLower(a.AppName)

	// Strip existing ID suffix if present
	name = strings.TrimSuffix(name, fmt.Sprintf("-%d", a.ID))

	name = appInstanceNameRegex.ReplaceAllString(name, "-")
	if len(name) > maxAppInstanceNameLen {
		name = name[:maxAppInstanceNameLen]
	}
	name = strings.Trim(name, "-")

	if name == "" {
		name = defaultAppName
	}

	return name
}

// ----------------------------------------------------------------
// Workspace input/output models for K8s client
// ----------------------------------------------------------------

type WorkspaceInputService struct {
	Chart                  WorkspaceServiceChart        `json:"chart"`
	Values                 map[string]any               `json:"values,omitempty"`
	Credentials            *WorkspaceServiceCredentials `json:"credentials,omitempty"`
	ConnectionInfoTemplate string                       `json:"connectionInfoTemplate,omitempty"`
	ComputedValues         map[string]string            `json:"computedValues,omitempty"`
}

type WorkspaceInput struct {
	TenantID      uint64
	Namespace     string
	NetworkPolicy string
	AllowedFQDNs  []string
	Clipboard     string
	Services      map[string]WorkspaceInputService
}

type WorkspaceOutput struct {
	CurrentGeneration  int64
	ObservedGeneration int64
	Namespace          string
	TenantID           uint64

	NetworkPolicyStatus  string
	NetworkPolicyMessage string

	ServiceStatuses map[string]WorkspaceServiceStatusOutput
}

type WorkspaceServiceStatusOutput struct {
	Status         string
	Message        string
	ConnectionInfo string
	SecretName     string
}
