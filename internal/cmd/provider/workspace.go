package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/store/postgres"
)

var workspaceOnce sync.Once
var workspace service.Workspaceer

func ProvideWorkspace() service.Workspaceer {
	workspaceOnce.Do(func() {
		workspace = service.NewWorkspaceService(
			ProvideWorkspaceStore(),
			ProvideHelmClient(),
		)
		workspace = service_mw.Logging(logger.BizLog)(workspace)
		workspace = service_mw.Validation(ProvideValidator())(workspace)
		workspace = service_mw.WorkspaceCaching(logger.TechLog)(workspace)
	})
	return workspace
}

var workspaceControllerOnce sync.Once
var workspaceController chorus.WorkspaceServiceServer

func ProvideWorkspaceController() chorus.WorkspaceServiceServer {
	workspaceControllerOnce.Do(func() {
		workspaceController = v1.NewWorkspaceController(ProvideWorkspace())
		workspaceController = ctrl_mw.WorkspaceAuthorizing(logger.SecLog, []string{user_model.RoleAuthenticated.String()})(workspaceController)
	})
	return workspaceController
}

var workspaceStoreOnce sync.Once
var workspaceStore service.WorkspaceStore

func ProvideWorkspaceStore() service.WorkspaceStore {
	workspaceStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("workspace-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			workspaceStore = postgres.NewWorkspaceStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		workspaceStore = store_mw.Logging(logger.TechLog)(workspaceStore)
	})
	return workspaceStore
}
