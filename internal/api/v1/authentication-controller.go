package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service"
	"github.com/pkg/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

	header := metadata.Pairs("Set-Cookie", "jwttoken="+res+"; Path=/")
	grpc.SetHeader(ctx, header)

	return &chorus.AuthenticationReply{Result: &chorus.AuthenticationResult{Token: res}}, nil
}

func (a AuthenticationController) AuthenticateOauth(ctx context.Context, req *chorus.AuthenticateOauthRequest) (*chorus.AuthenticateOauthReply, error) {
	if req == nil || req.Id == "" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid id: %v", "empty request")
	}

	uri, err := a.authenticator.AuthenticateOAuth(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	return &chorus.AuthenticateOauthReply{Result: &chorus.AuthenticateOauthResult{RedirectURI: uri}}, nil
}

func (a AuthenticationController) AuthenticateOauthRedirect(ctx context.Context, req *chorus.AuthenticateOauthRedirectRequest) (*chorus.AuthenticateOauthRedirectReply, error) {
	if req == nil || req.Id == "" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid id: %v", "empty request")
	}

	token, err := a.authenticator.OAuthCallback(ctx, req.Id, req.State, req.SessionState, req.Code)
	if err != nil {
		return nil, errors.Wrap(err, "error callback")
	}

	header := metadata.Pairs("Set-Cookie", "jwttoken="+token+"; Path=/")
	grpc.SetHeader(ctx, header)

	return &chorus.AuthenticateOauthRedirectReply{Result: &chorus.AuthenticateOauthRedirectResult{Token: token}}, nil
}
