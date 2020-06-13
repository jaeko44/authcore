package verifier

import (
	"context"
	"encoding/json"
	"time"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/sms"
	"authcore.io/authcore/pkg/cryptoutil"
	"authcore.io/authcore/pkg/ratelimiter"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// SMSOTP represents a SMS OTP verifier.
	SMSOTP string = "sms_otp"
)

// SMSOTPVerifier verifers SMS OTP. This verify depends on other services to provide its functions.
type SMSOTPVerifier struct {
	MethodName  string `json:"method"`
	PhoneNumber string `json:"phone_number"`

	smsService  *sms.Service
	rateLimiter *ratelimiter.RateLimiter
}

// NewSMSOTPVerifier returns a new SMSOTPVerifier instance.
func NewSMSOTPVerifier(phoneNumber string, smsService *sms.Service, redisClient *redis.Client) SMSOTPVerifier {
	return SMSOTPVerifier{
		MethodName:  SMSOTP,
		PhoneNumber: phoneNumber,

		smsService:  smsService,
		rateLimiter: smsRateLimiter(redisClient),
	}
}

// Method returns "sms_otp".
func (v SMSOTPVerifier) Method() string {
	return SMSOTP
}

// IsPrimary returns whether this method can be used as the primary authentication.
func (v SMSOTPVerifier) IsPrimary() bool {
	return false
}

// SkipMFA returns whether this method is sufficient for completing the authentication.
func (v SMSOTPVerifier) SkipMFA() bool {
	return false
}

// Salt returns a salt. Or nil if salt is not used by the verifier.
func (v SMSOTPVerifier) Salt() []byte {
	return nil
}

// Request requests a SMS challenge to be sent to the phone number.
func (v SMSOTPVerifier) Request(in []byte) (state State, challenge Challenge, err error) {
	if len(v.PhoneNumber) == 0 {
		err = errors.New(errors.ErrorInvalidArgument, "invalid verifier")
		return
	}
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

	codeLength := int64(viper.GetInt("sms_code_length"))
	code := cryptoutil.RandomCode(codeLength)
	expiry := viper.GetDuration("sms_code_expiry")
	expireAt := time.Now().Add(expiry)

	codeState := codeState{
		Code:     code,
		ExpireAt: expireAt,
	}

	log.WithFields(log.Fields{
		"phone_number": v.PhoneNumber,
	}).Info("send verification sms")

	ctx := context.Background()
	err = v.smsService.SendAuthenticationSMS(
		ctx,
		"",
		v.PhoneNumber,
		code,
	)
	if err != nil {
		return
	}
	state, err = codeState.ToState()
	return
}

// Verify verifies the incoming TOTP code. Returns true and an updated verifier if the code is
// valid. Otherwise, return false and nil.
func (v SMSOTPVerifier) Verify(state State, in []byte) (bool, Verifier) {
	if len(state) == 0 {
		return false, nil
	}
	if len(in) == 0 {
		return false, nil
	}

	cs, err := codeStateFromState(state)
	if err != nil {
		log.Error("invalid sms code state")
		return false, nil
	}

	if cs.Code == "" {
		log.Error("invalid sms code state")
		return false, nil
	}

	code := string(in)
	if cs.Expired() {
		log.WithFields(log.Fields{
			"phone_number": v.PhoneNumber,
		}).Error("SMS OTP expired")
		return false, nil
	}

	if code != cs.Code {
		log.WithFields(log.Fields{
			"phone_number": v.PhoneNumber,
		}).Error("SMS OTP rejected")
		return false, nil
	}

	return true, nil
}

// SMSOTPVerifierFactory returns a function that unmarshalls SMSOTPVerifier from a JSON data.
func SMSOTPVerifierFactory(smsService *sms.Service, redisClient *redis.Client) Unmarshaller {
	rateLimiter := smsRateLimiter(redisClient)
	return func(data []byte) (Verifier, error) {
		return smsotpVerifierFromJSON(data, smsService, rateLimiter)
	}
}

func smsotpVerifierFromJSON(data []byte, smsService *sms.Service, rateLimiter *ratelimiter.RateLimiter) (v Verifier, err error) {
	t := SMSOTPVerifier{}
	if err = json.Unmarshal(data, &t); err != nil {
		return
	}

	t.smsService = smsService
	t.rateLimiter = rateLimiter
	v = t
	return
}

func smsRateLimiter(redisClient *redis.Client) *ratelimiter.RateLimiter {
	contactRateLimitInterval := viper.GetDuration("contact_rate_limit_interval")
	contactRateLimitCount := viper.GetInt64("contact_rate_limit_count")

	return ratelimiter.NewRateLimiter(redisClient, "rate_limiter/sms/", contactRateLimitCount, contactRateLimitInterval)
}

type codeState struct {
	Code     string    `json:"code" validate:"required"`
	ExpireAt time.Time `json:"expire_at" validate:"required"`
}

func codeStateFromState(s State) (cs codeState, err error) {
	err = json.Unmarshal([]byte(s), &cs)
	return
}

func (s codeState) Expired() bool {
	return time.Now().After(s.ExpireAt)
}

func (s codeState) ToState() (State, error) {
	bytes, err := json.Marshal(&s)
	if err != nil {
		return State{}, err
	}
	return State(bytes), nil
}
