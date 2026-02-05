package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

var _ authorization_service.UserRoleStore = (*UserRoleStorage)(nil)

// UserRoleStorage is the handler through which a PostgresDB backend can be queried.
type UserRoleStorage struct {
	db *sqlx.DB
}

// NewUserRoleStorage returns a fresh user service storage instance.
func NewUserRoleStorage(db *sqlx.DB) *UserRoleStorage {
	return &UserRoleStorage{db: db}
}

// GetRoles queries all stocked roles.
func (s *UserRoleStorage) GetRoles(ctx context.Context) ([]*model.Role, error) {
	const query = `
SELECT id, name FROM role_definitions;
	`
	var roles []*model.Role
	if err := s.db.SelectContext(ctx, &roles, query); err != nil {
		return nil, err
	}
	return roles, nil
}
