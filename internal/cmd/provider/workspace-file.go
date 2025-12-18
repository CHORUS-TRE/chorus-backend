package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service/middleware"
)

var workspaceFileOnce sync.Once
var workspaceFile service.WorkspaceFiler

func ProvideWorkspaceFile() service.WorkspaceFiler {
	workspaceFileOnce.Do(func() {
		var err error
		workspaceFile, err = service.NewWorkspaceFileService(
			ProvideBlockStores(),
			ProvideConfig().Services.WorkspaceFileService.Stores,
		)
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to create workspace file service: "+err.Error())
		}
		workspaceFile = service_mw.Logging(logger.BizLog)(workspaceFile)
		workspaceFile = service_mw.Validation(ProvideValidator())(workspaceFile)
		workspaceFile = service_mw.WorkspaceCaching(logger.TechLog)(workspaceFile)
	})
	return workspaceFile
}

var workspaceFileControllerOnce sync.Once
var workspaceFileController chorus.WorkspaceFileServiceServer

func ProvideWorkspaceFileController() chorus.WorkspaceFileServiceServer {
	workspaceFileControllerOnce.Do(func() {
		workspaceFileController = v1.NewWorkspaceFileController(ProvideWorkspaceFile())
		workspaceFileController = ctrl_mw.WorkspaceFileAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator())(workspaceFileController)
	})
	return workspaceFileController
}
