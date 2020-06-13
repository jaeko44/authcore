package authapi

import (
	"authcore.io/authcore/internal/audit"
	"authcore.io/authcore/internal/v1/authentication"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/email"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/sms"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/ratelimiter"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

// Service provides GRPC Auth API implementations.
type Service struct {
	DB                    *db.DB
	Redis                 *redis.Client
	UserStore             *user.Store
	SessionStore          *session.Store
	AuthenticationService *authentication.Service
	AuditStore            *audit.Store
	RateLimiters          *RateLimiters
	EmailService          *email.Service
	SMSService            *sms.Service
}

// RateLimiters is the set of rate limiters used.
type RateLimiters struct {
	SecondFactorRateLimiter   *ratelimiter.RateLimiter
	ContactRateLimiter        *ratelimiter.RateLimiter
	AuthenticationRateLimiter *ratelimiter.RateLimiter
}

// NewService initialize a new Service.
func NewService(db *db.DB,
	redis *redis.Client,
	userStore *user.Store,
	sessionStore *session.Store,
	authenticationService *authentication.Service,
	auditStore *audit.Store,
	emailService *email.Service,
	smsService *sms.Service) *Service {
	secondFactorRateLimitInterval := viper.GetDuration("second_factor_rate_limit_interval")
	secondFactorRateLimtCount := viper.GetInt64("second_factor_rate_limit_count")
	contactRateLimitInterval := viper.GetDuration("contact_rate_limit_interval")
	contactRateLimitCount := viper.GetInt64("contact_rate_limit_count")
	authenticationLimitInterval := viper.GetDuration("authentication_rate_limit_interval")
	authenticationLimitCount := viper.GetInt64("authentication_rate_limit_count")

	secondFactorRateLimiter := ratelimiter.NewRateLimiter(redis, "rate_limiter/", secondFactorRateLimtCount, secondFactorRateLimitInterval)
	contactRateLimiter := ratelimiter.NewRateLimiter(redis, "rate_limiter/", contactRateLimitCount, contactRateLimitInterval)
	authenticationRateLimiter := ratelimiter.NewRateLimiter(redis, "rate_limiter/", authenticationLimitCount, authenticationLimitInterval)
	srv := &Service{
		DB:                    db,
		Redis:                 redis,
		UserStore:             userStore,
		SessionStore:          sessionStore,
		AuthenticationService: authenticationService,
		AuditStore:            auditStore,
		RateLimiters: &RateLimiters{
			SecondFactorRateLimiter:   secondFactorRateLimiter,
			ContactRateLimiter:        contactRateLimiter,
			AuthenticationRateLimiter: authenticationRateLimiter,
		},
		EmailService: emailService,
		SMSService:   smsService,
	}
	parsePrivateKeys()
	return srv
}
