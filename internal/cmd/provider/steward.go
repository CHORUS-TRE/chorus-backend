package provider

import (
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	"github.com/CHORUS-TRE/chorus-backend/pkg/steward/service"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
)

var stewardControllerOnce sync.Once
var stewardController chorus.StewardServiceServer

func ProvideStewardController() chorus.StewardServiceServer {
	stewardControllerOnce.Do(func() {
		stewardController = v1.NewStewardController(ProvideStewardService())
		stewardController = ctrl_mw.StewardAuthorizing(logger.SecLog, []string{model.RoleChorus.String()})(stewardController)

	})
	return stewardController
}

var stewardServiceOnce sync.Once
var stewardService service.Stewarder

func ProvideStewardService() service.Stewarder {
	stewardServiceOnce.Do(func() {
		stewardService = service.NewStewardService(
			ProvideConfig(),
			ProvideTenanter(),
			ProvideUser(),
		)
	})

	return stewardService
}
