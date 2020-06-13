package user

//FIXME: this file should be moved to another package

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/validator"
	"authcore.io/authcore/pkg/cryptoutil"

	// log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const closedLoopCodePrefix = "closed_loop_code/"

// CloseLoopCodeType is a type enumerating the type for contact code.
type CloseLoopCodeType int32

// Enumerates the CloseLoopCodeType
const (
	ContactCodeCODE  CloseLoopCodeType = 0
	ContactCodeTOKEN CloseLoopCodeType = 1
)

// ClosedLoopCode represents the code for authenticating / verificating a contact.
type ClosedLoopCode struct {
	Key               string
	CodeSentAt        time.Time
	CodeExpireAt      time.Time
	Code              string
	Token             string
	RemainingAttempts int64
}

// Validate validates a ContactCode
func (closedLoopCode *ClosedLoopCode) Validate() error {
	return validator.Validate.Struct(closedLoopCode)
}

// Expiry returns the duration of expiry
func (closedLoopCode *ClosedLoopCode) Expiry() time.Duration {
	return closedLoopCode.CodeExpireAt.Sub(closedLoopCode.CodeSentAt)
}

// GetContactID returns the contact id of the closed loop code
func (closedLoopCode *ClosedLoopCode) GetContactID() (int64, error) {
	if !strings.HasPrefix(closedLoopCode.Key, "contact_id:") {
		return -1, errors.New(errors.ErrorInvalidArgument, "")
	}
	contactID, err := strconv.ParseInt(closedLoopCode.Key[11:], 10, 64)
	if err != nil {
		return -1, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	return contactID, nil
}

// GetSecondFactorID returns the second factor id of the closed loop code
func (closedLoopCode *ClosedLoopCode) GetSecondFactorID() (int64, error) {
	if !strings.HasPrefix(closedLoopCode.Key, "second_factor_id:") {
		return -1, errors.New(errors.ErrorInvalidArgument, "")
	}
	secondFactorID, err := strconv.ParseInt(closedLoopCode.Key[17:], 10, 64)
	if err != nil {
		return -1, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	return secondFactorID, nil
}

// GetSecondFactorValue returns the second factor value of the closed loop code
func (closedLoopCode *ClosedLoopCode) GetSecondFactorValue() (string, error) {
	if !strings.HasPrefix(closedLoopCode.Key, "second_factor_value:") {
		return "", errors.New(errors.ErrorInvalidArgument, "")
	}
	secondFactorValue := closedLoopCode.Key[20:]
	return secondFactorValue, nil
}

// CreateClosedLoopCode creates a new closed loop code.
func (s *Store) CreateClosedLoopCode(ctx context.Context, key string, expiry time.Duration) (*ClosedLoopCode, error) {
	codeLength := int64(viper.GetInt("closed_loop_code_length"))
	remainingAttempts := int64(viper.GetInt("closed_loop_max_attempts"))

	codeSentAt := time.Now()
	codeExpireAt := codeSentAt.Add(expiry)

	code := cryptoutil.RandomCode(codeLength)
	token := cryptoutil.RandomToken32()

	closedLoopCode := &ClosedLoopCode{
		Key:               key,
		CodeSentAt:        codeSentAt,
		CodeExpireAt:      codeExpireAt,
		Code:              code,
		Token:             token,
		RemainingAttempts: remainingAttempts,
	}

	closedLoopCodeJSON, err := json.Marshal(closedLoopCode)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	// Setting the key "closed_loop_code/key/<key>" to be the closed loop code
	closedLoopCodeKey := fmt.Sprintf("%s/key/%s", closedLoopCodePrefix, key)
	err = s.redis.Set(closedLoopCodeKey, closedLoopCodeJSON, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	// Setting the key "closed_loop_code/token/<token>" to point to the key of the closed loop code
	closedLoopCodePointer := fmt.Sprintf("%s/token/%s", closedLoopCodePrefix, token)
	err = s.redis.Set(closedLoopCodePointer, closedLoopCodeKey, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return closedLoopCode, nil
}

// FindClosedLoopCodeByKey finds a closed loop code by key
func (s *Store) FindClosedLoopCodeByKey(ctx context.Context, key string) (*ClosedLoopCode, error) {
	closedLoopCodeKey := fmt.Sprintf("%s/key/%s", closedLoopCodePrefix, key)
	closedLoopCodeJSON, err := s.redis.Get(closedLoopCodeKey).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorNotFound, "")
	}

	closedLoopCode := &ClosedLoopCode{}
	err = json.Unmarshal([]byte(closedLoopCodeJSON), closedLoopCode)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return closedLoopCode, nil
}

// FindClosedLoopCodeByToken finds a contact code by token
func (s *Store) FindClosedLoopCodeByToken(ctx context.Context, token string) (*ClosedLoopCode, error) {
	closedLoopCodePointer := fmt.Sprintf("%s/token/%s", closedLoopCodePrefix, token)
	closedLoopCodeKey, err := s.redis.Get(closedLoopCodePointer).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorNotFound, "")
	}
	closedLoopCodeJSON, err := s.redis.Get(closedLoopCodeKey).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorNotFound, "")
	}

	closedLoopCode := &ClosedLoopCode{}
	err = json.Unmarshal([]byte(closedLoopCodeJSON), closedLoopCode)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return closedLoopCode, nil
}

// GetClosedLoopCodeLastSentAt gets the last timestamp the code is sent for this contact information with this intent
func (s *Store) GetClosedLoopCodeLastSentAt(ctx context.Context, key string) (time.Time, error) {
	closedLoopCode, err := s.FindClosedLoopCodeByKey(ctx, key)
	if err != nil {
		return time.Unix(0, 0), nil
	}
	return closedLoopCode.CodeSentAt, nil
}

// BurnClosedLoopCodeByCode verifies closed loop code and remove the closed loop code from Redis to prevent replay.
func (s *Store) BurnClosedLoopCodeByCode(ctx context.Context, key string, code string) (*ClosedLoopCode, error) {
	closedLoopCode, err := s.FindClosedLoopCodeByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if closedLoopCode.RemainingAttempts == 0 {
		return nil, errors.New(errors.ErrorResourceExhausted, "")
	}
	if closedLoopCode.Code != code {
		s.DecrementClosedLoopCodeRemainingAttempts(key)
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	closedLoopCodeKey := fmt.Sprintf("%s/key/%s", closedLoopCodePrefix, key)
	s.redis.Del(closedLoopCodeKey)
	if closedLoopCode.CodeExpireAt.Before(time.Now()) {
		return nil, errors.New(errors.ErrorDeadlineExceeded, "")
	}
	return closedLoopCode, nil
}

// BurnClosedLoopCodeByToken verifies closed loop token and remove the closed loop code from Redis to prevent replay.
func (s *Store) BurnClosedLoopCodeByToken(ctx context.Context, token string) (*ClosedLoopCode, error) {
	closedLoopCode, err := s.FindClosedLoopCodeByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if closedLoopCode.Token != token {
		return nil, errors.New(errors.ErrorNotFound, "")
	}
	if closedLoopCode.RemainingAttempts == 0 {
		return nil, errors.New(errors.ErrorResourceExhausted, "")
	}

	closedLoopCodeKey := fmt.Sprintf("%s/key/%s", closedLoopCodePrefix, closedLoopCode.Key)
	s.redis.Del(closedLoopCodeKey)
	if closedLoopCode.CodeExpireAt.Before(time.Now()) {
		return nil, errors.New(errors.ErrorDeadlineExceeded, "")
	}
	return closedLoopCode, nil
}

// DecrementClosedLoopCodeRemainingAttempts decreases the number of remaining attempts by one
func (s *Store) DecrementClosedLoopCodeRemainingAttempts(key string) error {
	closedLoopCodeKey := fmt.Sprintf("%s/key/%s", closedLoopCodePrefix, key)

	closedLoopCodeJSON, err := s.redis.Get(closedLoopCodeKey).Result()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}

	closedLoopCode := &ClosedLoopCode{}
	err = json.Unmarshal([]byte(closedLoopCodeJSON), closedLoopCode)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}

	if closedLoopCode.RemainingAttempts > 0 {
		closedLoopCode.RemainingAttempts--
	}

	updatedClosedLoopCodeJSON, err := json.Marshal(closedLoopCode)
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	expiry := closedLoopCode.Expiry()
	err = s.redis.Set(closedLoopCodeKey, updatedClosedLoopCodeJSON, expiry).Err()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// GetClosedLoopCodePartialKeyByContactID returns the close loop code partial code from contact id
func (s *Store) GetClosedLoopCodePartialKeyByContactID(id int64) string {
	return fmt.Sprintf("contact_id:%d", id)
}

// GetClosedLoopCodePartialKeyByUserIDAndContactValue returns the close loop code partial code from contact value
func (s *Store) GetClosedLoopCodePartialKeyByUserIDAndContactValue(id int64, value string) string {
	return fmt.Sprintf("user_id:%d/contact_value:%s", id, value)
}

// GetClosedLoopCodePartialKeyBySecondFactorID returns the close loop code partial code from second factor id
func (s *Store) GetClosedLoopCodePartialKeyBySecondFactorID(id int64) string {
	return fmt.Sprintf("second_factor_id:%d", id)
}

// GetClosedLoopCodePartialKeyBySecondFactorValue returns the close loop code partial code from second factor value
func (s *Store) GetClosedLoopCodePartialKeyBySecondFactorValue(value string) string {
	return fmt.Sprintf("second_factor_value:%s", value)
}
