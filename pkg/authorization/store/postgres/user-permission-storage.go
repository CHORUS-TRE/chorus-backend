package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ authorization_service.UserPermissionStore = (*UserPermissionStorage)(nil)

type UserPermissionStorage struct {
	db                       *sqlx.DB
	rolesGrantingPermissions map[authorization_model.PermissionName][]authorization_model.RoleName
}

func NewUserPermissionStorage(db *sqlx.DB, rolesGrantingPermissions map[authorization_model.PermissionName][]authorization_model.RoleName) *UserPermissionStorage {
	return &UserPermissionStorage{
		db:                       db,
		rolesGrantingPermissions: rolesGrantingPermissions,
	}
}

func (s *UserPermissionStorage) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter authorization_model.FindUsersWithPermissionFilter) ([]uint64, error) {
	rolesGranting, ok := s.rolesGrantingPermissions[filter.PermissionName]
	if !ok || len(rolesGranting) == 0 {
		return nil, fmt.Errorf("no roles grant permission %s", filter.PermissionName)
	}

	rolesToCheck := make([]string, len(rolesGranting))
	for i, r := range rolesGranting {
		rolesToCheck[i] = r.String()
	}
	if len(filter.ViaRoles) > 0 {
		rolesToCheck = make([]string, 0)
		viaRolesSet := make(map[string]bool)
		for _, r := range filter.ViaRoles {
			viaRolesSet[string(r)] = true
		}
		for _, rg := range rolesGranting {
			if viaRolesSet[rg.String()] {
				rolesToCheck = append(rolesToCheck, rg.String())
			}
		}
		if len(rolesToCheck) == 0 {
			return nil, nil
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

func (s *UserPermissionStorage) findUsersWithExactContext(ctx context.Context, tenantID uint64, rolesToCheck []string, filterContext authorization_model.Context) ([]uint64, error) {
	contextDimensions := make([]string, 0, len(filterContext))
	contextValues := make([]string, 0, len(filterContext))
	for dim, val := range filterContext {
		contextDimensions = append(contextDimensions, string(dim))
		contextValues = append(contextValues, val)
	}

	query := `
SELECT DISTINCT u.id
FROM users u
JOIN user_role ur ON ur.userid = u.id
JOIN role_definitions rd ON rd.id = ur.roleid
JOIN user_role_context urc ON urc.userroleid = ur.id
WHERE u.tenantid = $1
  AND u.status = 'active'
  AND rd.name = ANY($2)
  AND urc.contextdimension = ANY($3)
  AND urc.value = ANY($4)
GROUP BY u.id, ur.id
HAVING COUNT(DISTINCT urc.contextdimension) = $5
`
	var userIDs []uint64
	err := s.db.SelectContext(ctx, &userIDs, query, tenantID, pq.Array(rolesToCheck), pq.Array(contextDimensions), pq.Array(contextValues), len(filterContext))
	if err != nil {
		return nil, fmt.Errorf("failed to find users with exact context: %w", err)
	}
	return userIDs, nil
}

func (s *UserPermissionStorage) findUsersWithContextMatch(ctx context.Context, tenantID uint64, rolesToCheck []string, filterContext authorization_model.Context) ([]uint64, error) {
	contextDimensions := make([]string, 0, len(filterContext))
	contextValues := make([]string, 0, len(filterContext))
	for dim, val := range filterContext {
		contextDimensions = append(contextDimensions, string(dim))
		contextValues = append(contextValues, val)
	}

	query := `
SELECT DISTINCT u.id
FROM users u
JOIN user_role ur ON ur.userid = u.id
JOIN role_definitions rd ON rd.id = ur.roleid
JOIN user_role_context urc ON urc.userroleid = ur.id
WHERE u.tenantid = $1
  AND u.status = 'active'
  AND rd.name = ANY($2)
  AND urc.contextdimension = ANY($3)
  AND (urc.value = ANY($4) OR urc.value = '*')
GROUP BY u.id, ur.id
HAVING COUNT(DISTINCT urc.contextdimension) = $5
`
	var userIDs []uint64
	err := s.db.SelectContext(ctx, &userIDs, query, tenantID, pq.Array(rolesToCheck), pq.Array(contextDimensions), pq.Array(contextValues), len(filterContext))
	if err != nil {
		return nil, fmt.Errorf("failed to find users with context match: %w", err)
	}
	return userIDs, nil
}
