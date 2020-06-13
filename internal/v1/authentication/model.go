package authentication

import (
	"authcore.io/authcore/internal/validator"

	"github.com/go-redis/redis"
)

// ProofOfWorkChallenge represents a proof-of-work challenge.
type ProofOfWorkChallenge struct {
	Token      string `validate:"uuid"`
	Challenge  []byte `validate:"byte=16"`
	Difficulty int64  `validate:"required"`
}

// ChallengeDB is a database access object for authentication challenges
type ChallengeDB struct {
	redisClient *redis.Client
}

// Validate validates a ProofOfWorkChallenge
func (powChallenge *ProofOfWorkChallenge) Validate() error {
	return validator.Validate.Struct(powChallenge)
}

// State represents the state of an authentication request.
type State struct {
	ClientID           string   `validate:"client_id"`
	TemporaryToken     string   `validate:"byte=32"`
	UserID             int64    `validate:"oauth_user_id"`
	DeviceID           int64    `validate:"min=0"` // TODO: Device ID should have minimum 1
	Challenges         []string `validate:"challenge_set"`
	PKCEChallenge      string   ``
	SuccessRedirectURL string   `validate:"success_redirect_url"`
}

// Validate validates an State.
func (as *State) Validate() error {
	return validator.Validate.Struct(as)
}

// IsOAuth checks if the authentication state is for OAuth.
func (as *State) IsOAuth() bool {
	return len(as.Challenges) == 1 && as.Challenges[0] == "OAUTH"
}

// AuthorizationToken represents the authorization token that is used as one-time refresh token generator.
type AuthorizationToken struct {
	ClientID           string ``
	AuthorizationToken string `validate:"byte=32"`
	CodeChallenge      string ``
	UserID             int64  `validate:"min=1"`
	DeviceID           int64  `validate:"min=0"`
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
