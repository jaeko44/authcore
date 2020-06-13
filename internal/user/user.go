package user

import (
	"encoding/base64"
	"strconv"
	"time"

	"authcore.io/authcore/internal/authn/verifier"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/languages"
	"authcore.io/authcore/pkg/nulls"
	"authcore.io/authcore/pkg/paging"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
)

// User represents a user record in database. For Email and Phone fields validation, check
// UserStructLevelValidation as it requires customized logic for contact requirement.
type User struct {
	ID                          int64           `db:"id"`
	Name                        nulls.String    `db:"name" fieldtag:"insert,update"`
	Username                    nulls.String    `db:"username" fieldtag:"insert,update"`
	Email                       nulls.String    `db:"email" fieldtag:"insert,update" validate:"omitempty,email"`
	EmailVerifiedAt             nulls.Time      `db:"email_verified_at" fieldtag:"insert,update"`
	Phone                       nulls.String    `db:"phone" fieldtag:"insert,update" validate:"omitempty,phone"`
	PhoneVerifiedAt             nulls.Time      `db:"phone_verified_at" fieldtag:"insert,update"`
	RecoveryEmail               nulls.String    `db:"recovery_email"`                        // Unused
	RecoveryEmailVerifiedAt     nulls.Time      `db:"recovery_email_verified_at"`            // Unused
	DisplayNameOld              string          `db:"display_name" fieldtag:"insert,update"` // Depreciated
	PasswordSaltBase64          nulls.String    `db:"password_salt" fieldtag:"insert,update"`
	PasswordVerifierW0          nulls.ByteSlice `db:"-" encrypt:"" encryptPurpose:"users.password_verifier_w0"`
	PasswordVerifierL           nulls.ByteSlice `db:"-" encrypt:"" encryptPurpose:"users.password_verifier_l"`
	EncryptedPasswordVerifierW0 nulls.String    `db:"encrypted_password_verifier_w0" fieldtag:"insert,update"`
	EncryptedPasswordVerifierL  nulls.String    `db:"encrypted_password_verifier_l" fieldtag:"insert,update"`
	IsLocked                    bool            `db:"is_locked" fieldtag:"update"`
	LockExpiredAt               nulls.Time      `db:"lock_expired_at" fieldtag:"update"`
	LockDescription             nulls.String    `db:"lock_description" fieldtag:"update"`
	UserMetadata                nulls.JSON      `db:"user_metadata" fieldtag:"insert,update"`
	AppMetadata                 nulls.JSON      `db:"app_metadata" fieldtag:"insert,update"`
	UpdatedAt                   time.Time       `db:"updated_at"`
	CreatedAt                   time.Time       `db:"created_at"`
	ResetPasswordCount          int64           `db:"reset_password_count" fieldtag:"update"`
	Language                    nulls.String    `db:"language" fieldtag:"insert,update" validate:"omitempty,language"`
	LastSeenAt                  time.Time       `db:"last_seen_at" fieldtag:"update"`
}

// EmailVerified returns whether the email is verified.
func (user *User) EmailVerified() bool {
	return user.EmailVerifiedAt.Valid
}

// PhoneVerified returns whether the phone number is verified.
func (user *User) PhoneVerified() bool {
	return user.PhoneVerifiedAt.Valid
}

// SetPasswordVerifier updates the password verifier fields in a User model.
func (user *User) SetPasswordVerifier(salt, verifierW0, verifierL []byte) error {
	if len(salt) != 32 {
		return errors.New(errors.ErrorInvalidArgument, "salt must be 32 bytes long")
	}
	if len(verifierW0) == 0 || len(verifierL) == 0 {
		return errors.New(errors.ErrorInvalidArgument, "verifier must not be empty")
	}

	user.PasswordSaltBase64 = nulls.NewString(base64.RawURLEncoding.EncodeToString(salt))
	user.PasswordVerifierW0 = nulls.NewByteSlice(verifierW0)
	user.PasswordVerifierL = nulls.NewByteSlice(verifierL)
	return nil
}

// PublicID returns a ID string that is suitable to be used by clients.
func (user *User) PublicID() string {
	return strconv.FormatInt(user.ID, 10)
}

// ActorID returns a ID for audit logs.
func (user *User) ActorID() int64 {
	return user.ID
}

// DisplayName returns a string that is suitable for representing the user.
func (user *User) DisplayName() string {
	if user.Name.Valid {
		return user.Name.String
	}
	if user.Email.Valid {
		return user.Email.String
	}
	if user.Phone.Valid {
		return user.Phone.String
	}
	if user.DisplayNameOld != "" {
		return user.DisplayNameOld
	}
	return user.PublicID()
}

// PasswordSalt returns the decoded password salt of the User.
func (user *User) PasswordSalt() []byte {
	if user.PasswordSaltBase64.Valid {
		value, err := base64.RawURLEncoding.DecodeString(user.PasswordSaltBase64.String)
		if err != nil {
			log.Printf("unable to decode base64 salt: %v", err)
			return nil
		}
		return value
	}
	return nil
}

// IsCurrentlyLocked checks if a contact is currently locked
func (user *User) IsCurrentlyLocked() bool {
	return user.IsLocked && user.LockExpiredAt.Time.After(time.Now())
}

// IsPasswordAuthenticationEnabled checks if an user can be authenticated with password
func (user *User) IsPasswordAuthenticationEnabled() bool {
	return len(user.PasswordSalt()) > 0 && user.PasswordVerifierW0.Valid && user.PasswordVerifierL.Valid
}

// RealLanguage checks the language in user model is available in system and return it or the fallback value
func (user *User) RealLanguage() string {
	if languages.CheckAvailableLanguages(user.Language.String) {
		return user.Language.String
	}
	return viper.GetString("default_language")
}

// PasswordVerifier returns the password-based authentication method of a user.
func (user *User) PasswordVerifier() (v verifier.Verifier, err error) {
	salt := user.PasswordSalt()
	w0 := user.PasswordVerifierW0.ByteSlice
	l := user.PasswordVerifierL.ByteSlice

	if len(salt) == 0 || len(w0) == 0 || len(l) == 0 {
		err = errors.New(errors.ErrorInvalidArgument, "password was not set for user")
		return
	}

	v = verifier.NewSPAKE2PlusVerifier(salt, w0, l)

	if !v.IsPrimary() {
		v = nil
		err = errors.New(errors.ErrorPermissionDenied, "not a primary factor")
	}
	return
}

// UpdatePasswordWithVerifier updates a user password authentication with the verifier.
func (user *User) UpdatePasswordWithVerifier(v verifier.Verifier) (err error) {
	switch vt := v.(type) {
	case verifier.SPAKE2PlusVerifier:
		err = user.SetPasswordVerifier(vt.SaltValue, vt.W0, vt.L)
	default:
		err = errors.New(errors.ErrorInvalidArgument, "not a password-based verifier")
	}
	return
}

var userSortColumns = []string{"created_at", "last_seen_at", "username"}

// UsersQuery is a query for selecting users
type UsersQuery struct {
	PageToken string `query:"page_token"`
	SortBy    string `query:"sort_by"`
	Limit     uint   `query:"limit" validate:"omitempty,gte=0,lte=1000"`
	Search    string `query:"search"`
	Email     string `query:"email"`
	Phone     string `query:"phone_number"`
	Name      string `query:"name"`
	Username  string `query:"preferred_username"`
}

// PageOptions returns a PageOptions for the query.
func (q *UsersQuery) PageOptions() paging.PageOptions {
	sortColumn, sortDirection, err := paging.ParseSortBy(q.SortBy)
	if err != nil || !funk.ContainsString(userSortColumns, sortColumn) {
		sortColumn = userSortColumns[0] // default sort column
	}
	limit := q.Limit
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	return paging.PageOptions{
		SortColumn:     sortColumn,
		UniqueColumn:   "id",
		SortDirection:  sortDirection,
		CountFoundRows: true,
		Limit:          limit,
		PageToken:      paging.PageToken(q.PageToken),
	}
}
