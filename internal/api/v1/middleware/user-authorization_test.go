package middleware

import (
	"context"
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

type createUserServer struct {
	chorus.UnimplementedUserServiceServer
	called bool
}

func (s *createUserServer) CreateUser(context.Context, *chorus.User) (*chorus.CreateUserReply, error) {
	s.called = true
	return &chorus.CreateUserReply{}, nil
}

func TestUserAuthorizingAllowsAnonymousCreateUserWhenInternalPublicRegistrationEnabled(t *testing.T) {
	cfg := config.Config{}
	cfg.Services.AuthenticationService.Modes = map[string]config.Mode{
		"internal": {
			Type:                      "internal",
			Enabled:                   true,
			PublicRegistrationEnabled: true,
		},
	}

	next := &createUserServer{}
	server := UserAuthorizing(cfg, nil, nil)(next)

	if _, err := server.CreateUser(context.Background(), &chorus.User{}); err != nil {
		t.Fatalf("CreateUser() returned error: %v", err)
	}
	if !next.called {
		t.Fatal("CreateUser() did not call next server")
	}
}

func TestInternalPublicRegistrationRequiresEnabledInternalMode(t *testing.T) {
	cfg := config.Config{}
	cfg.Services.AuthenticationService.Modes = map[string]config.Mode{
		"disabled-internal": {
			Type:                      "internal",
			Enabled:                   false,
			PublicRegistrationEnabled: true,
		},
		"openid": {
			Type:                      "openid",
			Enabled:                   true,
			PublicRegistrationEnabled: true,
		},
	}

	if isInternalPublicRegistrationEnabled(cfg) {
		t.Fatal("expected public registration to be disabled without an enabled internal mode")
	}
}