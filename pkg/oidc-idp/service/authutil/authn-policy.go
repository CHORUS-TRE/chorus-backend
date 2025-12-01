package authutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	jwt "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	"github.com/google/uuid"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"go.uber.org/zap"
)

func Policy(cfg config.Config, userService Userer, authorizer authorization.Authorizer) goidc.AuthnPolicy {
	authenticator := authenticator{cfg: cfg, userService: userService, authorizer: authorizer}
	return goidc.NewPolicy(
		"main",
		func(r *http.Request, c *goidc.Client, as *goidc.AuthnSession) bool {
			// The flow starts at the login step.
			as.StoreParameter(paramStepID, stepIDLoadUser)

			if c.LogoURI != "" {
				as.StoreParameter(paramLogoURI, c.LogoURI)
			}
			if c.PolicyURI != "" {
				as.StoreParameter(paramPolicyURI, c.PolicyURI)
			}
			if c.TermsOfServiceURI != "" {
				as.StoreParameter(paramTermsOfServiceURI, c.TermsOfServiceURI)
			}

			return true
		},
		authenticator.authenticate,
	)
}

const (
	paramStepID         string = "step_id"
	stepIDLoadUser      string = "step_load_user"
	stepIDLogin         string = "step_login"
	stepIDCreateSession string = "step_create_session"
	stepIDConsent       string = "step_consent"
	stepIDFinishFlow    string = "step_finish_flow"

	paramAuthTime          string = "auth_time"
	paramTenantID          string = "tenant_id"
	paramRoles             string = "roles"
	paramUserSessionID     string = "user_session_id"
	paramLogoURI           string = "logo_uri"
	paramPolicyURI         string = "policy_uri"
	paramTermsOfServiceURI string = "tos_uri"

	consentFormParam string = "consent"

	cookieUserSessionID string = "goidc_username"
)

var userSessionStore = map[string]userSession{}

type userSession struct {
	ID       string
	Subject  string
	AuthTime int
}

type Userer interface {
	GetUser(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error)

	UpsertGrants(ctx context.Context, grants []user_model.UserGrant) error
	DeleteGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) error
	GetUserGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) ([]user_model.UserGrant, error)
}

type authenticator struct {
	authorizer  authorization.Authorizer
	cfg         config.Config
	userService Userer
}

func (a authenticator) authenticate(w http.ResponseWriter, r *http.Request, as *goidc.AuthnSession) (goidc.Status, error) {

	if as.StoredParameter(paramStepID) == stepIDLoadUser {
		if status, err := a.loadUser(r, as); status != goidc.StatusSuccess {
			return status, err
		}
		as.StoreParameter(paramStepID, stepIDLogin)
	}

	if as.StoredParameter(paramStepID) == stepIDLogin {
		if status, err := a.login(w, r, as); status != goidc.StatusSuccess {
			if status == goidc.StatusInProgress {
				as.StoreParameter(paramStepID, stepIDLoadUser)
			}
			return status, err
		}
		as.StoreParameter(paramStepID, stepIDCreateSession)
	}

	if as.StoredParameter(paramStepID) == stepIDCreateSession {
		if status, err := a.createUserSession(w, as); status != goidc.StatusSuccess {
			return status, err
		}
		as.StoreParameter(paramStepID, stepIDConsent)
	}

	if as.StoredParameter(paramStepID) == stepIDConsent {
		if status, err := a.grantConsent(w, r, as); status != goidc.StatusSuccess {
			return status, err
		}
		as.StoreParameter(paramStepID, stepIDFinishFlow)
	}

	if as.StoredParameter(paramStepID) == stepIDFinishFlow {
		return a.finishFlow(as)
	}

	return goidc.StatusFailure, errors.New("access denied")
}

func (a authenticator) loadUser(r *http.Request, as *goidc.AuthnSession) (goidc.Status, error) {
	ctx := r.Context()
	claims, err := jwt.ExtractJWTClaims(ctx)
	if err != nil {
		return goidc.StatusSuccess, nil
	}

	clientID := as.ClientID
	var client *config.OpenIDConnectProviderClient
	for _, c := range a.cfg.Services.OpenIDConnectProvider.Clients {
		if c.ID == clientID {
			client = &c
			break
		}
	}
	if client == nil {
		return goidc.StatusFailure, errors.New("client not found")
	}

	if client.OnlyPreLoggedForClient {
		if claims.ForClient != clientID {
			return goidc.StatusFailure, errors.New("user not pre-logged for this client")
		}
	}

	as.SetUserID(fmt.Sprintf("%d", claims.ID))
	as.StoreParameter(paramAuthTime, fmt.Sprintf("%d", claims.StandardClaims.IssuedAt))
	as.StoreParameter(paramTenantID, fmt.Sprintf("%d", claims.TenantID))
	rs, err := json.Marshal(claims.Roles)
	if err != nil {
		return goidc.StatusFailure, fmt.Errorf("unable to marshal roles: %w", err)
	}
	as.StoreParameter(paramRoles, string(rs))
	// as.StoreParameter(paramUserSessionID, session.ID)
	return goidc.StatusSuccess, nil
}

func (a authenticator) login(w http.ResponseWriter, r *http.Request, as *goidc.AuthnSession) (goidc.Status, error) {

	// If the user is unknown and the client requested no prompt for credentials,
	// return a login-required error.
	if as.Subject == "" && as.Prompt == goidc.PromptTypeNone {
		return goidc.StatusFailure, goidc.NewError(goidc.ErrorCodeLoginRequired, "user not logged in, cannot use prompt none")
	}

	// Determine if authentication is required.
	// Authentication is required if the user's identity is unknown or if the
	// client explicitly requested a login.
	mustAuthenticate := as.Subject == "" || as.Prompt == goidc.PromptTypeLogin
	// Additionally, check if the client specified a max age for the session.
	// If the max age is exceeded or 'auth_time' is unavailable, force re-authentication.
	if as.MaxAuthnAgeSecs != nil {
		maxAgeSecs := *as.MaxAuthnAgeSecs
		authTimeStr := as.StoredParameter(paramAuthTime)
		authTime, err := strconv.Atoi(fmt.Sprintf("%v", authTimeStr))
		if err != nil {
			return goidc.StatusFailure, fmt.Errorf("invalid auth time format: %w", err)
		}
		if TimestampNow() > authTime+maxAgeSecs {
			mustAuthenticate = true
		}
	}
	if !mustAuthenticate {
		return goidc.StatusSuccess, nil
	}

	callbackURL, err := callbackURL(a.cfg, "authorize", as.CallbackID, "load_user", nil)
	if err != nil {
		return goidc.StatusFailure, fmt.Errorf("unable to build callback URL: %w", err)
	}
	err = redirect(a.cfg, w, "login", callbackURL, nil)
	if err != nil {
		return goidc.StatusFailure, fmt.Errorf("unable to redirect to login page: %w", err)
	}

	return goidc.StatusInProgress, nil
}

func (a authenticator) createUserSession(w http.ResponseWriter, as *goidc.AuthnSession) (goidc.Status, error) {
	sessionID := uuid.NewString()
	if id := as.StoredParameter(paramUserSessionID); id != nil {
		sessionID = id.(string)
	}
	authTimeStr := as.StoredParameter(paramAuthTime).(string)
	authTime, err := strconv.Atoi(authTimeStr)
	if err != nil {
		return goidc.StatusFailure, fmt.Errorf("invalid auth time format: %w", err)
	}
	userSessionStore[sessionID] = userSession{
		ID:       sessionID,
		Subject:  as.Subject,
		AuthTime: authTime,
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieUserSessionID,
		Value:    sessionID,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})
	return goidc.StatusSuccess, nil
}

func (a authenticator) grantConsent(w http.ResponseWriter, r *http.Request, as *goidc.AuthnSession) (goidc.Status, error) {
	clientID := as.ClientID
	var client *config.OpenIDConnectProviderClient
	for _, c := range a.cfg.Services.OpenIDConnectProvider.Clients {
		if c.ID == clientID {
			client = &c
			break
		}
	}
	if client == nil {
		return goidc.StatusFailure, errors.New("client not found")
	}
	if client.GrantAutoApproved {
		return goidc.StatusSuccess, nil
	}

	user, err := a.GetUserFromSession(as)
	if err != nil {
		return goidc.StatusFailure, fmt.Errorf("unable to get user info: %w", err)
	}

	grants, err := a.userService.GetUserGrants(r.Context(), user.TenantID, user.ID, clientID)
	if err != nil {
		return goidc.StatusFailure, fmt.Errorf("unable to get user grants: %w", err)
	}

	requiredScopes := []string{}
	for _, s := range strings.Split(as.Scopes, " ") {
		if s == "openid" {
			continue
		}
		requiredScopes = append(requiredScopes, s)
	}

	scopesToGrant := []string{}
	for _, rs := range requiredScopes {
		found := false
		for _, g := range grants {
			if g.Scope == rs {
				found = true
				break
			}
		}
		if !found {
			scopesToGrant = append(scopesToGrant, rs)
		}
	}

	if len(scopesToGrant) == 0 {
		return goidc.StatusSuccess, nil
	}

	logger.TechLog.Debug(context.Background(), "consent request", zap.Any("as", as))

	consented := r.URL.Query().Get(consentFormParam)
	if consented == "" {
		v := url.Values{}
		scopesJSON, err := json.Marshal(scopesToGrant)
		if err != nil {
			return goidc.StatusFailure, fmt.Errorf("unable to marshal scopes to grant: %w", err)
		}

		v.Set("scopes", string(scopesJSON))
		v.Set("client_name", client.Name)

		callbackURL, err := callbackURL(a.cfg, "authorize", as.CallbackID, "grant", nil)
		if err != nil {
			return goidc.StatusFailure, fmt.Errorf("unable to build callback URL: %w", err)
		}
		err = redirect(a.cfg, w, "grant", callbackURL, v)
		if err != nil {
			return goidc.StatusFailure, fmt.Errorf("unable to redirect to consent page: %w", err)
		}

		return goidc.StatusInProgress, nil
	}

	if !isTrue(consented) {
		return goidc.StatusFailure, errors.New("consent not granted")
	}

	newGrants := []user_model.UserGrant{}
	var grantedUntil *time.Time
	if client.GrantDuration != nil {
		t := time.Now().Add(*client.GrantDuration)
		grantedUntil = &t
	}
	for _, s := range scopesToGrant {
		newGrants = append(newGrants, user_model.UserGrant{
			TenantID:     user.TenantID,
			UserID:       user.ID,
			ClientID:     client.ID,
			Scope:        s,
			GrantedUntil: grantedUntil,
		})
	}
	if err := a.userService.UpsertGrants(r.Context(), newGrants); err != nil {
		return goidc.StatusFailure, fmt.Errorf("unable to upsert user grants: %w", err)
	}

	return goidc.StatusSuccess, nil
}

func (a authenticator) finishFlow(as *goidc.AuthnSession) (goidc.Status, error) {
	as.GrantScopes(as.Scopes)
	as.GrantResources(as.Resources)
	as.GrantAuthorizationDetails(as.AuthDetails)

	authTimeStr := as.StoredParameter(paramAuthTime).(string)
	authTime, err := strconv.Atoi(authTimeStr)
	if err != nil {
		return goidc.StatusFailure, fmt.Errorf("invalid auth time format: %w", err)
	}
	as.SetIDTokenClaimAuthTime(authTime)
	as.SetIDTokenClaimACR(goidc.ACRMaceIncommonIAPSilver)

	logger.TechLog.Info(context.Background(), "finishing OIDC auth flow", zap.String("subject", as.Subject), zap.Int("auth_time", authTime), zap.Any("claims", as.Claims), zap.String("scopes", as.Scopes), zap.Any("response_type", as.ResponseType))

	// Add claims based on scope.
	setClaimFunc := as.SetUserInfoClaim
	if as.ResponseType == goidc.ResponseTypeIDToken {
		setClaimFunc = as.SetIDTokenClaim
	}

	if strings.Contains(as.Scopes, goidc.ScopeEmail.ID) || strings.Contains(as.Scopes, goidc.ScopeProfile.ID) || strings.Contains(as.Scopes, "roles") {
		user, err := a.GetUserFromSession(as)
		if err != nil {
			return goidc.StatusFailure, fmt.Errorf("unable to get user info: %w", err)
		}

		if strings.Contains(as.Scopes, goidc.ScopeEmail.ID) {
			setClaimFunc(goidc.ClaimEmail, user.Username)
		}

		if strings.Contains(as.Scopes, goidc.ScopeProfile.ID) {
			setClaimFunc(goidc.ClaimPreferredUsername, user.Username)
			setClaimFunc(goidc.ClaimGivenName, user.FirstName)
			setClaimFunc(goidc.ClaimUpdatedAt, user.UpdatedAt)
			setClaimFunc(goidc.ClaimName, user.FirstName+" "+user.LastName)
			setClaimFunc(goidc.ClaimFamilyName, user.LastName)
		}
		if strings.Contains(as.Scopes, "roles") {
			rolesStr := as.StoredParameter(paramRoles).(string)
			var jwtRoles []jwt.Role
			if err := json.Unmarshal([]byte(rolesStr), &jwtRoles); err != nil {
				return goidc.StatusFailure, fmt.Errorf("unable to unmarshal roles: %w", err)
			}

			roles := []string{}

			userRoles := []authorization.Role{}
			for _, r := range jwtRoles {
				ur, err := authorization.ToRole(r.Name, r.Context)
				if err != nil {
					return goidc.StatusFailure, fmt.Errorf("unable to convert jwt role to auth role: %w", err)
				}
				userRoles = append(userRoles, ur)
			}

			userWorkspacesMap := map[string]struct{}{}
			userWorkspaces := []string{}
			for _, r := range jwtRoles {
				if w, ok := r.Context["workspace"]; ok {
					if _, ok := userWorkspacesMap[w]; !ok {
						userWorkspacesMap[w] = struct{}{}
						userWorkspaces = append(userWorkspaces, w)
					}
				}
			}

			sort.Slice(userWorkspaces, func(i, j int) bool {
				w1ID, err := strconv.ParseUint(userWorkspaces[i], 10, 64)
				w2ID, err2 := strconv.ParseUint(userWorkspaces[j], 10, 64)
				if err != nil || err2 != nil {
					return userWorkspaces[i] < userWorkspaces[j]
				}
				return w1ID < w2ID
			})

			for _, w := range userWorkspaces {
				permission := authorization.NewPermission(authorization.PermissionModifyFilesInWorkspace, authorization.WithWorkspace(w))
				logger.TechLog.Debug(context.Background(), "checking user role for workspace", zap.String("workspace", w), zap.Uint64("user_id", uint64(user.ID)), zap.Any("permission", permission), zap.Any("user_roles", userRoles))
				authorized, err := a.authorizer.IsUserAllowed(userRoles, permission)
				if err != nil {
					logger.TechLog.Error(context.Background(), "authorization check error", zap.Error(err))
					continue
				}
				if !authorized {
					continue
				}

				rj := "workspace" + w

				roles = append(roles, rj)
			}
			setClaimFunc("roles", roles)
		}
	}

	return goidc.StatusSuccess, nil
}

func (a authenticator) GetUserFromSession(as *goidc.AuthnSession) (*user_model.User, error) {
	userIDStr := as.Subject
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}
	tenantIDStr := as.StoredParameter(paramTenantID).(string)
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID format: %w", err)
	}
	user, err := a.userService.GetUser(context.Background(), user_service.GetUserReq{
		ID:       uint64(userID),
		TenantID: uint64(tenantID),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get user info: %w", err)
	}

	return user, nil
}
