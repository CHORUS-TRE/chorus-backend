package postgres

import (
	"context"
	"fmt"

	authorization_model "github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"

	"github.com/jmoiron/sqlx"
)

// AuthenticationStorage is the handler through which a PostgresDB
// backend can be queried.
type AuthenticationStorage struct {
	db *sqlx.DB
}

// NewAuthenticationStorage returns a fresh PostgresDB authentication storage instance.
func NewAuthenticationStorage(db *sqlx.DB) *AuthenticationStorage {
	return &AuthenticationStorage{db: db}
}

// GetActiveUser fetches a user entry from the database that matches the provided username.
func (s *AuthenticationStorage) GetActiveUser(ctx context.Context, username, source string) (*user_model.User, error) {
	const query = `
SELECT id, tenantid, firstname, lastname, username, source, password, totpsecret, totpenabled
FROM users
WHERE username = $1 AND source = $2 AND status = 'active';
`
	var u user_model.User
	if err := s.db.GetContext(ctx, &u, query, username, source); err != nil {
		return nil, err
	}

	roles, err := s.getRoles(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles[:]
	return &u, nil
}

// getRoles fetches all the roles of a given user.
func (s *AuthenticationStorage) getRoles(ctx context.Context, userID uint64) ([]user_model.UserRole, error) {
	const query = `
SELECT id, name, contextdimension, value
FROM (
  SELECT user_role.id, role_definitions.name, user_role_context.contextdimension, user_role_context.value
  FROM user_role
  JOIN role_definitions
  ON user_role.roleid = role_definitions.id
  LEFT JOIN user_role_context
  ON user_role.id = user_role_context.userroleid
  WHERE user_role.userid = $1
) AS subquery;
`

	var flatRoles []struct {
		ID               uint64  `db:"id"`
		Name             string  `db:"name"`
		ContextDimension *string `db:"contextdimension"`
		Value            *string `db:"value"`
	}
	if err := s.db.SelectContext(ctx, &flatRoles, query, userID); err != nil {
		return nil, fmt.Errorf("failed to fetch roles for user %d: %w", userID, err)
	}

	roleMap := make(map[uint64]map[string]string)
	roleNameMap := make(map[uint64]string)
	for _, fr := range flatRoles {
		roleNameMap[fr.ID] = fr.Name
		_, exists := roleMap[fr.ID]
		if !exists {
			roleMap[fr.ID] = make(map[string]string)
		}
		if fr.ContextDimension == nil || fr.Value == nil {
			continue
		}
		roleMap[fr.ID][*fr.ContextDimension] = *fr.Value
	}

	var roles []user_model.UserRole
	for n, m := range roleMap {
		roleName := roleNameMap[n]
		role, err := authorization_model.ToRole(roleName, m)
		if err != nil {
			return nil, fmt.Errorf("failed to parse role %s: %w", roleName, err)
		}
		roles = append(roles, user_model.UserRole{Role: role, ID: n})
	}
	return roles, nil
}
