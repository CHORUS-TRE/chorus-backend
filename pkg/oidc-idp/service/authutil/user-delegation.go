package authutil

import (
	"fmt"
	"net/http"
	"slices"

	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	"github.com/luikyv/go-oidc/pkg/goidc"
)

func UserDelegationHandleGrantFunc(
	clientManager goidc.ClientManager,
	userService Userer,
) goidc.HandleGrantFunc {
	return func(r *http.Request, gi *goidc.GrantInfo) error {
		if gi.GrantType != goidc.GrantClientCredentials {
			return nil
		}

		client, err := clientManager.Client(r.Context(), gi.ClientID)
		if err != nil {
			return fmt.Errorf("unable to fetch client %s: %w", gi.ClientID, err)
		}

		rawUserID := client.CustomAttribute("user_delegation_user_id")
		if rawUserID == nil {
			return nil
		}

		userID, ok := rawUserID.(uint64)
		if !ok {
			return fmt.Errorf("user_delegation_user_id is not a valid uint64 for client %s", gi.ClientID)
		}

		rawTenantID := client.CustomAttribute("user_delegation_tenant_id")
		if rawTenantID == nil {
			return fmt.Errorf("user_delegation_tenant_id is missing for client %s", gi.ClientID)
		}
		tenantID, ok := rawTenantID.(uint64)
		if !ok {
			return fmt.Errorf("user_delegation_tenant_id is not a valid uint64 for client %s", gi.ClientID)
		}

		user, err := userService.GetUser(r.Context(), user_service.GetUserReq{
			ID:       userID,
			TenantID: tenantID,
		})
		if err != nil {
			return fmt.Errorf("unable to fetch delegated user %d: %w", userID, err)
		}

		gi.Subject = fmt.Sprintf("%d", user.ID)

		var claims []string
		if raw := client.CustomAttribute("user_delegation_claims"); raw != nil {
			if c, ok := raw.([]string); ok {
				claims = c
			}
		}

		if gi.AdditionalTokenClaims == nil {
			gi.AdditionalTokenClaims = make(map[string]any)
		}

		if slices.Contains(claims, "preferred_username") {
			gi.AdditionalTokenClaims[goidc.ClaimPreferredUsername] = user.Username
		}
		if slices.Contains(claims, "email") {
			gi.AdditionalTokenClaims[goidc.ClaimEmail] = user.Email
		}
		if slices.Contains(claims, "name") {
			gi.AdditionalTokenClaims[goidc.ClaimName] = user.FirstName + " " + user.LastName
		}
		if slices.Contains(claims, "given_name") {
			gi.AdditionalTokenClaims[goidc.ClaimGivenName] = user.FirstName
		}
		if slices.Contains(claims, "family_name") {
			gi.AdditionalTokenClaims[goidc.ClaimFamilyName] = user.LastName
		}

		return nil
	}
}
