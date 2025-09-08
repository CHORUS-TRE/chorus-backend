package provider

import (
	"context"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	gatekeeper_model "github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/model"
	gatekeeper_service "github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/service"
	"go.uber.org/zap"
)

var gatekeeperOnce sync.Once
var gatekeeper gatekeeper_service.AuthorizationServiceInterface

func ProvideGatekeeper() gatekeeper_service.AuthorizationServiceInterface {
	gatekeeperOnce.Do(func() {
		schema, err := gatekeeper_model.GetDefaultSchema()
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to get default gatekeeper model schema", zap.Error(err))
		}
		gatekeeper, err = gatekeeper_service.NewAuthorizationService(&schema)
	})
	return gatekeeper
}

var authorizerOnce sync.Once
var authorizer authorization.Authorizer

func ProvideAuthorizer() authorization.Authorizer {
	authorizerOnce.Do(func() {
		var err error
		authorizer, err = authorization.NewAuthorizer(ProvideGatekeeper())
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to create authorizer", zap.Error(err))
		}
	})
	return authorizer
}
