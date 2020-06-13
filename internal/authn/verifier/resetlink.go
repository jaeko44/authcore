package verifier

import (
	"context"
	"encoding/json"
	"time"
	"net/url"

	"authcore.io/authcore/internal/email"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/sms"
	"authcore.io/authcore/pkg/cryptoutil"
	"authcore.io/authcore/pkg/ratelimiter"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// ResetLink represents a reset link verifier.
	ResetLink string = "reset_link"
)

// ResetLinkVerifier verifers reset links. This verify depends on other services to provide its functions.
type ResetLinkVerifier struct {
	MethodName  string `json:"method"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`

	// These params are used to build the reset link instead of for verification. They should be
	// refactored out to keep the interface clean.
	ClientID string `json:"client_id"`
	StateToken string `json:"state_token"`
	Lang string `json:"lang"`

	smsService   *sms.Service
	emailService *email.Service
	rateLimiter  *ratelimiter.RateLimiter
}

// NewResetLinkVerifier returns a new ResetLinkVerifier instance.
func NewResetLinkVerifier(phoneNumber string, email string, clientID string, smsService *sms.Service, emailService *email.Service, redisClient *redis.Client) ResetLinkVerifier {
	return ResetLinkVerifier{
		MethodName:  ResetLink,
		PhoneNumber: phoneNumber,
		Email:       email,
		ClientID: clientID,

		smsService:   smsService,
		emailService: emailService,
		rateLimiter:  rateLimiter(redisClient),
	}
}

// Method returns "sms_otp".
func (v ResetLinkVerifier) Method() string {
	return v.MethodName
}

// IsPrimary returns whether this method can be used as the primary authentication.
func (v ResetLinkVerifier) IsPrimary() bool {
	return false
}

// SkipMFA returns whether this method is sufficient for completing the authentication.
func (v ResetLinkVerifier) SkipMFA() bool {
	return false
}

// Salt returns a salt. Or nil if salt is not used by the verifier.
func (v ResetLinkVerifier) Salt() []byte {
	return nil
}

// Request requests a reset link challenge to be sent to the email or phone number.
func (v ResetLinkVerifier) Request(in []byte) (state State, challenge Challenge, err error) {
	if len(v.PhoneNumber) == 0 && len(v.Email) == 0 {
		err = errors.New(errors.ErrorInvalidArgument, "invalid verifier")
		return
	}

	token := cryptoutil.RandomToken()
	expiry := viper.GetDuration("reset_link_expiry")
	expireAt := time.Now().Add(expiry)
	linkState := resetLinkState{
		Token:     token,
		ExpireAt: expireAt,
	}
	baseURL, err := url.Parse(viper.GetString("base_url"))
	resetLinkURL, err := baseURL.Parse("/widgets/reset-password")
	if err != nil {
		log.Fatal(err.Error())
	}
	q := resetLinkURL.Query()
	q.Add("clientId", v.ClientID)
	q.Add("stateToken", v.StateToken)
	q.Add("resetToken", token)
	resetLinkURL.RawQuery = q.Encode()

	ctx := context.Background()
	if len(v.PhoneNumber) > 0 {
		err = v.rateLimiter.Check(v.PhoneNumber)
		if err != nil {
			err = errors.New(errors.ErrorResourceExhausted, "")
			return
		}
		err = v.rateLimiter.Increment(v.PhoneNumber)
		if err != nil {
			err = errors.New(errors.ErrorResourceExhausted, "")
			return
		}
		log.WithFields(log.Fields{
			"phone_number": v.PhoneNumber,
		}).Info("sending reset link")

		err = v.smsService.SendResetLinkV2(
			ctx,
			resetLinkURL.String(),
			v.PhoneNumber,
		)
		if err != nil {
			return
		}
	} else {
		err = v.rateLimiter.Check(v.Email)
		if err != nil {
			err = errors.New(errors.ErrorResourceExhausted, "")
			return
		}
		err = v.rateLimiter.Increment(v.Email)
		if err != nil {
			err = errors.New(errors.ErrorResourceExhausted, "")
			return
		}
		log.WithFields(log.Fields{
			"email": v.Email,
		}).Info("sending reset link")

		err = v.emailService.SendResetLinkV2(
			ctx,
			resetLinkURL.String(),
			v.Email,
			v.Lang,
		)
		if err != nil {
			return
		}
	}

	state, err = linkState.ToState()
	return
}

// Verify verifies the incoming TOTP code. Returns true and an updated verifier if the code is
// valid. Otherwise, return false and nil.
func (v ResetLinkVerifier) Verify(state State, in []byte) (bool, Verifier) {
	if len(state) == 0 {
		return false, nil
	}
	if len(in) == 0 {
		return false, nil
	}

	linkState, err := resetLinkStateFromState(state)
	if err != nil {
		log.Error("invalid reset link state")
		return false, nil
	}

	if linkState.Token == "" {
		log.Error("invalid reset link state")
		return false, nil
	}

	token := string(in)
	if linkState.Expired() {
		log.WithFields(log.Fields{
			"phone_number": v.PhoneNumber,
			"email": v.Email,
		}).Error("reset link expired")
		return false, nil
	}

	if token != linkState.Token {
		log.WithFields(log.Fields{
			"phone_number": v.PhoneNumber,
			"email": v.Email,
		}).Error("reset link is invalid")
		return false, nil
	}

	return true, nil
}

// ResetLinkVerifierFactory returns a function that unmarshalls ResetLinkVerifier from a JSON data.
func ResetLinkVerifierFactory(smsService *sms.Service, emailService *email.Service, redisClient *redis.Client) Unmarshaller {
	rateLimiter := rateLimiter(redisClient)
	return func(data []byte) (Verifier, error) {
		return resetLinkVerifierFromJSON(data, smsService, emailService, rateLimiter)
	}
}

func resetLinkVerifierFromJSON(data []byte, smsService *sms.Service, emailService *email.Service, rateLimiter *ratelimiter.RateLimiter) (v Verifier, err error) {
	t := ResetLinkVerifier{}
	if err = json.Unmarshal(data, &t); err != nil {
		return
	}

	t.smsService = smsService
	t.emailService = emailService
	t.rateLimiter = rateLimiter
	v = t
	return
}

func rateLimiter(redisClient *redis.Client) *ratelimiter.RateLimiter {
	rateLimitInterval := viper.GetDuration("reset_link_rate_limit_interval")
	rateLimitCount := viper.GetInt64("reset_link_rate_limit_count")

	return ratelimiter.NewRateLimiter(redisClient, "rate_limiter/reset_link/", rateLimitCount, rateLimitInterval)
}

type resetLinkState struct {
	Token    string    `json:"token" validate:"required"`
	ExpireAt time.Time `json:"expire_at" validate:"required"`
}

func resetLinkStateFromState(s State) (cs resetLinkState, err error) {
	err = json.Unmarshal([]byte(s), &cs)
	return
}

func (s resetLinkState) Expired() bool {
	return time.Now().After(s.ExpireAt)
}

func (s resetLinkState) ToState() (State, error) {
	bytes, err := json.Marshal(&s)
	if err != nil {
		return State{}, err
	}
	return State(bytes), nil
}
