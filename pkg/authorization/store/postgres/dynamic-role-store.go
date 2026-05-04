package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ authorization_service.DynamicRoleStore = (*DynamicRoleStorage)(nil)

type DynamicRoleStorage struct {
	db *sqlx.DB
}

func NewDynamicRoleStorage(db *sqlx.DB) *DynamicRoleStorage {
	return &DynamicRoleStorage{db: db}
}

func (s *DynamicRoleStorage) ListDynamicRoles(ctx context.Context) ([]*authorization_model.RoleDefinition, error) {
	const rolesQuery = `
SELECT id, name, COALESCE(description, '') AS description, scope
FROM role_definitions
WHERE dynamic = true
ORDER BY name;
`
	var dbRoles []struct {
		ID          uint64 `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
		Scope       string `db:"scope"`
	}
	if err := s.db.SelectContext(ctx, &dbRoles, rolesQuery); err != nil {
		return nil, fmt.Errorf("unable to list dynamic roles: %w", err)
	}

	roles := make([]*authorization_model.RoleDefinition, 0, len(dbRoles))
	for _, dbRole := range dbRoles {
		scope, err := authorization_model.ToRoleScope(dbRole.Scope)
		if err != nil {
			return nil, err
		}
		permissions, err := s.listRolePermissions(ctx, dbRole.ID)
		if err != nil {
			return nil, err
		}

		roles = append(roles, &authorization_model.RoleDefinition{
			Name:                      authorization_model.RoleName(dbRole.Name),
			Description:               dbRole.Description,
			Scope:                     scope,
			Dynamic:                   true,
			RequiredContextDimensions: requiredContextForScope(scope),
			Permissions:               permissions,
		})
	}

	return roles, nil
}

func (s *DynamicRoleStorage) CreateDynamicRole(ctx context.Context, role *authorization_model.RoleDefinition) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer tx.Rollback()

	const roleQuery = `
INSERT INTO role_definitions (name, description, scope, dynamic)
VALUES ($1, $2, $3, true)
RETURNING id;
`
	var roleID uint64
	if err := tx.GetContext(ctx, &roleID, roleQuery, role.Name.String(), role.Description, role.Scope.String()); err != nil {
		return fmt.Errorf("unable to create dynamic role: %w", err)
	}

	const permissionQuery = `
INSERT INTO dynamic_role_permissions (roledefinitionid, permissionname)
VALUES ($1, $2);
`
	for _, permission := range role.Permissions {
		if _, err := tx.ExecContext(ctx, permissionQuery, roleID, permission.String()); err != nil {
			return fmt.Errorf("unable to create dynamic role permission: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("unable to commit dynamic role: %w", err)
	}
	return nil
}

func (s *DynamicRoleStorage) listRolePermissions(ctx context.Context, roleID uint64) ([]authorization_model.PermissionName, error) {
	const query = `
SELECT permissionname
FROM dynamic_role_permissions
WHERE roledefinitionid = $1
ORDER BY permissionname;
`
	var permissionNames []string
	if err := s.db.SelectContext(ctx, &permissionNames, query, roleID); err != nil {
		return nil, fmt.Errorf("unable to list dynamic role permissions: %w", err)
	}

	permissions := make([]authorization_model.PermissionName, 0, len(permissionNames))
	for _, permissionName := range permissionNames {
		permission, err := authorization_model.ToPermissionName(permissionName)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

func requiredContextForScope(scope authorization_model.RoleScope) map[authorization_model.ContextDimension]authorization_model.ContextQuantifier {
	switch scope {
	case authorization_model.RoleScopeWorkspace:
		return map[authorization_model.ContextDimension]authorization_model.ContextQuantifier{
			authorization_model.RoleContextWorkspace: authorization_model.ContextQuantifierOne,
		}
	case authorization_model.RoleScopePlatform:
		return nil
	default:
		return nil
	}
}
