package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ chorus.AuthorizationServiceServer = (*AuthorizationController)(nil)

type AuthorizationController struct {
	authorization authorization_service.Authorizer
}

func NewAuthorizationController(authorization authorization_service.Authorizer) chorus.AuthorizationServiceServer {
	return &AuthorizationController{
		authorization: authorization,
	}
}

func (c AuthorizationController) ListRoles(ctx context.Context, req *chorus.ListRolesRequest) (*chorus.ListRolesReply, error) {
	schema := c.authorization.GetAuthorizationSchema()
	if schema == nil {
		return &chorus.ListRolesReply{
			Result: &chorus.ListRolesResult{
				Roles: []*chorus.AuthorizationRole{},
			},
		}, nil
	}

	roles := make([]*chorus.AuthorizationRole, 0, len(schema.Roles))
	for _, role := range schema.Roles {
		contextDimensions := make([]string, 0, len(role.RequiredContextDimensions))
		for dim := range role.RequiredContextDimensions {
			contextDimensions = append(contextDimensions, string(dim))
		}

		roles = append(roles, &chorus.AuthorizationRole{
			Name:        role.Name.String(),
			Description: role.Description,
			Context:     contextDimensions,
			Permissions: converter.PermissionNames(role.Permissions),
			Scope:       role.Scope.String(),
			Dynamic:     role.Dynamic,
		})
	}

	return &chorus.ListRolesReply{
		Result: &chorus.ListRolesResult{
			Roles: roles,
		},
	}, nil
}

func (c AuthorizationController) CreateDynamicRole(ctx context.Context, req *chorus.CreateDynamicRoleRequest) (*chorus.CreateDynamicRoleReply, error) {
	if req == nil || req.Role == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	callerRoles, err := extractRolesFromContext(ctx)
	if err != nil {
		return nil, err
	}

	role, err := converter.DynamicRoleToBusiness(req.Role)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.Wrap(err, "Invalid dynamic role")
	}

	validationContext, err := converter.ContextToBusiness(req.ValidationContext)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.Wrap(err, "Invalid validation context")
	}

	createdRole, err := c.authorization.CreateDynamicRole(ctx, callerRoles, role, validationContext)
	if err != nil {
		return nil, err
	}

	return &chorus.CreateDynamicRoleReply{Result: &chorus.CreateDynamicRoleResult{Role: converter.RoleFromBusiness(createdRole)}}, nil
}

func extractRolesFromContext(ctx context.Context) ([]authorization_model.Role, error) {
	claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims)
	if !ok {
		return nil, cerr.ErrInvalidRequest.WithMessage("Malformed JWT token")
	}

	roles := make([]authorization_model.Role, 0, len(claims.Roles))
	for _, claimRole := range claims.Roles {
		role, err := authorization_model.ToRole(claimRole.Name, claimRole.Context)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (c AuthorizationController) ListPermissions(ctx context.Context, req *chorus.ListPermissionsRequest) (*chorus.ListPermissionsReply, error) {
	schema := c.authorization.GetAuthorizationSchema()
	if schema == nil {
		return &chorus.ListPermissionsReply{
			Result: &chorus.ListPermissionsResult{
				Permissions: []*chorus.AuthorizationPermission{},
			},
		}, nil
	}

	permissions := make([]*chorus.AuthorizationPermission, 0, len(schema.Permissions))
	for _, perm := range schema.Permissions {
		contextDimensions := make([]string, 0, len(perm.RequiredContextDimensions))
		for _, dim := range perm.RequiredContextDimensions {
			contextDimensions = append(contextDimensions, string(dim))
		}

		permissions = append(permissions, &chorus.AuthorizationPermission{
			Name:        perm.Name.String(),
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
