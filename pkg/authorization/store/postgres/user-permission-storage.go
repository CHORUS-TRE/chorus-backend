package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
)

type UserPermissionStorage struct {
	db *sqlx.DB
}

func NewUserPermissionStorage(db *sqlx.DB) *UserPermissionStorage {
	return &UserPermissionStorage{db: db}
}

// FindUsersWithPermission returns user ids that hold the requested permission,
// computed from the provided list of roles known to grant it.
func (s *UserPermissionStorage) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter, rolesGranting []model.RoleName) ([]uint64, error) {
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

func (s *UserPermissionStorage) findUsersWithRoles(ctx context.Context, tenantID uint64, rolesToCheck []string) ([]uint64, error) {
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

func (s *UserPermissionStorage) findUsersWithExactContext(ctx context.Context, tenantID uint64, rolesToCheck []string, filterContext model.Context) ([]uint64, error) {
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

func (s *UserPermissionStorage) findUsersWithContextMatch(ctx context.Context, tenantID uint64, rolesToCheck []string, filterContext model.Context) ([]uint64, error) {
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
