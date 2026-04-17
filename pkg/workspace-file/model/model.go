package model

type WorkspaceFileStoreStatus string

const (
	WorkspaceFileStoreStatusReady        WorkspaceFileStoreStatus = "Ready"
	WorkspaceFileStoreStatusDisconnected WorkspaceFileStoreStatus = "Disconnected"
	WorkspaceFileStoreStatusDisabled     WorkspaceFileStoreStatus = "Disabled"
)

type WorkspaceFileStoreInfo struct {
	Name        string
	Type        string
	Description string
	Status      WorkspaceFileStoreStatus
}