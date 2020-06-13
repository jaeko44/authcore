package user

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"authcore.io/authcore/internal/authn/verifier"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/validator"
	"authcore.io/authcore/pkg/messageencryptor"
	"authcore.io/authcore/pkg/nulls"
	"authcore.io/authcore/pkg/paging"

	"github.com/go-redis/redis"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

var userStruct = sqlbuilder.NewStruct(new(User))

// Store manages User, Contact, and Role models.
type Store struct {
	db              *db.DB
	redis           *redis.Client
	encryptor       *messageencryptor.MessageEncryptor
	verifierFactory *verifier.Factory
}

// NewStore retrusn a new Store instance.
func NewStore(db *db.DB, redis *redis.Client, encryptor *messageencryptor.MessageEncryptor) *Store {
	store := &Store{
		db:              db,
		redis:           redis,
		encryptor:       encryptor,
		verifierFactory: verifier.NewFactory(),
	}
	return store
}

// InsertUser inserts the User and refresh the struct with data from database.
func (s *Store) InsertUser(ctx context.Context, user *User) error {
	if err := s.BeforeInsert(user); err != nil {
		return err
	}

	ib := userStruct.InsertIntoForTag("users", "insert", user)
	iq, ia := ib.Build()
	result, err := s.db.ExecContext(ctx, iq, ia...)
	if err != nil {
		return errors.WithSQLError(err)
	}

	user.ID, err = result.LastInsertId()
	if err != nil {
		return errors.WithSQLError(err)
	}

	return s.SelectUser(ctx, user)
}

// SelectUser selects a User with the given ID.
func (s *Store) SelectUser(ctx context.Context, user *User) error {
	sb := userStruct.SelectFrom("users")
	sb.Where(sb.E("id", user.ID))
	sq, sa := sb.Build()
	return selectUser(ctx, s.db, s, user, sq, sa...)
}

// UpdateUser updates a User with the given ID.
func (s *Store) UpdateUser(ctx context.Context, user *User) error {
	if err := s.BeforeUpdate(user); err != nil {
		return err
	}
	ub := userStruct.UpdateForTag("users", "update", user)
	ub.Where(ub.E("id", user.ID))
	uq, ua := ub.Build()
	_, err := s.db.ExecContext(ctx, uq, ua...)
	return errors.WithSQLError(err)
}

// UpdateUserLastSeenAt updates the last seen at for the given user.
func (s *Store) UpdateUserLastSeenAt(ctx context.Context, id int64, lastSeenAt time.Time) error {
	_, err := s.db.ExecContext(ctx, "UPDATE users SET last_seen_at = ? WHERE id = ?", lastSeenAt, id)
	return errors.WithSQLError(err)
}

// UserByID lookups a User by ID. This function is depreciated by SelectUser.
func (s *Store) UserByID(ctx context.Context, id int64) (*User, error) {
	user := &User{ID: id}
	err := s.SelectUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UserByPublicID lookups a User a public ID returned by User's PublicID.
func (s *Store) UserByPublicID(ctx context.Context, publicID string) (*User, error) {
	id, err := strconv.ParseInt(publicID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	user := &User{ID: id}
	err = s.SelectUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UserByEmail lookups a User by email.
func (s *Store) UserByEmail(ctx context.Context, email string) (*User, error) {
	sb := userStruct.SelectFrom("users")
	sb.Where(sb.E("email", email))
	sq, sa := sb.Build()
	user := &User{}
	if err := selectUser(ctx, s.db, s, user, sq, sa...); err != nil {
		return nil, err
	}
	return user, nil
}

// UserByPhone lookups a User by phone number.
func (s *Store) UserByPhone(ctx context.Context, phone string) (*User, error) {
	sb := userStruct.SelectFrom("users")
	sb.Where(sb.E("phone", phone))
	sq, sa := sb.Build()
	user := &User{}
	if err := selectUser(ctx, s.db, s, user, sq, sa...); err != nil {
		return nil, err
	}
	return user, nil
}

// UserByHandle lookups a User by email or phone number.
func (s *Store) UserByHandle(ctx context.Context, handle string) (*User, error) {
	sb := userStruct.SelectFrom("users")
	sb.Where(sb.Or(sb.E("email", handle), sb.E("phone", handle), sb.E("username", handle)))
	sq, sa := sb.Build()
	user := &User{}
	if err := selectUser(ctx, s.db, s, user, sq, sa...); err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUserByID deletes a user and its associated record in database.
func (s *Store) DeleteUserByID(ctx context.Context, id int64) error {
	return s.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		user := &User{ID: id}
		// FIXME: SelectUser is not using in-progress transaction thread
		err := s.SelectUser(ctx, user)
		if err != nil {
			return errors.Wrap(err, errors.ErrorNotFound, "user not found")
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM contacts WHERE user_id = ?", id)
		if err != nil {
			return errors.Wrap(err, errors.ErrorUnknown, "")
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM oauth_factors WHERE user_id = ?", id)
		if err != nil {
			return errors.Wrap(err, errors.ErrorUnknown, "")
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM roles_users WHERE user_id = ?", id)
		if err != nil {
			return errors.Wrap(err, errors.ErrorUnknown, "")
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = ?", id)
		if err != nil {
			return errors.Wrap(err, errors.ErrorUnknown, "")
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM second_factors WHERE user_id = ?", id)
		if err != nil {
			return errors.Wrap(err, errors.ErrorUnknown, "")
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM secrets WHERE user_id = ?", id)
		if err != nil {
			return errors.Wrap(err, errors.ErrorUnknown, "")
		}

		db := userStruct.DeleteFrom("users")
		db.Where(db.E("id", id))
		dq, da := db.Build()
		result, err := tx.ExecContext(ctx, dq, da...)
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return errors.Wrap(err, errors.ErrorUnknown, "")
		}
		if rowsAffected == 0 {
			return errors.New(errors.ErrorNotFound, "no record found")
		}
		return nil
	})
}

// AllUsersWithPageOptions returns all users with pagination.
func (s *Store) AllUsersWithPageOptions(ctx context.Context, pageOptions paging.PageOptions) (*[]User, *paging.Page, error) {
	sb := userStruct.SelectFrom("users")
	sq, sa := sb.Build()
	users := []User{}
	page, err := paging.SelectContext(ctx, s.db, pageOptions, &users, sq, sa...)
	if err != nil {
		return nil, nil, err
	}

	return &users, page, nil
}

// AllUsersWithQuery find users that match the given query.
func (s *Store) AllUsersWithQuery(ctx context.Context, query UsersQuery) (*[]User, *paging.Page, error) {
	sb := userStruct.SelectFrom("users")

	if query.Search != "" {
		value := fmt.Sprintf("%%%s%%", query.Search)
		sb.Where(sb.Or(
			sb.Like("email", value),
			sb.Like("phone", value),
			sb.Like("name", value),
			sb.Like("display_name", value),
			sb.Like("username", value),
		))
	}
	if query.Email != "" {
		sb.Where(sb.Like("email", fmt.Sprintf("%%%s%%", query.Email)))
	}
	if query.Phone != "" {
		sb.Where(sb.Like("phone", fmt.Sprintf("%%%s%%", query.Phone)))
	}
	if query.Name != "" {
		name := fmt.Sprintf("%%%s%%", query.Name)
		sb.Where(sb.Or(sb.Like("name", name), sb.Like("display_name", name)))
	}
	if query.Username != "" {
		sb.Where(sb.Like("username", fmt.Sprintf("%%%s%%", query.Username)))
	}

	sq, sa := sb.Build()
	pageOptions := query.PageOptions()
	users := []User{}
	page, err := paging.SelectContext(ctx, s.db, pageOptions, &users, sq, sa...)

	if err != nil {
		return nil, nil, err
	}

	return &users, page, nil
}

// IncreaseResetPasswordCount increases the reset password count by 1, or throws an error if limit is reached.
func (s *Store) IncreaseResetPasswordCount(ctx context.Context, user *User) (*User, error) {
	ResetPasswordCountLimit := viper.GetInt64("reset_password_count_limit")
	if user.ResetPasswordCount == ResetPasswordCountLimit {
		return nil, errors.New(errors.ErrorResourceExhausted, "")
	}
	user.ResetPasswordCount++

	result, err := sqlx.NamedExecContext(ctx, s.db, "UPDATE users SET reset_password_count = :reset_password_count WHERE id = :id", &user)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	if rowsAffected != 1 {
		return nil, errors.New(errors.ErrorUnknown, "no rows changed")
	}

	return user, nil
}

// ClearResetPasswordCount clears the reset password count to 0.
func (s *Store) ClearResetPasswordCount(ctx context.Context, user *User) (*User, error) {
	user.ResetPasswordCount = 0

	_, err := sqlx.NamedExecContext(ctx, s.db, "UPDATE users SET reset_password_count = :reset_password_count WHERE id = :id", &user)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return user, nil
}

// CreateSecondFactor creates a second factor.
func (s *Store) CreateSecondFactor(ctx context.Context, secondFactor *SecondFactor) (*SecondFactor, error) {
	err := s.encryptor.EncryptStruct(secondFactor)
	if err != nil {
		return nil, err
	}

	content, err := secondFactor.Content.Value()
	if err != nil {
		return nil, err
	}

	result, err := s.db.ExecContext(
		ctx,
		"INSERT INTO second_factors (user_id, type, content) VALUES (?, ?, ?)",
		secondFactor.UserID,
		secondFactor.Type,
		content,
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return s.FindSecondFactorByID(ctx, id)
}

// FindSecondFactorByID finds a second factor by id.
func (s *Store) FindSecondFactorByID(ctx context.Context, ID int64) (*SecondFactor, error) {
	secondFactor := &SecondFactor{}
	err := s.db.QueryRowxContext(ctx, "SELECT * FROM second_factors WHERE id = ?", ID).StructScan(secondFactor)
	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	err = s.encryptor.DecryptStruct(secondFactor)
	if err != nil {
		return nil, err
	}

	return secondFactor, nil
}

// LockSecondFactor acquires an exclusive lock to the second factor for update.
func (s *Store) LockSecondFactor(ctx context.Context, factor *SecondFactor) error {
	var id int64
	err := s.db.QueryRowxContext(ctx, "SELECT id FROM second_factors WHERE id = ? FOR UPDATE", factor.ID).Scan(&id)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// FindAllSecondFactorsByUserID finds the list of second factors by a user id.
func (s *Store) FindAllSecondFactorsByUserID(ctx context.Context, userID int64) (*[]SecondFactor, error) {
	secondFactors := &[]SecondFactor{}
	err := sqlx.SelectContext(ctx, s.db, secondFactors, "SELECT * FROM second_factors WHERE user_id = ?", userID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	for ind := range *secondFactors {
		err = s.encryptor.DecryptStruct(&(*secondFactors)[ind])
		if err != nil {
			return nil, err
		}
	}
	return secondFactors, nil
}

// FindAllSecondFactorsByUserIDAndType finds the list of second factors by a user id and a second factor type.
func (s *Store) FindAllSecondFactorsByUserIDAndType(ctx context.Context, userID int64, secondFactorType SecondFactorType) (*[]SecondFactor, error) {
	secondFactors := &[]SecondFactor{}
	err := sqlx.SelectContext(ctx, s.db, secondFactors, "SELECT * FROM second_factors WHERE user_id = ? and type = ?", userID, secondFactorType)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	for ind := range *secondFactors {
		err = s.encryptor.DecryptStruct(&(*secondFactors)[ind])
		if err != nil {
			return nil, err
		}
	}
	return secondFactors, nil
}

// UpdateSecondFactorLastUsedAtByID updates the last used time for second factor by id.
func (s *Store) UpdateSecondFactorLastUsedAtByID(ctx context.Context, id int64) (*SecondFactor, error) {
	_, err := s.db.ExecContext(ctx, "UPDATE second_factors SET last_used_at = NOW() WHERE id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return s.FindSecondFactorByID(ctx, id)
}

// UpdateSecondFactorUsedCodeMaskByID updates the used code mask (for backup code) by id.
func (s *Store) UpdateSecondFactorUsedCodeMaskByID(ctx context.Context, id int64, usedCodeMask int64) (*SecondFactor, error) {
	secondFactor, err := s.FindSecondFactorByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !secondFactor.Content.UsedCodeMask.Valid {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	secondFactor.Content.UsedCodeMask.Int64 = usedCodeMask
	return s.UpdateSecondFactorContent(ctx, secondFactor)
}

// UpdateSecondFactorContent updates the second factor content.
func (s *Store) UpdateSecondFactorContent(ctx context.Context, secondFactor *SecondFactor) (*SecondFactor, error) {
	err := s.encryptor.EncryptStruct(secondFactor)
	if err != nil {
		return nil, err
	}
	secondFactorContent, err := secondFactor.Content.Value()
	if err != nil {
		return nil, err
	}
	_, err = s.db.ExecContext(ctx, "UPDATE second_factors SET content = ? WHERE id = ?", secondFactorContent, secondFactor.ID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return s.FindSecondFactorByID(ctx, secondFactor.ID)
}

// DeleteSecondFactorByID deletes a second factor by id.
func (s *Store) DeleteSecondFactorByID(ctx context.Context, id int64) error {
	_, err := s.db.QueryxContext(ctx, "DELETE FROM second_factors WHERE id = ?", id)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// FindAllOAuthFactorsByUserID finds the list of OAuth factors by a user id.
func (s *Store) FindAllOAuthFactorsByUserID(ctx context.Context, userID int64) (*[]OAuthFactor, error) {
	oauthFactors := &[]OAuthFactor{}
	err := sqlx.SelectContext(ctx, s.db, oauthFactors, "SELECT * FROM oauth_factors WHERE user_id = ?", userID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return oauthFactors, nil
}

// FindAllOAuthFactorsByUserIDAndService finds the list of OAuth factors by a user id and service.
func (s *Store) FindAllOAuthFactorsByUserIDAndService(ctx context.Context, userID int64, service OAuthService) (*[]OAuthFactor, error) {
	oauthFactors := &[]OAuthFactor{}
	err := sqlx.SelectContext(ctx, s.db, oauthFactors, "SELECT * FROM oauth_factors WHERE user_id = ? AND service = ?", userID, service)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return oauthFactors, nil
}

// FindOAuthFactorByID finds an OAuth factor by id.
func (s *Store) FindOAuthFactorByID(ctx context.Context, ID int64) (*OAuthFactor, error) {
	oauthFactor := &OAuthFactor{}
	err := s.db.QueryRowxContext(ctx, "SELECT * FROM oauth_factors WHERE id = ?", ID).StructScan(oauthFactor)
	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return oauthFactor, nil
}

// DeleteOAuthFactorByID deletes an OAuth factor by id.
func (s *Store) DeleteOAuthFactorByID(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM oauth_factors WHERE id = ?", id)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// DeleteOAuthFactorByUserIDAndService deletes an OAuth factor by user id and service.
func (s *Store) DeleteOAuthFactorByUserIDAndService(ctx context.Context, userID int64, service OAuthService) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM oauth_factors WHERE user_id = ? AND service = ?", userID, service)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	if rowsAffected == 0 {
		return errors.New(errors.ErrorNotFound, "no record found")
	}
	return nil
}

// FindOAuthFactorByOAuthIdentity finds a user id by OAuth user.
func (s *Store) FindOAuthFactorByOAuthIdentity(ctx context.Context, service OAuthService, oauthUserID string) (*OAuthFactor, error) {
	oauthFactor := &OAuthFactor{}
	err := s.db.QueryRowxContext(
		ctx,
		"SELECT * FROM oauth_factors WHERE service = ? and oauth_user_id = ?",
		service,
		oauthUserID,
	).StructScan(oauthFactor)
	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return oauthFactor, nil
}

// CreateOAuthFactor creates an oauth factor.
func (s *Store) CreateOAuthFactor(ctx context.Context, userID int64, service OAuthService, oauthUserID string, metadata nulls.JSON) (*OAuthFactor, error) {
	result, err := s.db.ExecContext(
		ctx,
		"INSERT INTO oauth_factors (user_id, service, oauth_user_id, metadata) VALUES (?, ?, ?, ?)",
		userID,
		service,
		oauthUserID,
		metadata,
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return s.FindOAuthFactorByID(ctx, id)
}

// UpdateOAuthFactorMetadata update an oauth factor metadata.
func (s *Store) UpdateOAuthFactorMetadata(ctx context.Context, id int64, metadata nulls.JSON) (*OAuthFactor, error) {
	_, err := s.db.ExecContext(
		ctx,
		"UPDATE oauth_factors SET metadata = ? WHERE id = ?",
		metadata,
		id,
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return s.FindOAuthFactorByID(ctx, id)
}

// UpdateOAuthFactorLastUsedAt update an oauth factor last used time.
func (s *Store) UpdateOAuthFactorLastUsedAt(ctx context.Context, id int64, lastUsedAt time.Time) (*OAuthFactor, error) {
	_, err := s.db.ExecContext(ctx, "UPDATE oauth_factors SET last_used_at = ? WHERE id = ?", lastUsedAt, id)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return s.FindOAuthFactorByID(ctx, id)
}

// CreateRole creates a new non-system role
func (s *Store) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	err := role.Validate()
	if err != nil {
		return nil, err
	}

	result, err := sqlx.NamedExecContext(
		ctx,
		s.db,
		"INSERT INTO roles (name) VALUES (:name)",
		&role)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return s.FindRoleByID(ctx, id)
}

// FindRoleByID finds the role with the given ID.
func (s *Store) FindRoleByID(ctx context.Context, id int64) (*Role, error) {
	role := &Role{}
	err := s.db.QueryRowxContext(ctx, "SELECT * FROM roles WHERE id = ?", id).StructScan(role)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return role, nil
}

// FindRoleByName finds the role with the given name.
func (s *Store) FindRoleByName(ctx context.Context, name string) (*Role, error) {
	role := &Role{}
	err := s.db.QueryRowxContext(ctx, "SELECT * FROM roles WHERE name = ?", name).StructScan(role)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return role, nil
}

// FindAllRoles lists the roles
func (s *Store) FindAllRoles(ctx context.Context) (*[]Role, error) {
	roles := &[]Role{}
	err := sqlx.SelectContext(ctx, s.db, roles, "SELECT * FROM roles")
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return roles, nil
}

// DeleteRoleByID deletes a role by id
func (s *Store) DeleteRoleByID(ctx context.Context, id int64) error {
	role, err := s.FindRoleByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	if role.IsSystemRole {
		return errors.New(errors.ErrorInvalidArgument, "")
	}

	_, err = s.db.QueryxContext(ctx, "DELETE FROM roles WHERE id = ?", id)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// AssignRole creates a subject assignment (role-user relation)
func (s *Store) AssignRole(ctx context.Context, roleUser *RoleUser) error {
	err := roleUser.Validate()
	if err != nil {
		return err
	}

	_, err = s.FindRoleByID(ctx, roleUser.RoleID)
	if err != nil {
		return errors.Wrap(err, errors.ErrorNotFound, "")
	}

	selectedRoleUser := &RoleUser{}
	err = s.db.QueryRowxContext(
		ctx,
		"SELECT * FROM roles_users WHERE role_id = ? AND user_id = ? FOR UPDATE",
		roleUser.RoleID, roleUser.UserID,
	).StructScan(selectedRoleUser)
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	if err == nil {
		return errors.New(errors.ErrorAlreadyExists, "")
	}

	_, err = sqlx.NamedExecContext(
		ctx,
		s.db,
		"INSERT INTO roles_users (role_id, user_id) VALUES (:role_id, :user_id)",
		&roleUser)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// UnassignByRoleIDAndUserID removes a subject assignment (role-user relation)
func (s *Store) UnassignByRoleIDAndUserID(ctx context.Context, roleID, userID int64) error {
	_, err := s.db.QueryxContext(ctx, "DELETE FROM roles_users WHERE role_id = ? AND user_id = ?", roleID, userID)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// FindAllRolesByUserID finds all assigned roles for a given user id
func (s *Store) FindAllRolesByUserID(ctx context.Context, userID int64) (*[]Role, error) {
	roles := &[]Role{}
	err := sqlx.SelectContext(
		ctx, s.db, roles,
		"SELECT roles.* FROM roles JOIN roles_users ON roles_users.role_id = roles.id WHERE roles_users.user_id = ?",
		userID,
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return roles, nil
}

// AllUsersByRoleID finds all users with a given assigned role id
func (s *Store) AllUsersByRoleID(ctx context.Context, roleID int64) (*[]User, error) {
	users := &[]User{}
	err := sqlx.SelectContext(
		ctx, s.db, users,
		"SELECT users.* FROM users JOIN roles_users ON roles_users.user_id = users.id WHERE roles_users.role_id = ?",
		roleID,
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return users, nil
}

// FindSubjectAssignmentByRoleIDAndUserID finds the subject assignment by role ID and user ID
func (s *Store) FindSubjectAssignmentByRoleIDAndUserID(ctx context.Context, roleID, userID int64) (*RoleUser, error) {
	roleUser := &RoleUser{}
	err := s.db.QueryRowxContext(ctx, "SELECT * FROM roles_users WHERE role_id = ? AND user_id = ?", roleID, userID).StructScan(roleUser)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return roleUser, nil
}

// BeforeInsert is called before the model is inserted.
func (s *Store) BeforeInsert(i interface{}) error {
	return s.BeforeUpdate(i)
}

// BeforeUpdate is called before the model is updated.
func (s *Store) BeforeUpdate(i interface{}) error {
	if err := validator.Validate.Struct(i); err != nil {
		return errors.WithValidateError(err)
	}

	if err := s.encryptor.EncryptStruct(i); err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// AfterSelect is called after the model is selected.
func (s *Store) AfterSelect(i interface{}) error {
	return s.encryptor.DecryptStruct(i)
}

type hooks interface {
	BeforeInsert(i interface{}) error
	BeforeUpdate(i interface{}) error
	AfterSelect(i interface{}) error
}

func selectUser(ctx context.Context, q sqlx.QueryerContext, h hooks, user *User, query string, args ...interface{}) error {
	if err := sqlx.GetContext(ctx, q, user, query, args...); err != nil {
		return errors.WithSQLError(err)
	}
	return h.AfterSelect(user)
}
