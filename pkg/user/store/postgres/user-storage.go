package postgres

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"

	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// UserStorage is the handler through which a PostgresDB backend can be queried.
type UserStorage struct {
	db *sqlx.DB
}

// NewUserStorage returns a fresh user service storage instance.
func NewUserStorage(db *sqlx.DB) *UserStorage {
	return &UserStorage{db: db}
}

// GetUsers queries all stocked users that are not 'deleted'.
func (s *UserStorage) GetUsers(ctx context.Context, tenantID uint64) ([]*model.User, error) {
	const query = `
SELECT id, tenantid, firstname, lastname, username, status, createdat, updatedat
FROM users
WHERE tenantid = $1 AND status != 'deleted';
`
	var users []*model.User
	if err := s.db.SelectContext(ctx, &users, query, tenantID); err != nil {
		return nil, err
	}

	for _, u := range users {
		roles, err := s.getUserRoles(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		u.Roles = roles[:]
	}

	return users, nil
}

func (s *UserStorage) GetUser(ctx context.Context, tenantID uint64, userID uint64) (*model.User, error) {
	const query = `
		SELECT id, tenantid, firstname, lastname, username, status, password, passwordChanged,totpenabled, totpsecret, createdat, updatedat
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
		return errors.Wrap(err, "unable to delete recovery code")
	}
	return nil
}

func (s *UserStorage) UpdateUserWithRecoveryCodes(ctx context.Context, tenantID uint64, user *model.User, totpRecoveryCodes []string) (err error) {
	const deleteRecoveryCodesQuery = `
		DELETE FROM totp_recovery_codes WHERE tenantid = $1 AND userid = $2;
	`

	const insertRecoveryCodeQuery = `
		INSERT INTO totp_recovery_codes (tenantid, userid, code) VALUES ($1, $2, $3);
	`

	tx, txErr := s.db.Beginx()
	if txErr != nil {
		return txErr
	}

	defer func() {
		if txErr = tx.Rollback(); txErr != nil {
			err = errors.Wrap(err, txErr.Error())
		}
	}()

	if err = s.updateUserAndRoles(ctx, tx, tenantID, user); err != nil {
		return errors.Wrap(err, "unable to update user and roles")
	}

	if _, err = s.db.ExecContext(ctx, deleteRecoveryCodesQuery, tenantID, user.ID); err != nil {
		return errors.Wrap(err, "unable to delete recovery codes")
	}

	for _, rc := range totpRecoveryCodes {
		if _, err = tx.ExecContext(ctx, insertRecoveryCodeQuery, tenantID, user.ID, rc); err != nil {
			return errors.Wrap(err, "unable to insert recovery codes")
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "unable to commit")
	}

	return err
}

func (s *UserStorage) SoftDeleteUser(ctx context.Context, tenantID uint64, userID uint64) error {
	const query = `
		UPDATE users
		SET (status, username, updatedat) = ($3, concat(username, $4), NOW())
		WHERE tenantid = $1 AND id = $2;
	`

	rows, err := s.db.ExecContext(ctx, query, tenantID, userID, model.UserDeleted.String(), "-"+uuid.Next())
	if err != nil {
		return errors.Wrap(err, "unable to exec")
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "unable to get rows affected")
	}

	if affected == 0 {
		return database.ErrNoRowsDeleted
	}

	return nil
}

func (s *UserStorage) UpdateUser(ctx context.Context, tenantID uint64, user *model.User) (err error) {
	tx, txErr := s.db.Beginx()
	if txErr != nil {
		return txErr
	}

	defer func() {
		if txErr = tx.Rollback(); txErr != nil {
			err = errors.Wrap(err, txErr.Error())
		}
	}()

	err = s.updateUserAndRoles(ctx, tx, tenantID, user)
	if err != nil {
		return errors.Wrap(err, "unable to update")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "unable to commit")
	}

	return err
}

func (s *UserStorage) updateUserAndRoles(ctx context.Context, tx *sqlx.Tx, tenantID uint64, user *model.User) error {
	const userUpdateQuery = `
		UPDATE users
		SET firstname = $3, lastname = $4, username = $5, status = $6, password = $7, passwordChanged = $8, totpenabled = $9, totpsecret = $10, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2;
	`

	const deleteUserRolesQuery = `
		DELETE FROM user_role
		WHERE userid = $1;
	`

	const insertUserRoleQuery = `
		INSERT INTO user_role (userid, roleid) VALUES ($1, $2);
	`

	// Update User
	rows, err := tx.ExecContext(ctx, userUpdateQuery, tenantID, user.ID, user.FirstName, user.LastName, user.Username,
		user.Status, user.Password, user.PasswordChanged, user.TotpEnabled, user.TotpSecret)
	if err != nil {
		return errors.Wrap(err, "unable to exec")
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "unable to get rows affected")
	}

	if affected == 0 {
		return database.ErrNoRowsUpdated
	}

	roles, err := s.GetRoles(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get roles")
	}

	// Delete Old User Roles
	_, err = tx.ExecContext(ctx, deleteUserRolesQuery, user.ID)
	if err != nil {
		return errors.Wrap(err, "unable to delete old roles")
	}

	// Add new User Roles
	for _, ur := range user.Roles {
		found := false
		for _, r := range roles {
			if ur.String() == r.Name {
				if _, err = tx.ExecContext(ctx, insertUserRoleQuery, user.ID, r.ID); err != nil {
					return errors.Wrap(err, "unable to exec in")
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown user role: %v", ur)
		}
	}

	return nil
}

// CreateUser saves the provided user object in the database 'users' table.
func (s *UserStorage) CreateUser(ctx context.Context, tenantID uint64, user *model.User) (uint64, error) {
	const userQuery = `
INSERT INTO users (tenantid, firstname, lastname, username, password, passwordChanged, status,
                   totpsecret, createdat, updatedat)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()) RETURNING id;
	`
	const userRoleQuery = `
INSERT INTO user_role (userid, roleid) VALUES ($1, $2);
	`
	const recoveryCodeQuery = `
INSERT INTO totp_recovery_codes (tenantid, userid, code) VALUES ($1, $2, $3);
	`

	// We do not want to insert a user if the subsequent creation of
	// the user_role entries fails.
	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}

	var id uint64
	err = tx.GetContext(ctx, &id,
		userQuery, tenantID, user.FirstName, user.LastName, user.Username, user.Password, user.PasswordChanged, user.Status, user.TotpSecret,
	)
	if err != nil {
		return 0, storage.Rollback(tx, err)
	}

	roles, err := s.GetRoles(ctx)
	if err != nil {
		return 0, storage.Rollback(tx, err)
	}

	// For each user role that matches a role insert an entry in the 'user_role' table.
	var loopErr error
	for _, ur := range user.Roles {
		found := false
		for _, r := range roles {
			if ur.String() == r.Name {
				if _, loopErr = tx.ExecContext(ctx, userRoleQuery, id, r.ID); loopErr != nil {
					return 0, storage.Rollback(tx, loopErr)
				}
				found = true
				break
			}
		}
		if !found {
			loopErr = fmt.Errorf("unknown user role: %v", ur)
			return 0, storage.Rollback(tx, loopErr)
		}
	}
	// Insert TOTP recovery codes.
	if user.TotpRecoveryCodes != nil {
		for _, rc := range user.TotpRecoveryCodes {
			if _, loopErr = tx.ExecContext(ctx, recoveryCodeQuery, tenantID, id, rc); loopErr != nil {
				return 0, storage.Rollback(tx, loopErr)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// getUserRoles fetches all the roles of a given user.
func (s *UserStorage) getUserRoles(ctx context.Context, userID uint64) ([]model.UserRole, error) {
	const query = `
SELECT name
FROM (
  SELECT * FROM user_role
  JOIN roles
  ON user_role.userid = $1 AND user_role.roleid = roles.id
);
`
	var roles []model.UserRole
	if err := s.db.SelectContext(ctx, &roles, query, userID); err != nil {
		return nil, err
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
		return errors.Wrap(err, "unable to create role")
	}

	return nil
}
