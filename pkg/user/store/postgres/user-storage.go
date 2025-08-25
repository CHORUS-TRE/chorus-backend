package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	authorization_model "github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

// UserStorage is the handler through which a PostgresDB backend can be queried.
type UserStorage struct {
	db *sqlx.DB
}

// NewUserStorage returns a fresh user service storage instance.
func NewUserStorage(db *sqlx.DB) *UserStorage {
	return &UserStorage{db: db}
}

// ListUsers queries all stocked users that are not 'deleted'.
func (s *UserStorage) ListUsers(ctx context.Context, tenantID uint64, pagination *common.Pagination) (users []*model.User, paginationRes *common.PaginationResult, err error) {
	// Get total count query
	countQuery := `SELECT COUNT(*) FROM users WHERE tenantid = $1 AND status != 'deleted'`
	var totalCount int64
	if err = s.db.GetContext(ctx, &totalCount, countQuery, tenantID); err != nil {
		return nil, nil, err
	}

	// Get users query
	query := `
		SELECT id, tenantid, firstname, lastname, username, source, status, createdat, updatedat
		FROM users
		WHERE tenantid = $1 AND status != 'deleted'
	`

	// Add pagination
	clause, validatedPagination := storage.BuildPaginationClause(pagination, model.User{})
	query += clause

	// Build pagination result
	paginationRes = &common.PaginationResult{
		Total: uint64(totalCount),
	}

	if validatedPagination != nil {
		paginationRes.Limit = validatedPagination.Limit
		paginationRes.Offset = validatedPagination.Offset
		paginationRes.Sort = validatedPagination.Sort
	}

	args := []interface{}{tenantID}
	if err := s.db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, nil, err
	}

	for _, u := range users {
		roles, err := s.getUserRoles(ctx, u.ID)
		if err != nil {
			return nil, nil, err
		}
		u.Roles = roles
	}

	return users, paginationRes, nil
}

func (s *UserStorage) GetUser(ctx context.Context, tenantID uint64, userID uint64) (*model.User, error) {
	const query = `
		SELECT id, tenantid, firstname, lastname, username, source, status, password, passwordChanged,totpenabled, totpsecret, createdat, updatedat
		FROM users
		WHERE tenantid = $1 AND id = $2;
	`

	var user model.User
	if err := s.db.GetContext(ctx, &user, query, tenantID, userID); err != nil {
		return nil, err
	}

	roles, err := s.getUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return &user, nil
}

func (s *UserStorage) GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error) {
	const query = `
		SELECT id, tenantid, userid, code FROM totp_recovery_codes WHERE tenantid = $1 AND userid = $2;
	`
	var codes []*model.TotpRecoveryCode
	if err := s.db.Select(&codes, query, tenantID, userID); err != nil {
		return nil, err
	}

	return codes, nil
}

// DeleteTotpRecoveryCode removes a TOTP recovery code specified by codeId from the database.
func (s *UserStorage) DeleteTotpRecoveryCode(ctx context.Context, tenantID, codeID uint64) error {
	const query = `
DELETE FROM totp_recovery_codes WHERE tenantid = $1 AND id = $2;
`
	if _, err := s.db.ExecContext(ctx, query, tenantID, codeID); err != nil {
		return fmt.Errorf("unable to delete recovery code: %w", err)
	}
	return nil
}

func (s *UserStorage) UpdateUserWithRecoveryCodes(ctx context.Context, tenantID uint64, user *model.User, totpRecoveryCodes []string) (updatedUser *model.User, err error) {
	const deleteRecoveryCodesQuery = `
		DELETE FROM totp_recovery_codes WHERE tenantid = $1 AND userid = $2;
	`

	const insertRecoveryCodeQuery = `
		INSERT INTO totp_recovery_codes (tenantid, userid, code) VALUES ($1, $2, $3);
	`

	tx, txErr := s.db.Beginx()
	if txErr != nil {
		return nil, txErr
	}

	defer func() {
		if err != nil {
			if txErr = tx.Rollback(); txErr != nil {
				err = fmt.Errorf("%s: %w", txErr.Error(), err)
			}
		}
	}()

	updatedUser, err = s.updateUserAndRoles(ctx, tx, tenantID, user)
	if err != nil {
		return nil, fmt.Errorf("unable to update user and roles: %w", err)
	}

	if _, err = s.db.ExecContext(ctx, deleteRecoveryCodesQuery, tenantID, user.ID); err != nil {
		return nil, fmt.Errorf("unable to delete recovery codes: %w", err)
	}

	for _, rc := range totpRecoveryCodes {
		if _, err = tx.ExecContext(ctx, insertRecoveryCodeQuery, tenantID, user.ID, rc); err != nil {
			return nil, fmt.Errorf("unable to insert recovery codes: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("unable to commit: %w", err)
	}

	return updatedUser, err
}

func (s *UserStorage) SoftDeleteUser(ctx context.Context, tenantID uint64, userID uint64) error {
	const query = `
		UPDATE users
		SET (status, username, updatedat) = ($3, username || $4::text, NOW())
		WHERE tenantid = $1 AND id = $2;
	`
	uuidSuffix := "-" + uuid.Next()
	rows, err := s.db.ExecContext(ctx, query, tenantID, userID, model.UserDeleted.String(), uuidSuffix)
	if err != nil {
		return fmt.Errorf("unable to exec: %w", err)
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to get rows affected: %w", err)
	}

	if affected == 0 {
		return database.ErrNoRowsDeleted
	}

	return nil
}

func (s *UserStorage) UpdateUser(ctx context.Context, tenantID uint64, user *model.User) (updatedUser *model.User, err error) {
	tx, txErr := s.db.Beginx()
	if txErr != nil {
		return nil, txErr
	}

	defer func() {
		if err != nil {
			if txErr = tx.Rollback(); txErr != nil {
				err = fmt.Errorf("%s: %w", txErr.Error(), err)
			}
		}
	}()

	updatedUser, err = s.updateUserAndRoles(ctx, tx, tenantID, user)
	if err != nil {
		return nil, fmt.Errorf("unable to update: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("unable to commit: %w", err)
	}

	return updatedUser, err
}

func (s *UserStorage) updateUserAndRoles(ctx context.Context, tx *sqlx.Tx, tenantID uint64, user *model.User) (*model.User, error) {
	const userUpdateQuery = `
		UPDATE users
		SET firstname = $3, lastname = $4, username = $5, source = $6, status = $7, password = $8, passwordChanged = $9, totpenabled = $10, totpsecret = $11, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2
		RETURNING id, tenantid, firstname, lastname, username, source, status, passwordChanged, totpenabled, totpsecret, createdat, updatedat;
	`

	const deleteUserRolesQuery = `
		DELETE FROM user_role
		WHERE userid = $1;
	`

	// Update User
	var updatedUser model.User
	err := tx.GetContext(ctx, &updatedUser, userUpdateQuery, tenantID, user.ID, user.FirstName, user.LastName, user.Username, user.Source,
		user.Status, user.Password, user.PasswordChanged, user.TotpEnabled, user.TotpSecret)
	if err != nil {
		return nil, fmt.Errorf("unable to update user: %w", err)
	}

	// Delete Old User Roles
	_, err = tx.ExecContext(ctx, deleteUserRolesQuery, user.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to delete old roles: %w", err)
	}

	err = s.createUserRoles(ctx, tx, user.ID, user.Roles)
	if err != nil {
		return nil, fmt.Errorf("unable to create user roles: %w", err)
	}

	updatedUser.Roles = append(updatedUser.Roles, user.Roles...)

	// Set the roles on the updated user
	// updatedUser.Roles = userRoles

	return &updatedUser, nil
}

// CreateUser saves the provided user object in the database 'users' table.
func (s *UserStorage) CreateUser(ctx context.Context, tenantID uint64, user *model.User) (*model.User, error) {
	const userQuery = `
		INSERT INTO users (tenantid, firstname, lastname, username, source, password, passwordChanged, status, totpsecret, createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) 
		RETURNING id, tenantid, firstname, lastname, username, source, status, passwordChanged, totpenabled, totpsecret, createdat, updatedat;
	`
	const recoveryCodeQuery = `
		INSERT INTO totp_recovery_codes (tenantid, userid, code) VALUES ($1, $2, $3);
	`

	// We do not want to insert a user if the subsequent creation of
	// the user_role entries fails.
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}

	var newUser model.User
	err = tx.GetContext(ctx, &newUser,
		userQuery, tenantID, user.FirstName, user.LastName, user.Username, user.Source, user.Password, user.PasswordChanged, user.Status, user.TotpSecret,
	)
	if err != nil {
		return nil, storage.Rollback(tx, err)
	}

	err = s.createUserRoles(ctx, tx, newUser.ID, user.Roles)
	if err != nil {
		return nil, storage.Rollback(tx, err)
	}
	newUser.Roles = append(newUser.Roles, user.Roles...)

	// Insert TOTP recovery codes.
	if user.TotpRecoveryCodes != nil {
		for _, rc := range user.TotpRecoveryCodes {
			if _, loopErr := tx.ExecContext(ctx, recoveryCodeQuery, tenantID, newUser.ID, rc); loopErr != nil {
				return nil, storage.Rollback(tx, loopErr)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &newUser, nil
}

func (s *UserStorage) createUserRoles(ctx context.Context, tx *sqlx.Tx, userID uint64, userRoles []authorization_model.Role) error {
	const userRoleQuery = `
		INSERT INTO user_role (userid, roleid) VALUES ($1, $2) RETURNING id;
	`
	const userRoleContextQuery = `
		INSERT INTO user_role_context (userroleid, contextdimension, value) VALUES ($1, $2, $3);
	`

	roles, err := s.GetRoles(ctx)
	if err != nil {
		return fmt.Errorf("unable to get roles: %w", err)
	}

	// For each user role that matches a role insert an entry in the 'user_role' table.
	var loopErr error
	for _, ur := range userRoles {
		found := false
		for _, r := range roles {
			if ur.Name.String() == r.Name {
				var userRoleID uint64

				loopErr := tx.GetContext(ctx, &userRoleID, userRoleQuery, userID, r.ID)
				if loopErr != nil {
					return fmt.Errorf("unable to create user role: %w", loopErr)
				}

				for dimension, value := range ur.Context {
					if _, loopErr = tx.ExecContext(ctx, userRoleContextQuery, userRoleID, dimension, value); loopErr != nil {
						return fmt.Errorf("unable to create user role context: %w", loopErr)
					}
				}

				found = true
				break
			}
		}
		if !found {
			loopErr = fmt.Errorf("unknown user role: %v", ur)
			return fmt.Errorf("unable to create user role: %w", loopErr)
		}
	}

	return nil
}

type DBUserRoleContext struct {
	UserRoleID       uint64
	ContextDimension string
	Value            string
}

// getUserRoles fetches all the roles of a given user.
func (s *UserStorage) getUserRoles(ctx context.Context, userID uint64) ([]authorization_model.Role, error) {
	const query = `
SELECT user_role.id, roles.name
FROM (
  SELECT * FROM user_role
  JOIN roles
  ON user_role.userid = $1 AND user_role.roleid = roles.id
);
`
	var dbRoles []model.Role
	if err := s.db.SelectContext(ctx, &dbRoles, query, userID); err != nil {
		return nil, err
	}

	const dimensionsQuery = `
SELECT user_role.id AS user_role_id, contextdimension, value
FROM user_role_context
JOIN user_role
ON user_role.userid = $1 AND user_role.id = user_role_context.user_role_id;
`

	var dimensions []DBUserRoleContext
	if err := s.db.SelectContext(ctx, &dimensions, dimensionsQuery, userID); err != nil {
		return nil, err
	}

	roles := make([]authorization_model.Role, 0, len(dbRoles))
	for _, r := range dbRoles {
		roleName, err := authorization_model.ToRoleName(r.Name)
		if err != nil {
			return nil, err
		}
		role := authorization_model.Role{
			Name:    roleName,
			Context: authorization_model.Context{},
		}

		for _, d := range dimensions {
			if d.UserRoleID == r.ID {
				role.Context[authorization_model.ContextDimension(d.ContextDimension)] = d.Value
			}
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetRoles queries all stocked roles.
func (s *UserStorage) GetRoles(ctx context.Context) ([]*model.Role, error) {
	const query = `
SELECT id, name FROM roles;
	`
	var roles []*model.Role
	if err := s.db.SelectContext(ctx, &roles, query); err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *UserStorage) CreateRole(ctx context.Context, role string) error {
	const query = `
		insert into roles (name)
			select $1
		where not exists
			(select * from roles where name = $1)`

	_, err := s.db.ExecContext(ctx, query, role)
	if err != nil {
		return fmt.Errorf("unable to create role: %w", err)
	}

	return nil
}
