package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	gatekeeper_service "github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/service"
)

var _ chorus.AuthorizationServiceServer = (*AuthorizationController)(nil)

type AuthorizationController struct {
	gatekeeper gatekeeper_service.AuthorizationServiceInterface
}

func NewAuthorizationController(gatekeeper gatekeeper_service.AuthorizationServiceInterface) chorus.AuthorizationServiceServer {
	return &AuthorizationController{
		gatekeeper: gatekeeper,
	}
}

func (c AuthorizationController) ListRoles(ctx context.Context, req *chorus.ListRolesRequest) (*chorus.ListRolesReply, error) {
	schema := c.gatekeeper.GetAuthorizationSchema()
	if schema == nil {
		return &chorus.ListRolesReply{
			Result: &chorus.ListRolesResult{
				Roles: []*chorus.AuthorizationRole{},
			},
		}, nil
	}

	roles := make([]*chorus.AuthorizationRole, 0, len(schema.Roles))
	for _, role := range schema.Roles {
		contextDimensions := make([]string, 0, len(role.Attributes))
		for dim := range role.Attributes {
			contextDimensions = append(contextDimensions, string(dim))
		}

		roles = append(roles, &chorus.AuthorizationRole{
			Name:        role.Name,
			Description: role.Description,
			Context:     contextDimensions,
		})
	}

	return &chorus.ListRolesReply{
		Result: &chorus.ListRolesResult{
			Roles: roles,
		},
	}, nil
}

func (c AuthorizationController) ListPermissions(ctx context.Context, req *chorus.ListPermissionsRequest) (*chorus.ListPermissionsReply, error) {
	schema := c.gatekeeper.GetAuthorizationSchema()
	if schema == nil {
		return &chorus.ListPermissionsReply{
			Result: &chorus.ListPermissionsResult{
				Permissions: []*chorus.AuthorizationPermission{},
			},
		}, nil
	}

	permissions := make([]*chorus.AuthorizationPermission, 0, len(schema.Permissions))
	for _, perm := range schema.Permissions {
		contextDimensions := make([]string, 0, len(perm.Context))
		for dim := range perm.Context {
			contextDimensions = append(contextDimensions, string(dim))
		}

		permissions = append(permissions, &chorus.AuthorizationPermission{
			Name:        perm.Name,
			Description: perm.Description,
			Context:     contextDimensions,
		})
	}

	return &chorus.ListPermissionsReply{
		Result: &chorus.ListPermissionsResult{
			Permissions: permissions,
		},
	}, nil
}
