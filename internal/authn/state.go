package authn

import (
	"net/url"
	"strings"

	"authcore.io/authcore/internal/authn/idp"
	"authcore.io/authcore/internal/authn/verifier"
	"authcore.io/authcore/internal/clientapp"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/cryptoutil"
	"authcore.io/authcore/pkg/httputil"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// StatusPrimary represents that the user has requested to authenticate with primary methods.
	StatusPrimary string = "PRIMARY"
	// StatusMFARequired represents that the user must complete a secondary authentication.
	StatusMFARequired string = "MFA_REQUIRED"
	// StatusSuccess represents the transaction completed successfully.
	StatusSuccess string = "SUCCESS"
	// StatusBlocked represents the user account is locked.
	StatusBlocked string = "BLOCKED"
	// StatusIDP represents that the user has requested to authenticate with a third-party identity provider.
	StatusIDP string = "IDP"
	// StatusIDPAlreadyExists represents that the user has authenticated with an IDP but the email
	// has already registered by a user.
	StatusIDPAlreadyExists string = "IDP_ALREADY_EXISTS"
	// StatusIDPBinding represents that the user has requested to link with a third-party identity provider.
	StatusIDPBinding string = "IDP_BINDING"
	// StatusIDPBindingSuccess represents that the IDP binding transaction completed successfuly.
	StatusIDPBindingSuccess string = "IDP_BINDING_SUCCESS"
	// StatusStepUp represents that the user has requested a session step-up authentication.
	StatusStepUp string = "STEP_UP"
	// StatusStepUpSuccess represents that the step-up password verification transaction completed successfuly.
	StatusStepUpSuccess string = "STEP_UP_SUCCESS"
	// StatusPasswordReset represents that a password reset link is requested.
	StatusPasswordReset string = "PASSWORD_RESET"
	// StatusPasswordResetSuccess represents that a password reset is completed successfully.
	StatusPasswordResetSuccess string = "PASSWORD_RESET_SUCCESS"

	// FactorPassword is the password factor
	FactorPassword string = "password"
	// FactorSMS is the SMS factor
	FactorSMS string = "sms"
	// FactorTOTP is the TOTP factor
	FactorTOTP string = "totp"
)

var builtInURLPaths = []string{
	"/widgets/settings",
}

var validate = validator.New()

// State represents the state of an authentication request.
type State struct {
	StateToken            string         `json:"state_token" validate:"required"`
	Status                string         `json:"status" validate:"required"`
	ClientID              string         `json:"client_id" validate:"required"`
	UserID                int64          `json:"user_id,string"`
	SessionID             int64          `json:"session_id,string"`
	PasswordVerifierState verifier.State `json:"password_verifier_state"`
	PasswordVerified      bool           `json:"password_verified"`
	MFAMethod             string         `json:"mfa_method"`
	MFAVerifierState      verifier.State `json:"mfa_verifier_state"`
	ResetLinkState        verifier.State `json:"reset_link_state"`
	IDP                   string         `json:"idp"`
	IDPState              idp.State      `json:"idp_state"`
	RedirectURI           string         `json:"redirect_uri" validate:"omitempty,uri"`
	PKCEChallenge         string         `json:"code_challenge"`
	PKCEChallengeMethod   string         `json:"code_challenge_method" validate:"required_with=PKCEChallenge"`
	AuthorizationCode     string         `json:"authorization_code"`
	ClientState           string         `json:"client_state"`

	Factors             []string `json:"-"`
	PasswordMethod      string   `json:"-"`
	PasswordSalt        []byte   `json:"-"`
	IDPAuthorizationURL string   `json:"-"`
}

// Validate validates an State.
func (s *State) Validate() error {
	return validate.Struct(s)
}

// AppendFactor appends a factor to the factors list.
func (s *State) AppendFactor(factor string) {
	s.Factors = append(s.Factors, factor)
}

// ClearFactors clears the factors list.
func (s *State) ClearFactors() {
	s.Factors = nil
	s.PasswordMethod = ""
	s.PasswordSalt = nil
}

// GenerateAuthorizationCode geneartes a new authorization code and return the instance.
func (s *State) GenerateAuthorizationCode() *AuthorizationCode {
	code := cryptoutil.RandomToken32()
	s.AuthorizationCode = code
	return &AuthorizationCode{
		Code:                code,
		ClientID:            s.ClientID,
		UserID:              s.UserID,
		RedirectURI:         s.RedirectURI,
		PKCEChallengeMethod: s.PKCEChallengeMethod,
		PKCEChallenge:       s.PKCEChallenge,
		PasswordVerified:    s.PasswordVerified,
	}
}

// AuthorizationCode represents an one-time token that can be exchanged for a session. It is issued
// when an authentication transaction completes with the SUCCESS status.
type AuthorizationCode struct {
	Code                string `json:"code" validate:"required"`
	ClientID            string `json:"client_id" validate:"required"`
	UserID              int64  `json:"string" validate:"required"`
	RedirectURI         string `json:"redirect_uri" validate:"uri"`
	PKCEChallenge       string `json:"code_challenge"`
	PKCEChallengeMethod string `json:"code_challenge_method" validate:"required_with=PKCEChallenge"`
	PasswordVerified    bool   `json:"password_verified"`
}

// Validate validates an AuthorizationToken.
func (c *AuthorizationCode) Validate() error {
	return validate.Struct(c)
}

// ValidateRedirectURI validates if the redirect URI allowed by the given client ID.
func ValidateRedirectURI(clientID, redirectURI string) error {
	clientApp, err := clientapp.GetByClientID(clientID)
	if clientApp == nil {
		return errors.New(errors.ErrorInvalidArgument, "invalid client_id")
	}
	acceptURIPrefixes := clientApp.AllowedCallbackURLs

	normalizedURI, err := httputil.NormalizeURI(redirectURI)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "redirect_uri is not a valid uri")
	}
	for _, acceptURIPrefix := range acceptURIPrefixes {
		if strings.HasPrefix(normalizedURI, acceptURIPrefix) {
			return nil
		}
	}
	if isBuiltInURL(normalizedURI) {
		return nil
	}

	return errors.Errorf(errors.ErrorInvalidArgument, "redirect_uri %v is not allowed", redirectURI)
}

func isBuiltInURL(s string) bool {
	baseURL, err := url.Parse(viper.GetString("base_url"))
	if err != nil {
		log.Fatalf("invalid base_url: %v", err)
	}
	for _, path := range builtInURLPaths {
		t, err := baseURL.Parse(path)
		if err != nil {
			log.Fatalf("error building OAuth callback URL: %v", err)
		}
		if strings.HasPrefix(s, t.String()) {
			return true
		}
	}
	return false
}
