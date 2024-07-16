package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthenticationController is the authentication service controller handler.
type AuthenticationController struct {
	authenticator service.Authenticator
}

// NewAuthenticationController returns a fresh authentication service controller instance.
func NewAuthenticationController(authenticator service.Authenticator) AuthenticationController {
	return AuthenticationController{authenticator: authenticator}
}

// Authenticate extracts the fields from an 'AuthenticationRequest' and passes them to the service.
func (a AuthenticationController) Authenticate(ctx context.Context, req *chorus.Credentials) (*chorus.AuthenticationReply, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials: %v", "empty request")
	}

	res, err := a.authenticator.Authenticate(ctx, req.Username, req.Password, req.Totp)
	if err != nil {
		switch err {
		case &service.Err2FARequired{}:
			return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
		default:
			return nil, status.Errorf(codes.Unauthenticated, "%v", err)
		}
	}
	return &chorus.AuthenticationReply{Result: &chorus.AuthenticationResult{Token: res}}, nil
}
