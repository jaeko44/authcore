package authn

import (
	"context"
	"encoding/json"
	"fmt"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/messageencryptor"
	"authcore.io/authcore/pkg/ratelimiter"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

const (
	authnStateKeyPrefix        = "authn_state/"
	authorizationCodeKeyPrefix = "authorization_code/"
)

// Store manages the State model
type Store struct {
	redis       *redis.Client
	encryptor   *messageencryptor.MessageEncryptor
	rateLimiter *ratelimiter.RateLimiter
}

// NewStore initialize a new Store.
func NewStore(redis *redis.Client, encryptor *messageencryptor.MessageEncryptor) *Store {
	rateLimitInterval := viper.GetDuration("authentication_rate_limit_interval")
	rateLimitCount := viper.GetInt64("authentication_rate_limit_count")
	rateLimiter := ratelimiter.NewRateLimiter(redis, "rate_limiter/authn/", rateLimitCount, rateLimitInterval)

	return &Store{
		redis:       redis,
		encryptor:   encryptor,
		rateLimiter: rateLimiter,
	}
}

// PutState save a state to the store.
func (s *Store) PutState(ctx context.Context, state *State) error {
	err := state.Validate()
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	key := authnStateKeyPrefix + state.StateToken
	data, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	encryptedData, err := s.encryptor.Encrypt(data, []byte(key))
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}

	expiry := viper.GetDuration("authentication_time_limit")
	err = s.redis.Set(key, encryptedData, expiry).Err()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// GetState retrieves a state from the store.
func (s *Store) GetState(ctx context.Context, stateToken string) (*State, error) {
	key := authnStateKeyPrefix + stateToken
	encryptedData, err := s.redis.Get(key).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	data, err := s.encryptor.Decrypt(encryptedData, []byte(key))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	state := &State{}
	err = json.Unmarshal(data, state)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return state, nil
}

// PutAuthorizationCode save an AuthorizationCode to the store.
func (s *Store) PutAuthorizationCode(ctx context.Context, code *AuthorizationCode) error {
	err := code.Validate()
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	key := authorizationCodeKeyPrefix + code.Code
	data, err := json.Marshal(code)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	encryptedData, err := s.encryptor.Encrypt(data, []byte(key))
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}

	expiry := viper.GetDuration("authorization_token_expires_in")
	err = s.redis.Set(key, encryptedData, expiry).Err()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// GetAuthorizationCode retrieves an AuthorizationCode from the store.
func (s *Store) GetAuthorizationCode(ctx context.Context, code string) (*AuthorizationCode, error) {
	key := authorizationCodeKeyPrefix + code
	encryptedData, err := s.redis.Get(key).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	data, err := s.encryptor.Decrypt(encryptedData, []byte(key))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	obj := &AuthorizationCode{}
	err = json.Unmarshal(data, obj)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return obj, nil
}

// DeleteAuthorizationCode deletes an AuthorizationCode from the store.
func (s *Store) DeleteAuthorizationCode(ctx context.Context, code string) error {
	key := authorizationCodeKeyPrefix + code
	deleted, err := s.redis.Del(key).Result()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	if deleted == 0 {
		return errors.New(errors.ErrorNotFound, "")
	}
	return nil
}

// CheckRateLimiter checks a user if it has exceeded authentication rate limiting.
func (s *Store) CheckRateLimiter(ctx context.Context, userID int64) error {
	return s.rateLimiter.Check(fmt.Sprintf("%d/password", userID))
}

// IncrementRateLimiter increments a user's authentication rate limiting.
func (s *Store) IncrementRateLimiter(ctx context.Context, userID int64) error {
	return s.rateLimiter.Increment(fmt.Sprintf("%d/password", userID))
}
