package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.AuthenticationServiceServer = (*authenticationControllerAudit)(nil)

type authenticationControllerAudit struct {
	next        chorus.AuthenticationServiceServer
	auditWriter service.AuditWriter
}

func NewAuthenticationAuditMiddleware(auditWriter service.AuditWriter) func(chorus.AuthenticationServiceServer) chorus.AuthenticationServiceServer {
	return func(next chorus.AuthenticationServiceServer) chorus.AuthenticationServiceServer {
		return &authenticationControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (a authenticationControllerAudit) GetAuthenticationModes(ctx context.Context, req *chorus.GetAuthenticationModesRequest) (*chorus.GetAuthenticationModesReply, error) {
	// No audit for getting authentication modes - this is public information
	return a.next.GetAuthenticationModes(ctx, req)
}

func (a authenticationControllerAudit) Authenticate(ctx context.Context, req *chorus.Credentials) (*chorus.AuthenticationReply, error) {
	res, err := a.next.Authenticate(ctx, req)
	if err != nil {
		// No audit on failure
	} else {
		audit.Record(ctx, a.auditWriter,
			model.AuditActionUserLogin,
			audit.WithDescription("User authenticated successfully."),
			audit.WithDetail("username", req.Username),
		)
	}

	return res, err
}

func (a authenticationControllerAudit) RefreshToken(ctx context.Context, req *chorus.RefreshTokenRequest) (*chorus.RefreshTokenReply, error) {
	// No audit for token refresh - this is an internal mechanism
	return a.next.RefreshToken(ctx, req)
}

func (a authenticationControllerAudit) AuthenticateOauth(ctx context.Context, req *chorus.AuthenticateOauthRequest) (*chorus.AuthenticateOauthReply, error) {
	res, err := a.next.AuthenticateOauth(ctx, req)
	if err != nil {
		audit.Record(ctx, a.auditWriter,
			model.AuditActionUserLoginFailed,
			audit.WithDescription("Failed to initiate OAuth authentication."),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("oauth_provider_id", req.Id),
		)
	}
	// No audit here - the actual login happens in the redirect

	return res, err
}

func (a authenticationControllerAudit) AuthenticateOauthRedirect(ctx context.Context, req *chorus.AuthenticateOauthRedirectRequest) (*chorus.AuthenticateOauthRedirectReply, error) {
	res, err := a.next.AuthenticateOauthRedirect(ctx, req)
	if err != nil {
		// No audit on failure
	} else {
		audit.Record(ctx, a.auditWriter,
			model.AuditActionUserLogin,
			audit.WithDescription("User authenticated successfully via OAuth."),
			audit.WithDetail("oauth_provider_id", req.Id),
		)
	}

	return res, err
}

func (a authenticationControllerAudit) Logout(ctx context.Context, req *chorus.LogoutRequest) (*chorus.LogoutReply, error) {
	res, err := a.next.Logout(ctx, req)
	if err != nil {
		audit.Record(ctx, a.auditWriter,
			model.AuditActionUserLogout,
			audit.WithDescription("Failed to logout user."),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
		)
	} else {
		audit.Record(ctx, a.auditWriter,
			model.AuditActionUserLogout,
			audit.WithDescription("User logged out successfully."),
		)
	}

	return res, err
}
