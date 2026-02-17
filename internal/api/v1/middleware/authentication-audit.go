package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
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

	var opts []audit.Option

	if err != nil {
		// No audit on failure to prevent DDOS, infra level logs for this
	} else {
		opts = append(opts,
			audit.WithDescription("User authenticated successfully."),
			audit.WithDetail("username", req.Username),
		)
	}

	if len(opts) > 0 {
		audit.Record(ctx, a.auditWriter, model.AuditActionUserLogin, opts...)
	}

	return res, err
}

func (a authenticationControllerAudit) RefreshToken(ctx context.Context, req *chorus.RefreshTokenRequest) (*chorus.RefreshTokenReply, error) {
	// No audit for token refresh - this is an internal mechanism
	return a.next.RefreshToken(ctx, req)
}

func (a authenticationControllerAudit) AuthenticateOauth(ctx context.Context, req *chorus.AuthenticateOauthRequest) (*chorus.AuthenticateOauthReply, error) {
	res, err := a.next.AuthenticateOauth(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("oauth_provider_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to initiate OAuth authentication."),
			audit.WithError(err),
		)
		audit.Record(ctx, a.auditWriter, model.AuditActionUserLoginFailed, opts...)
	}
	// No audit on success - the actual login happens in the redirect

	return res, err
}

func (a authenticationControllerAudit) AuthenticateOauthRedirect(ctx context.Context, req *chorus.AuthenticateOauthRedirectRequest) (*chorus.AuthenticateOauthRedirectReply, error) {
	res, err := a.next.AuthenticateOauthRedirect(ctx, req)

	var opts []audit.Option

	if err != nil {
		// No audit on failure
	} else {
		opts = append(opts,
			audit.WithDescription("User authenticated successfully via OAuth."),
			audit.WithDetail("oauth_provider_id", req.Id),
		)
	}

	if len(opts) > 0 {
		audit.Record(ctx, a.auditWriter, model.AuditActionUserLogin, opts...)
	}

	return res, err
}

func (a authenticationControllerAudit) Logout(ctx context.Context, req *chorus.LogoutRequest) (*chorus.LogoutReply, error) {
	res, err := a.next.Logout(ctx, req)

	var opts []audit.Option

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to logout user."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription("User logged out successfully."),
		)
	}

	audit.Record(ctx, a.auditWriter, model.AuditActionUserLogout, opts...)

	return res, err
}
