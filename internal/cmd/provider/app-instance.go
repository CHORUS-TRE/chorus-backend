package provider

import (
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

var appInstanceControllerOnce sync.Once
var appInstanceController chorus.AppInstanceServiceServer

func ProvideAppInstanceController() chorus.AppInstanceServiceServer {
	appInstanceControllerOnce.Do(func() {
		appInstanceController = v1.NewAppInstanceController(ProvideWorkbench())
		appInstanceController = ctrl_mw.AppInstanceAuthorizing(logger.SecLog, []string{user_model.RoleAuthenticated.String()})(appInstanceController)
	})
	return appInstanceController
}
