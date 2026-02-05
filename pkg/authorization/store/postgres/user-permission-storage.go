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

type userPermissionRow struct {
	UserID           uint64  `db:"userid"`
	RoleName         string  `db:"rolename"`
	ContextDimension *string `db:"contextdimension"`
	ContextValue     *string `db:"contextvalue"`
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

	var args []interface{}
	args = append(args, tenantID)
	args = append(args, pq.Array(rolesToCheck))

	query := `
SELECT 
    u.id AS userid,
    rd.name AS rolename,
    urc.contextdimension,
    urc.value AS contextvalue
FROM users u
JOIN user_role ur ON ur.userid = u.id
JOIN role_definitions rd ON rd.id = ur.roleid
LEFT JOIN user_role_context urc ON urc.userroleid = ur.id
WHERE u.tenantid = $1
  AND u.status = 'active'
  AND rd.name = ANY($2)
`
	var rows []userPermissionRow
	if err := s.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("failed to find users with permission: %w", err)
	}

	type userRoleInfo struct {
		RoleName string
		Context  map[string]string
	}

	userRoles := make(map[uint64][]userRoleInfo)
	tempRoleCtx := make(map[uint64]map[string]map[string]string)

	for _, row := range rows {
		if _, ok := tempRoleCtx[row.UserID]; !ok {
			tempRoleCtx[row.UserID] = make(map[string]map[string]string)
		}
		if _, ok := tempRoleCtx[row.UserID][row.RoleName]; !ok {
			tempRoleCtx[row.UserID][row.RoleName] = make(map[string]string)
		}
		if row.ContextDimension != nil && row.ContextValue != nil {
			tempRoleCtx[row.UserID][row.RoleName][*row.ContextDimension] = *row.ContextValue
		}
	}

	for userID, roleMap := range tempRoleCtx {
		for roleName, ctxMap := range roleMap {
			userRoles[userID] = append(userRoles[userID], userRoleInfo{
				RoleName: roleName,
				Context:  ctxMap,
			})
		}
	}

	var exactMatchUsers []uint64
	var wildcardMatchUsers []uint64

	for userID, roles := range userRoles {
		hasExactMatch := false
		hasWildcardMatch := false

		for _, role := range roles {
			matchType := s.matchContext(filter.Context, role.Context)
			if matchType == "exact" {
				hasExactMatch = true
			} else if matchType == "wildcard" {
				hasWildcardMatch = true
			}
		}

		if hasExactMatch {
			exactMatchUsers = append(exactMatchUsers, userID)
		} else if hasWildcardMatch {
			wildcardMatchUsers = append(wildcardMatchUsers, userID)
		}
	}

	if filter.PreferExactContextMatch && len(exactMatchUsers) > 0 {
		return exactMatchUsers, nil
	}

	result := append(exactMatchUsers, wildcardMatchUsers...)
	return s.uniqueUserIDs(result), nil
}

func (s *UserPermissionStorage) matchContext(requiredCtx authorization_model.Context, roleCtx map[string]string) string {
	if len(requiredCtx) == 0 {
		return "exact"
	}

	hasWildcard := false
	for dim, requiredVal := range requiredCtx {
		roleVal, ok := roleCtx[string(dim)]
		if !ok {
			return "none"
		}
		if roleVal == "*" {
			hasWildcard = true
		} else if roleVal != requiredVal {
			return "none"
		}
	}

	if hasWildcard {
		return "wildcard"
	}
	return "exact"
}

func (s *UserPermissionStorage) uniqueUserIDs(ids []uint64) []uint64 {
	seen := make(map[uint64]bool)
	result := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}
