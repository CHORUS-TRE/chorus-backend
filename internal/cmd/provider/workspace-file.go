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

var workspaceFileControllerOnce sync.Once
var workspaceFileController chorus.WorkspaceFileServiceServer

func ProvideWorkspaceFileController() chorus.WorkspaceFileServiceServer {
	workspaceFileControllerOnce.Do(func() {
		workspaceFileController = v1.NewWorkspaceFileController(ProvideWorkspaceFileService())
		workspaceFileController = ctrl_mw.WorkspaceFileAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator())(workspaceFileController)
		if ProvideConfig().Services.AuditService.Enabled {
			workspaceFileController = ctrl_mw.NewWorkspaceFileAuditMiddleware(ProvideAuditWriter())(workspaceFileController)
		}
	})
	return workspaceFileController
}

var workspaceFileServiceOnce sync.Once
var workspaceFileService service.WorkspaceFiler

func ProvideWorkspaceFileService() service.WorkspaceFiler {
	workspaceFileServiceOnce.Do(func() {
		var err error
		workspaceFileService, err = service.NewWorkspaceFileService(
			ProvideFileStores(),
			ProvideConfig().Services.WorkspaceFileService.Stores,
		)
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to create workspace file service: "+err.Error())
		}
		workspaceFileService = service_mw.Logging(logger.BizLog)(workspaceFileService)
		workspaceFileService = service_mw.Validation(ProvideValidator())(workspaceFileService)
		workspaceFileService = service_mw.WorkspaceCaching(logger.TechLog)(workspaceFileService)
	})
	return workspaceFileService
}
