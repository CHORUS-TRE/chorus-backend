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

// findUsersBaseQuery selects active users in a tenant holding any of the given
// role names. Callers append context conditions as needed.
const findUsersBaseQuery = `
SELECT DISTINCT u.id
FROM users u
JOIN user_role ur ON ur.userid = u.id
JOIN role_definitions rd ON rd.id = ur.roleid
WHERE u.tenantid = $1
  AND u.status = 'active'
  AND rd.name = ANY($2)`

// FindUsersWithPermission returns user ids that hold the requested permission,
// computed from the provided list of roles known to grant it.
func (s *AuthorizationStorage) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter, rolesGranting []model.RoleName) ([]uint64, error) {
	if len(rolesGranting) == 0 {
		return nil, fmt.Errorf("no roles grant permission %s", filter.PermissionName)
	}

	roles := rolesToCheck(rolesGranting, filter.ViaRoles)
	if len(roles) == 0 {
		return nil, nil // the via-roles filter excluded every granting role
	}

	if len(filter.Context) == 0 {
		return s.findUsersWithRoles(ctx, tenantID, roles)
	}

	// Prefer an exact context match when asked; fall back to wildcard-tolerant.
	if filter.PreferExactContextMatch {
		userIDs, err := s.findUsersWithContext(ctx, tenantID, roles, filter.Context, false)
		if err != nil {
			return nil, err
		}
		if len(userIDs) > 0 {
			return userIDs, nil
		}
	}

	return s.findUsersWithContext(ctx, tenantID, roles, filter.Context, true)
}

// rolesToCheck reduces the granting roles to their names, optionally restricted
// to a caller-supplied subset (filter.ViaRoles).
func rolesToCheck(rolesGranting, viaRoles []model.RoleName) []string {
	if len(viaRoles) == 0 {
		names := make([]string, 0, len(rolesGranting))
		for _, r := range rolesGranting {
			names = append(names, r.String())
		}
		return names
	}

	allowed := make(map[string]bool, len(viaRoles))
	for _, r := range viaRoles {
		allowed[r.String()] = true
	}
	var names []string
	for _, r := range rolesGranting {
		if allowed[r.String()] {
			names = append(names, r.String())
		}
	}
	return names
}

func (s *AuthorizationStorage) findUsersWithRoles(ctx context.Context, tenantID uint64, roles []string) ([]uint64, error) {
	var userIDs []uint64
	if err := s.db.SelectContext(ctx, &userIDs, findUsersBaseQuery, tenantID, pq.Array(roles)); err != nil {
		return nil, fmt.Errorf("find users with roles: %w", err)
	}
	return userIDs, nil
}

// findUsersWithContext narrows findUsersBaseQuery to users whose role context
// matches every dimension in filterContext. When matchWildcard is true a stored
// "*" also satisfies a dimension.
func (s *AuthorizationStorage) findUsersWithContext(ctx context.Context, tenantID uint64, roles []string, filterContext model.Context, matchWildcard bool) ([]uint64, error) {
	args := []any{tenantID, pq.Array(roles)}

	conditions := make([]string, 0, len(filterContext))
	for dimension, value := range filterContext {
		dimPlaceholder, valuePlaceholder := len(args)+1, len(args)+2
		valueMatch := fmt.Sprintf("urc.value = $%d", valuePlaceholder)
		if matchWildcard {
			valueMatch = fmt.Sprintf("(urc.value = $%d OR urc.value = '*')", valuePlaceholder)
		}
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM user_role_context urc WHERE urc.userroleid = ur.id AND urc.contextdimension = $%d AND %s)",
			dimPlaceholder, valueMatch,
		))
		args = append(args, string(dimension), value)
	}

	query := findUsersBaseQuery + "\n  AND " + strings.Join(conditions, "\n  AND ")

	var userIDs []uint64
	if err := s.db.SelectContext(ctx, &userIDs, query, args...); err != nil {
		return nil, fmt.Errorf("find users with context: %w", err)
	}
	return userIDs, nil
}
