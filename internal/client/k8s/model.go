package k8s

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	workbenchUserIDOffset uint64 = 1001
	defaultImageTag              = "latest"
	appInstanceNamePrefix        = "app-instance-"
	maxAppInstanceNameLen        = 15
	defaultAppName               = "unknown"

	maxWorkspaceServiceNameLen = 20
	defaultServiceName         = "unknown"
)

var appInstanceNameRegex = regexp.MustCompile("[^a-z0-9]+")
var workbenchUsernameRegex = regexp.MustCompile("[^a-z0-9_]")
var workspaceServiceNameRegex = regexp.MustCompile("[^a-z0-9]+")

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

	ShmSize               string
	BrowserConfigURL      string
	BrowserConfigJWTURL   string
	BrowserConfigJWTToken string
	MaxCPU                string
	MinCPU                string
	MaxMemory             string
	MinMemory             string
	MaxEphemeralStorage   string
	MinEphemeralStorage   string
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
	ID                     uint64                       `json:"-"`
	Name                   string                       `json:"-"`
	State                  string                       `json:"state,omitempty"`
	Chart                  WorkspaceServiceChart        `json:"chart"`
	Values                 map[string]any               `json:"values,omitempty"`
	Credentials            *WorkspaceServiceCredentials `json:"credentials,omitempty"`
	ConnectionInfoTemplate string                       `json:"connectionInfoTemplate,omitempty"`
	ComputedValues         map[string]string            `json:"computedValues,omitempty"`
}

// UID returns a unique, id-recoverable key for the service within the Workspace CR
// spec/status maps. The trailing "-<id>" lets the backend map operator-reported
// statuses back to the correct service instance even when a service name is reused
// across create/delete generations.
func (s WorkspaceInputService) UID() string {
	return fmt.Sprintf("%s-%d", s.SanitizedName(), s.ID)
}

// SanitizedName returns the service name reduced to a K8s/Helm-safe form.
func (s WorkspaceInputService) SanitizedName() string {
	name := strings.ToLower(s.Name)
	name = workspaceServiceNameRegex.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if len(name) > maxWorkspaceServiceNameLen {
		name = strings.Trim(name[:maxWorkspaceServiceNameLen], "-")
	}
	if name == "" {
		name = defaultServiceName
	}
	return name
}

// ParseWorkspaceServiceID extracts the service instance ID from a Workspace CR service
// key produced by WorkspaceInputService.UID (i.e. "<name>-<id>").
func ParseWorkspaceServiceID(key string) (uint64, error) {
	idx := strings.LastIndex(key, "-")
	if idx < 0 || idx == len(key)-1 {
		return 0, fmt.Errorf("invalid workspace service key %q", key)
	}
	id, err := strconv.ParseUint(key[idx+1:], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse workspace service ID from key %q: %w", key, err)
	}
	return id, nil
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
