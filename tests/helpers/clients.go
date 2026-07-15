//go:build unit || integration || acceptance
// +build unit integration acceptance

package helpers

import (
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/openapi"
	app_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/app/client"
	auth_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/authentication/client"
	health_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/health/client"
	notification_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/notification/client"
	organization_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/organization/client"
	steward_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/steward/client"
	user_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/user/client"
	workspace_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/workspace/client"
	"github.com/go-openapi/strfmt"
)

var schemes = []string{"http"}

func AppServiceHTTPClient() *app_client.ChorusAppService {
	return app_client.NewHTTPClient(strfmt.Default)
}

func AuthenticationServiceHTTPClient() *auth_client.ChorusAuthenticationService {
	return auth_client.New(openapi.NewNopCloserClientTransport(ComponentURL(), "", schemes, logger.TechLog), strfmt.Default)
}

func UserServiceHTTPClient() *user_client.ChorusUserService {
	return user_client.New(openapi.NewNopCloserClientTransport(ComponentURL(), "", schemes, logger.TechLog), strfmt.Default)
}

func NotificationServiceHTTPClient() *notification_client.ChorusNotificationService {
	return notification_client.New(openapi.NewNopCloserClientTransport(ComponentURL(), "", schemes, logger.TechLog), strfmt.Default)
}

func HealthServiceHTTPClient() *health_client.ChorusHealthService {
	return health_client.New(openapi.NewNopCloserClientTransport(ComponentURL(), "", schemes, logger.TechLog), strfmt.Default)
}

func StewardServiceHTTPClient() *steward_client.ChorusStewardService {
	return steward_client.New(openapi.NewNopCloserClientTransport(ComponentURL(), "", schemes, logger.TechLog), strfmt.Default)
}

func WorkspaceServiceHTTPClient() *workspace_client.ChorusWorkspaceService {
	return workspace_client.New(openapi.NewNopCloserClientTransport(ComponentURL(), "", schemes, logger.TechLog), strfmt.Default)
}

func OrganizationServiceHTTPClient() *organization_client.ChorusOrganizationService {
	return organization_client.New(openapi.NewNopCloserClientTransport(ComponentURL(), "", schemes, logger.TechLog), strfmt.Default)
}
