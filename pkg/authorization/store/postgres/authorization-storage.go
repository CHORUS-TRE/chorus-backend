package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var _ service.AuthorizationStore = (*AuthorizationStorage)(nil)

type AuthorizationStorage struct {
	db *sqlx.DB
}

func NewAuthorizationStorage(db *sqlx.DB) *AuthorizationStorage {
	return &AuthorizationStorage{db: db}
}

// SyncSystemRoles atomically replaces the set of non-dynamic role definitions
// with the provided ones. Dynamic roles are left untouched.
func (s *AuthorizationStorage) SyncSystemRoles(ctx context.Context, roles []*model.RoleDefinition) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin sync transaction: %w", err)
	}
	defer tx.Rollback()

	names := make([]string, 0, len(roles))
	for _, role := range roles {
		names = append(names, role.Name.String())
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM role_definitions WHERE dynamic = false AND name <> ALL($1)`,
		pq.Array(names),
	); err != nil {
		return fmt.Errorf("prune stale system roles: %w", err)
	}

	for _, role := range roles {
		if err := upsertRole(ctx, tx, role, false); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *AuthorizationStorage) CreateDynamicRole(ctx context.Context, role *model.RoleDefinition) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()
	if err := upsertRole(ctx, tx, role, true); err != nil {
		return err
	}
	return tx.Commit()
}

func upsertRole(ctx context.Context, tx *sqlx.Tx, role *model.RoleDefinition, dynamic bool) error {
	var roleID uint64
	err := tx.GetContext(ctx, &roleID, `
INSERT INTO role_definitions (name, description, scope, dynamic)
VALUES ($1, $2, $3, $4)
ON CONFLICT (name) DO UPDATE
SET description = EXCLUDED.description, scope = EXCLUDED.scope, dynamic = EXCLUDED.dynamic
RETURNING id`, role.Name.String(), role.Description, role.Scope.String(), dynamic)
	if err != nil {
		return fmt.Errorf("upsert role %s: %w", role.Name, err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM role_definition_permissions WHERE roledefinitionid = $1`, roleID); err != nil {
		return fmt.Errorf("clear permissions for %s: %w", role.Name, err)
	}
	for _, permission := range role.Permissions {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO role_definition_permissions (roledefinitionid, permissionname) VALUES ($1, $2)`,
			roleID, permission.String()); err != nil {
			return fmt.Errorf("insert permission %s for role %s: %w", permission, role.Name, err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM role_definitions_required_contexts WHERE roledefinitionid = $1`, roleID); err != nil {
		return fmt.Errorf("clear required contexts for %s: %w", role.Name, err)
	}
	for dimension, quantifier := range role.RequiredContextDimensions {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO role_definitions_required_contexts (roledefinitionid, contextdimension, contextquantifier)
VALUES ($1, $2, $3)`, roleID, string(dimension), string(quantifier)); err != nil {
			return fmt.Errorf("insert required context %s for role %s: %w", dimension, role.Name, err)
		}
	}
	return nil
}

func (s *AuthorizationStorage) ListRoles(ctx context.Context) ([]*model.RoleDefinition, error) {
	var roleRows []struct {
		ID          uint64 `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
		Scope       string `db:"scope"`
		Dynamic     bool   `db:"dynamic"`
	}
	if err := s.db.SelectContext(ctx, &roleRows, `
SELECT id, name, COALESCE(description, '') AS description, scope, dynamic
FROM role_definitions
ORDER BY name`); err != nil {
		return nil, fmt.Errorf("list role_definitions: %w", err)
	}

	var permRows []struct {
		RoleID         uint64 `db:"roledefinitionid"`
		PermissionName string `db:"permissionname"`
	}
	if err := s.db.SelectContext(ctx, &permRows,
		`SELECT roledefinitionid, permissionname FROM role_definition_permissions`); err != nil {
		return nil, fmt.Errorf("list role permissions: %w", err)
	}
	permissionsByRole := map[uint64][]model.PermissionName{}
	for _, row := range permRows {
		name, err := model.ToPermissionName(row.PermissionName)
		if err != nil {
			return nil, fmt.Errorf("role %d references unknown permission %q: %w", row.RoleID, row.PermissionName, err)
		}
		permissionsByRole[row.RoleID] = append(permissionsByRole[row.RoleID], name)
	}

	var ctxRows []struct {
		RoleID     uint64 `db:"roledefinitionid"`
		Dimension  string `db:"contextdimension"`
		Quantifier string `db:"contextquantifier"`
	}
	if err := s.db.SelectContext(ctx, &ctxRows, `
SELECT roledefinitionid, contextdimension, contextquantifier
FROM role_definitions_required_contexts`); err != nil {
		return nil, fmt.Errorf("list role required contexts: %w", err)
	}
	contextsByRole := map[uint64]map[model.ContextDimension]model.ContextQuantifier{}
	for _, row := range ctxRows {
		dimension, err := model.ToContextDimension(row.Dimension)
		if err != nil {
			return nil, err
		}
		if _, ok := contextsByRole[row.RoleID]; !ok {
			contextsByRole[row.RoleID] = map[model.ContextDimension]model.ContextQuantifier{}
		}
		contextsByRole[row.RoleID][dimension] = model.ContextQuantifier(row.Quantifier)
	}

	roles := make([]*model.RoleDefinition, 0, len(roleRows))
	for _, row := range roleRows {
		scope, err := model.ToRoleScope(row.Scope)
		if err != nil {
			return nil, err
		}
		roleName, err := model.ToRoleName(row.Name)
		if err != nil {
			return nil, err
		}
		roles = append(roles, &model.RoleDefinition{
			Name:                      roleName,
			Description:               row.Description,
			Scope:                     scope,
			Dynamic:                   row.Dynamic,
			Permissions:               permissionsByRole[row.ID],
			RequiredContextDimensions: contextsByRole[row.ID],
		})
	}
	return roles, nil
}

// FindUsersWithPermission returns user ids that hold the requested permission,
// computed from the provided list of roles known to grant it.
func (s *AuthorizationStorage) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter, rolesGranting []model.RoleName) ([]uint64, error) {
	if len(rolesGranting) == 0 {
		return nil, fmt.Errorf("no roles grant permission %s", filter.PermissionName)
	}

	var rolesToCheck []string
	if len(filter.ViaRoles) > 0 {
		viaRolesSet := make(map[string]bool, len(filter.ViaRoles))
		for _, r := range filter.ViaRoles {
			viaRolesSet[string(r)] = true
		}
		for _, r := range rolesGranting {
			if viaRolesSet[r.String()] {
				rolesToCheck = append(rolesToCheck, r.String())
			}
		}
		if len(rolesToCheck) == 0 {
			return nil, nil
		}
	} else {
		rolesToCheck = make([]string, 0, len(rolesGranting))
		for _, r := range rolesGranting {
			rolesToCheck = append(rolesToCheck, r.String())
		}
	}

	if len(filter.Context) == 0 {
		return s.findUsersWithRoles(ctx, tenantID, rolesToCheck)
	}

	if filter.PreferExactContextMatch {
		userIDs, err := s.findUsersWithExactContext(ctx, tenantID, rolesToCheck, filter.Context)
		if err != nil {
			return nil, err
		}
		if len(userIDs) > 0 {
			return userIDs, nil
		}
	}

	return s.findUsersWithContextMatch(ctx, tenantID, rolesToCheck, filter.Context)
}

func (s *AuthorizationStorage) findUsersWithRoles(ctx context.Context, tenantID uint64, rolesToCheck []string) ([]uint64, error) {
	query := `
SELECT DISTINCT u.id
FROM users u
JOIN user_role ur ON ur.userid = u.id
JOIN role_definitions rd ON rd.id = ur.roleid
WHERE u.tenantid = $1
  AND u.status = 'active'
  AND rd.name = ANY($2)
`
	var userIDs []uint64
	if err := s.db.SelectContext(ctx, &userIDs, query, tenantID, pq.Array(rolesToCheck)); err != nil {
		return nil, fmt.Errorf("failed to find users with roles: %w", err)
	}
	return userIDs, nil
}

func (s *AuthorizationStorage) findUsersWithExactContext(ctx context.Context, tenantID uint64, rolesToCheck []string, filterContext model.Context) ([]uint64, error) {
	args := []interface{}{tenantID, pq.Array(rolesToCheck)}

	conditions := make([]string, 0, len(filterContext))
	for dim, val := range filterContext {
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM user_role_context urc WHERE urc.userroleid = ur.id AND urc.contextdimension = $%d AND urc.value = $%d)",
			len(args)+1, len(args)+2,
		))
		args = append(args, string(dim), val)
	}

	query := fmt.Sprintf(`
SELECT DISTINCT u.id
FROM users u
JOIN user_role ur ON ur.userid = u.id
JOIN role_definitions rd ON rd.id = ur.roleid
WHERE u.tenantid = $1
  AND u.status = 'active'
  AND rd.name = ANY($2)
  AND %s
`, strings.Join(conditions, " AND "))

	var userIDs []uint64
	if err := s.db.SelectContext(ctx, &userIDs, query, args...); err != nil {
		return nil, fmt.Errorf("failed to find users with exact context: %w", err)
	}
	return userIDs, nil
}

func (s *AuthorizationStorage) findUsersWithContextMatch(ctx context.Context, tenantID uint64, rolesToCheck []string, filterContext model.Context) ([]uint64, error) {
	args := []interface{}{tenantID, pq.Array(rolesToCheck)}

	conditions := make([]string, 0, len(filterContext))
	for dim, val := range filterContext {
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM user_role_context urc WHERE urc.userroleid = ur.id AND urc.contextdimension = $%d AND (urc.value = $%d OR urc.value = '*'))",
			len(args)+1, len(args)+2,
		))
		args = append(args, string(dim), val)
	}

	query := fmt.Sprintf(`
SELECT DISTINCT u.id
FROM users u
JOIN user_role ur ON ur.userid = u.id
JOIN role_definitions rd ON rd.id = ur.roleid
WHERE u.tenantid = $1
  AND u.status = 'active'
  AND rd.name = ANY($2)
  AND %s
`, strings.Join(conditions, " AND "))

	var userIDs []uint64
	if err := s.db.SelectContext(ctx, &userIDs, query, args...); err != nil {
		return nil, fmt.Errorf("failed to find users with context match: %w", err)
	}
	return userIDs, nil
}
