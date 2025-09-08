package provider

import (
	"context"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"go.uber.org/zap"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	"github.com/CHORUS-TRE/chorus-backend/pkg/steward/service"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
)

var stewardControllerOnce sync.Once
var stewardController chorus.StewardServiceServer

func ProvideStewardController() chorus.StewardServiceServer {
	stewardControllerOnce.Do(func() {
		stewardController = v1.NewStewardController(ProvideStewardService())
		stewardController = ctrl_mw.StewardAuthorizing(logger.SecLog, ProvideAuthorizer())(stewardController)

	})
	return stewardController
}

var stewardServiceOnce sync.Once
var stewardService service.Stewarder

func ProvideStewardService() service.Stewarder {
	stewardServiceOnce.Do(func() {
		var err error
		stewardService, err = service.NewStewardService(
			ProvideConfig(),
			ProvideTenanter(),
			ProvideUser(),
			ProvideWorkspace(),
		)
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to create steward service", zap.Error(err))
		}
	})

	return stewardService
}
