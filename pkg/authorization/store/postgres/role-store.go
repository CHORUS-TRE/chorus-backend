package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ service.Store = (*RoleStorage)(nil)

type RoleStorage struct {
	db *sqlx.DB
	*UserPermissionStorage
}

func NewRoleStorage(db *sqlx.DB) *RoleStorage {
	return &RoleStorage{
		db:                    db,
		UserPermissionStorage: NewUserPermissionStorage(db),
	}
}

// SyncSystemRoles atomically replaces the set of non-dynamic role definitions
// with the provided ones. Dynamic roles are left untouched.
func (s *RoleStorage) SyncSystemRoles(ctx context.Context, roles []*model.RoleDefinition) error {
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

func (s *RoleStorage) CreateDynamicRole(ctx context.Context, role *model.RoleDefinition) error {
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

func (s *RoleStorage) ListRoles(ctx context.Context) ([]*model.RoleDefinition, error) {
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
		dimension, err := model.ToRoleContext(row.Dimension)
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
