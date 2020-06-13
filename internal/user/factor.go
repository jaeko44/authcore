package user

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
	"time"

	"authcore.io/authcore/internal/authn/verifier"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/validator"
	"authcore.io/authcore/pkg/nulls"
)

// SecondFactor represents a second factor record in database.
type SecondFactor struct {
	ID         int64
	UserID     int64               `db:"user_id"`
	Type       SecondFactorType    `db:"type"`
	Content    SecondFactorContent `encrypt:"-" db:"content"`
	UpdatedAt  time.Time           `db:"updated_at"`
	CreatedAt  time.Time           `db:"created_at"`
	LastUsedAt time.Time           `db:"last_used_at"`
}

// Validate validates a SecondFactor.
func (secondFactor *SecondFactor) Validate() error {
	return validator.Validate.Struct(secondFactor)
}

// UpdateWithVerifier updates a SecondFactor content with verifier.
func (secondFactor *SecondFactor) UpdateWithVerifier(v verifier.Verifier) (err error) {
	if secondFactor.Type.String() != v.Method() {
		err = errors.New(errors.ErrorInvalidArgument, "mismatch factor type")
		return
	}

	switch vt := v.(type) {
	case verifier.TOTPVerifier:
		break
	case verifier.BackupCodeVerifier:
		secondFactor.Content.UsedCodeMask.Valid = true
		secondFactor.Content.UsedCodeMask.Int64 = vt.UsedCodeMask
	case verifier.SMSOTPVerifier:
		break
	default:
		err = errors.New(errors.ErrorInvalidArgument, "unknown factor type")
	}
	return
}

// ToVerifier converts a SecondFactor model into a verifier.Verifier instance. This method must
// handle legacy data.
func (secondFactor *SecondFactor) ToVerifier(factory *verifier.Factory) (v verifier.Verifier, err error) {
	m := make(map[string]interface{})
	switch secondFactor.Type {
	case SecondFactorTOTP:
		m["method"] = verifier.TOTP
		m["secret"] = secondFactor.Content.Secret
		m["last_used"] = strconv.FormatInt(secondFactor.LastUsedAt.Unix(), 10)
	case SecondFactorSMS:
		m["method"] = verifier.SMSOTP
		m["phone_number"] = secondFactor.Content.PhoneNumber
	case SecondFactorBackupCode:
		m["method"] = verifier.BackupCode
		m["secret"] = secondFactor.Content.Secret
		m["used_code_mask"] = strconv.FormatInt(secondFactor.Content.UsedCodeMask.Int64, 10)
	default:
		err = errors.New(errors.ErrorInvalidArgument, "unknown factor type")
		return
	}

	data, err := json.Marshal(m)
	if err != nil {
		return
	}

	return factory.Unmarshal(data)
}

// SecondFactorType is a type enumerating the types for second factor authentications
type SecondFactorType int32

// Enumerates the SecondFactorType
const (
	SecondFactorSMS        SecondFactorType = 0
	SecondFactorTOTP       SecondFactorType = 1
	SecondFactorBackupCode SecondFactorType = 2
)

// SecondFactorTypeFromString returns a SecondFactorType from the given string.
func SecondFactorTypeFromString(s string) (SecondFactorType, error) {
	switch s {
	case "sms_otp":
		return SecondFactorSMS, nil
	case "totp":
		return SecondFactorTOTP, nil
	case "backup_code":
		return SecondFactorBackupCode, nil
	}
	return SecondFactorSMS, errors.New(errors.ErrorInvalidArgument, "invalid factor type")
}

// StringV1 stringifies the type of the second factor for V1 API.
func (t SecondFactorType) StringV1() (string, error) {
	switch t {
	case SecondFactorSMS:
		return "SMS_CODE", nil
	case SecondFactorTOTP:
		return "TIME_BASED_ONE_TIME_PASSWORD", nil
	case SecondFactorBackupCode:
		return "BACKUP_CODE", nil
	default:
		return "", errors.New(errors.ErrorUnknown, "unknown second factor type")
	}
}

// String returns a string identifier for the second factor type.
func (t SecondFactorType) String() string {
	switch t {
	case SecondFactorSMS:
		return verifier.SMSOTP
	case SecondFactorTOTP:
		return verifier.TOTP
	case SecondFactorBackupCode:
		return verifier.BackupCode
	}
	return ""
}

// SecondFactorContent represents a second factor content.
type SecondFactorContent struct {
	PhoneNumber     nulls.String `json:"phone_number"`                                                // For SMS
	Identifier      nulls.String `json:"identifier"`                                                  // For TOTP
	Secret          nulls.String `json:"-" encrypt:"" encryptPurpose:"second_factors.content.secret"` // For TOTP
	EncryptedSecret nulls.String `json:"encrypted_secret"`                                            // For TOTP & backup code
	UsedCodeMask    nulls.Int64  `json:"used_code_mask"`                                              // For backup code
}

// Scan scans the byte array / string into a SecondFactorContent object.
// source: https://gist.github.com/rrafal/09534862e05cd98e4eb9b17dd5fcc1fc
func (secondFactorContent *SecondFactorContent) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &secondFactorContent)
		return nil
	case string:
		json.Unmarshal([]byte(v), &secondFactorContent)
		return nil
	default:
		return errors.Errorf(errors.ErrorUnknown, "unsupported type %T", v)
	}
}

// Value marshals the SecondFactorContent object into the driver.Value object.
func (secondFactorContent *SecondFactorContent) Value() (driver.Value, error) {
	val, err := json.Marshal(&secondFactorContent)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return val, nil
}

// AuthorizationToken represents the authorization token that is used as one-time refresh token generator.
type AuthorizationToken struct {
	AuthorizationToken string `validate:"byte=32"`
	CodeChallenge      string
	UserID             int64 `validate:"min=1"`
	DeviceID           int64 `validate:"min=0"`
}

// Validate validates an AuthorizationToken.
func (as *AuthorizationToken) Validate() error {
	return validator.Validate.Struct(as)
}

// ResetPasswordToken represents the reset password token that is used as one-time authorization token for reset password.
type ResetPasswordToken struct {
	ResetPasswordToken string `validate:"byte=32"`
	UserID             int64  `validate:"min=1"`
}

// Validate validates an ResetPasswordToken.
func (as *ResetPasswordToken) Validate() error {
	return validator.Validate.Struct(as)
}

// OAuthService is a type enumerating the oauth services
type OAuthService int32

// MarshalJSON marshal OAuthService into service string.
func (o OAuthService) MarshalJSON() ([]byte, error) {
	return json.Marshal(OAuthServiceToName(o))
}

// Enumerates the OAuthService
const (
	OAuthFacebook OAuthService = 0
	OAuthGoogle   OAuthService = 1
	OAuthApple    OAuthService = 2
	OAuthMatters  OAuthService = 3
	OAuthTwitter  OAuthService = 4
)

// ToOAuthService converts string to OAuthService.
func ToOAuthService(s string) (o OAuthService, err error) {
	switch s {
	case "facebook":
		o = OAuthFacebook
	case "google":
		o = OAuthGoogle
	case "apple":
		o = OAuthApple
	case "matters":
		o = OAuthMatters
	case "twitter":
		o = OAuthTwitter
	default:
		return -1, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return
}

// OAuthServiceToName converts OAuthService to name.
func OAuthServiceToName(o OAuthService) string {
	switch o {
	case OAuthFacebook:
		return "facebook"
	case OAuthGoogle:
		return "google"
	case OAuthApple:
		return "apple"
	case OAuthMatters:
		return "matters"
	case OAuthTwitter:
		return "twitter"
	default:
		return strconv.FormatInt(int64(o), 10)
	}
}

// OAuthFactor is a struct describing the oauth factor
type OAuthFactor struct {
	ID          int64        `json:"id"`
	UserID      int64        `db:"user_id" json:"user_id"`
	Service     OAuthService `db:"service" json:"service"`
	OAuthUserID string       `db:"oauth_user_id" json:"oauth_user_id"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	LastUsedAt  time.Time    `db:"last_used_at" json:"last_used_at"`
	Metadata    nulls.JSON   `db:"metadata" json:"metadata"`
}

// ServiceName returns the service name.
func (f OAuthFactor) ServiceName() string {
	return OAuthServiceToName(f.Service)
}
