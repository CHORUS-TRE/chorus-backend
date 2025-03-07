package v1

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

var _ chorus.AuthenticationServiceServer = (*AuthenticationController)(nil)

// AuthenticationController is the authentication service controller handler.
type AuthenticationController struct {
	authenticator             service.Authenticator
	refreshTokenAuthorization middleware.Authorization
}

// NewAuthenticationController returns a fresh authentication service controller instance.
func NewAuthenticationController(authenticator service.Authenticator) AuthenticationController {
	return AuthenticationController{
		authenticator:             authenticator,
		refreshTokenAuthorization: middleware.NewAuthorization(logger.SecLog, []string{user_model.RoleAuthenticated.String()}),
	}
}

func (a AuthenticationController) GetAuthenticationModes(ctx context.Context, req *chorus.GetAuthenticationModesRequest) (*chorus.GetAuthenticationModesReply, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials: %v", "empty request")
	}

	modes := a.authenticator.GetAuthenticationModes()

	res := []*chorus.AuthenticationMode{}

	for _, mode := range modes {
		if mode.Type == "internal" {
			res = append(res, &chorus.AuthenticationMode{
				Type: mode.Type,
				Internal: &chorus.Internal{
					PublicRegistrationEnabled: mode.Internal.PublicRegistrationEnabled,
				},
				ButtonText: mode.ButtonText,
				IconURL:    mode.IconURL,
			})
		}
		if mode.Type == "openid" {
			res = append(res, &chorus.AuthenticationMode{
				Type: mode.Type,
				Openid: &chorus.OpenID{
					Id: mode.OpenID.ID,
				},
				ButtonText: mode.ButtonText,
				IconURL:    mode.IconURL,
			})
		}
	}

	return &chorus.GetAuthenticationModesReply{Result: res}, nil
}

// Authenticate extracts the fields from an 'AuthenticationRequest' and passes them to the service.
func (a AuthenticationController) Authenticate(ctx context.Context, req *chorus.Credentials) (*chorus.AuthenticationReply, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials: %v", "empty request")
	}

	res, t, err := a.authenticator.Authenticate(ctx, req.Username, req.Password, req.Totp)
	if err != nil {
		switch err {
		case &service.Err2FARequired{}:
			return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
		default:
			return nil, status.Errorf(codes.Unauthenticated, "%v", err)
		}
	}

	expiresDate := time.Now().Add(t)
	expires := expiresDate.Format(time.RFC1123)

	header := metadata.Pairs("Set-Cookie", "jwttoken="+res+"; Path=/; SameSite=None; Expires="+expires)
	if err := grpc.SetHeader(ctx, header); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &chorus.AuthenticationReply{Result: &chorus.AuthenticationResult{Token: res}}, nil
}

func (a AuthenticationController) RefreshToken(ctx context.Context, req *chorus.RefreshTokenRequest) (*chorus.AuthenticationReply, error) {
	err := a.refreshTokenAuthorization.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}

	res, t, err := a.authenticator.RefreshToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%v", err)
	}

	expiresDate := time.Now().Add(t)
	expires := expiresDate.Format(time.RFC1123)

	header := metadata.Pairs("Set-Cookie", "jwttoken="+res+"; Path=/; SameSite=None; Expires="+expires)
	if err := grpc.SetHeader(ctx, header); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

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

	token, t, url, err := a.authenticator.OAuthCallback(ctx, req.Id, req.State, req.SessionState, req.Code)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	expiresDate := time.Now().Add(t)
	expires := expiresDate.Format(time.RFC1123)

	header := metadata.Pairs("Set-Cookie", "jwttoken="+token+"; Path=/; SameSite=None; Expires="+expires)
	if err := grpc.SetHeader(ctx, header); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	if url != "" {
		header := metadata.Pairs("Location", url)
		if err := grpc.SetHeader(ctx, header); err != nil {
			return nil, status.Errorf(codes.Internal, "%v", err)
		}
	}

	return &chorus.AuthenticateOauthRedirectReply{Result: &chorus.AuthenticateOauthRedirectResult{Token: token}}, nil
}

func (a AuthenticationController) Logout(ctx context.Context, req *chorus.LogoutRequest) (*chorus.LogoutReply, error) {
	err := a.refreshTokenAuthorization.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}

	redirectURL, err := a.authenticator.Logout(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	header := metadata.Pairs("Set-Cookie", "jwttoken=; Path=/; SameSite=None; Expires=Thu, 01 Jan 1970 00:00:00 GMT")
	if err := grpc.SetHeader(ctx, header); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	if redirectURL != "" {
		header := metadata.Pairs("Location", redirectURL)
		if err := grpc.SetHeader(ctx, header); err != nil {
			return nil, status.Errorf(codes.Internal, "%v", err)
		}
	}

	return &chorus.LogoutReply{RedirectURL: utils.FromString(redirectURL)}, nil
}
