//go:build unit || integration || acceptance
// +build unit integration acceptance

package helpers

import (
	attachment_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/attachment/client"
	auth_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/authentication/client"
	health_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/health/client"
	notification_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/notification/client"
	steward_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/steward/client"
	user_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/user/client"
	"github.com/go-openapi/strfmt"
)

var schemes = []string{"http"}

func AuthenticationServiceHTTPClient() *auth_client.ChorusAuthenticationService {
	return auth_client.NewHTTPClient(strfmt.Default)
}

func AttachmentServiceHTTPClient() *attachment_client.ChorusAttachmentService {
	return attachment_client.NewHTTPClient(strfmt.Default)
}

func UserServiceHTTPClient() *user_client.ChorusUserService {
	return user_client.NewHTTPClient(strfmt.Default)
}

func NotificationServiceHTTPClient() *notification_client.ChorusNotificationService {
	return notification_client.NewHTTPClient(strfmt.Default)
}

func HealthServiceHTTPClient() *health_client.ChorusHealthService {
	return health_client.NewHTTPClient(strfmt.Default)
}

func StewardServiceHTTPClient() *steward_client.ChorusStewardService {
	return steward_client.NewHTTPClient(strfmt.Default)
}
