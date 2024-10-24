package postgres

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/model"

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
func (s *AuthenticationStorage) GetActiveUser(ctx context.Context, username, source string) (*model.User, error) {
	const query = `
SELECT id, tenantid, firstname, lastname, username, source, password, totpsecret, totpenabled
FROM users
WHERE username = $1 AND source = $2 AND status = 'active';
`
	var u model.User
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
func (s *AuthenticationStorage) getRoles(ctx context.Context, userID uint64) ([]string, error) {
	const query = `
SELECT name
FROM (
  SELECT roles.name
  FROM user_role
  JOIN roles
  ON user_role.roleid = roles.id
  WHERE user_role.userid = $1
) AS subquery;
`
	var roles []string
	if err := s.db.SelectContext(ctx, &roles, query, userID); err != nil {
		return nil, err
	}
	return roles, nil
}
