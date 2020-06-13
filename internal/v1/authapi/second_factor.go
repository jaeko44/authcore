package authapi

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/cryptoutil"

	"github.com/golang/protobuf/ptypes/timestamp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// StartCreateSecondFactor initiates the create second factor flow.
func (s *Service) StartCreateSecondFactor(ctx context.Context, in *authapi.StartCreateSecondFactorRequest) (*authapi.StartCreateSecondFactorResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}
	info, secondFactor := dispatchStartCreateSecondFactorResponse(in)
	err := secondFactor.StartCreate(ctx, s, info, currentUser)
	if err != nil {
		return nil, err
	}
	return &authapi.StartCreateSecondFactorResponse{}, nil
}

// CreateSecondFactor creates a second factor for the current user.
func (s *Service) CreateSecondFactor(ctx context.Context, in *authapi.CreateSecondFactorRequest) (*authapi.SecondFactor, error) {
	// return nil, errors.New(apiserver.TestError, "")
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}
	info, secondFactor := dispatchCreateSecondFactorResponse(in)
	cSecondFactor, err := secondFactor.Create(ctx, s, info, in.Answer, currentUser)
	if err != nil {
		return nil, err
	}
	pSecondFactor, err := MarshalSecondFactor(cSecondFactor, true)
	if err != nil {
		return nil, err
	}
	return pSecondFactor, nil
}

// ListSecondFactors lists the second factors of the current user.
func (s *Service) ListSecondFactors(ctx context.Context, in *authapi.ListSecondFactorsRequest) (*authapi.ListSecondFactorsResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}
	var err error
	secondFactors := &[]user.SecondFactor{}

	// TODO: "in.Type" is now a string, which a parse function is used to convert it as a ContactType
	// We allow this parameter to be empty to search both _phone_ and _email_ contacts.
	// If we use ContactType directly, sending an empty argument to the API would make it use the default value (which is email),
	// making it unable to search arbitrary type of contacts.
	// Ref: https://gitlab.com/blocksq/authcore/issues/239
	secondFactorType, ok := parseSecondFactorTypeFromString(in.Type)
	if ok {
		secondFactors, err = s.UserStore.FindAllSecondFactorsByUserIDAndType(ctx, currentUser.ID, secondFactorType)
	} else {
		secondFactors, err = s.UserStore.FindAllSecondFactorsByUserID(ctx, currentUser.ID)
	}
	if err != nil {
		return nil, err
	}

	var pbSecondFactors []*authapi.SecondFactor
	for _, secondFactor := range *secondFactors {
		pbSecondFactor, err := MarshalSecondFactor(&secondFactor, false)
		if err != nil {
			return nil, err
		}
		pbSecondFactors = append(pbSecondFactors, pbSecondFactor)
	}

	return &authapi.ListSecondFactorsResponse{
		SecondFactors: pbSecondFactors,
	}, nil
}

// DeleteSecondFactor deletes a second factor from the current user.
func (s *Service) DeleteSecondFactor(ctx context.Context, in *authapi.DeleteSecondFactorRequest) (*authapi.DeleteSecondFactorResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}
	id, err := strconv.ParseInt(in.Id, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	secondFactor, err := s.UserStore.FindSecondFactorByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// The authenticator is not belong to the current user
	if secondFactor == nil || secondFactor.UserID != currentUser.ID {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err = s.UserStore.DeleteSecondFactorByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &authapi.DeleteSecondFactorResponse{}, nil
}

// MarshalSecondFactor marshals a *user.SecondFactor into Protobuf message
func MarshalSecondFactor(in *user.SecondFactor, creating bool) (*authapi.SecondFactor, error) {
	content := FactorTypeToSecondFactorDelegate(in.Type).MarshalContent(in.Content, creating)
	return &authapi.SecondFactor{
		Id:      in.ID,
		UserId:  in.UserID,
		Type:    authapi.SecondFactor_Type(in.Type),
		Content: content,
		CreatedAt: &timestamp.Timestamp{
			Seconds: in.CreatedAt.Unix(),
			Nanos:   int32(in.CreatedAt.Nanosecond()),
		},
		LastUsedAt: &timestamp.Timestamp{
			Seconds: in.LastUsedAt.Unix(),
			Nanos:   int32(in.LastUsedAt.Nanosecond()),
		},
	}, nil
}

func parseSecondFactorTypeFromString(in string) (user.SecondFactorType, bool) {
	switch in {
	case "sms":
		return user.SecondFactorSMS, true
	case "totp":
		return user.SecondFactorTOTP, true
	case "backup_code":
		return user.SecondFactorBackupCode, true
	}
	return user.SecondFactorSMS, false
}

// ChallengeTypeToSecondFactorDelegate returns a SecondFactorDelegate by parsing authapi.AuthenticationState_ChallengeType
func ChallengeTypeToSecondFactorDelegate(challengeType authapi.AuthenticationState_ChallengeType) SecondFactorDelegate {
	switch challengeType {
	case authapi.AuthenticationState_TIME_BASED_ONE_TIME_PASSWORD:
		return TOTPSecondFactor{}
	case authapi.AuthenticationState_SMS_CODE:
		return SMSSecondFactor{}
	case authapi.AuthenticationState_BACKUP_CODE:
		return BackupCodeSecondFactor{}
	default:
		log.Panic("cannot dispatch challenge type")
		return nil
	}
}

// FactorTypeToSecondFactorDelegate returns a SecondFactorDelegate by parsing user.SecondFactorType
func FactorTypeToSecondFactorDelegate(factorType user.SecondFactorType) SecondFactorDelegate {
	switch factorType {
	case user.SecondFactorTOTP:
		return TOTPSecondFactor{}
	case user.SecondFactorSMS:
		return SMSSecondFactor{}
	case user.SecondFactorBackupCode:
		return BackupCodeSecondFactor{}
	default:
		log.Panic("cannot dispatch factor type")
		return nil
	}
}

func dispatchStartCreateSecondFactorResponse(in *authapi.StartCreateSecondFactorRequest) (authapi.SecondFactorInfo, SecondFactorDelegate) {
	switch info := in.Info.(type) {
	case *authapi.StartCreateSecondFactorRequest_SmsInfo:
		return info.SmsInfo, SMSSecondFactor{}
	default:
		log.Panic("cannot dispatch start create second factor response")
		return nil, nil
	}
}

func dispatchCreateSecondFactorResponse(in *authapi.CreateSecondFactorRequest) (authapi.SecondFactorInfo, SecondFactorDelegate) {
	switch info := in.Info.(type) {
	case *authapi.CreateSecondFactorRequest_SmsInfo:
		return info.SmsInfo, SMSSecondFactor{}
	case *authapi.CreateSecondFactorRequest_TotpInfo:
		return info.TotpInfo, TOTPSecondFactor{}
	case *authapi.CreateSecondFactorRequest_BackupCodeInfo:
		return info.BackupCodeInfo, BackupCodeSecondFactor{}
	default:
		log.Panic("cannot dispatch create second factor response")
		return nil, nil
	}
}

// SecondFactorDelegate is an interface implements StartCreate, Create, StartAuthenticate, Authenticate.
type SecondFactorDelegate interface {
	GetID() int64
	GetType() user.SecondFactorType
	ParseAuthenticationSecondFactor(secondFactor user.SecondFactor) SecondFactorDelegate
	MarshalContent(content user.SecondFactorContent, creating bool) *authapi.SecondFactor_Content
	StartCreate(ctx context.Context, s *Service, info authapi.SecondFactorInfo, user *user.User) error
	Create(ctx context.Context, s *Service, info authapi.SecondFactorInfo, answer string, user *user.User) (*user.SecondFactor, error)
	StartAuthenticate(ctx context.Context, s *Service, user *user.User) error
	Authenticate(ctx context.Context, s *Service, answer string, user *user.User) error
}

// SMSSecondFactor is a struct of SMS second factor.
type SMSSecondFactor struct {
	ID          int64
	PhoneNumber string
}

// GetID returns the id of the SMS second factor.
func (sf SMSSecondFactor) GetID() int64 {
	return sf.ID
}

// GetType returns the type of the SMS second factor.
func (SMSSecondFactor) GetType() user.SecondFactorType {
	return user.SecondFactorSMS
}

// ParseAuthenticationSecondFactor parses the user.SecondFactor
func (SMSSecondFactor) ParseAuthenticationSecondFactor(secondFactor user.SecondFactor) SecondFactorDelegate {
	return SMSSecondFactor{
		ID:          secondFactor.ID,
		PhoneNumber: secondFactor.Content.PhoneNumber.String,
	}
}

// MarshalContent marshals the content of the SMS second factor.
func (sf SMSSecondFactor) MarshalContent(content user.SecondFactorContent, creating bool) *authapi.SecondFactor_Content {
	return &authapi.SecondFactor_Content{
		PhoneNumber: content.PhoneNumber.String,
	}
}

// StartCreate initiates the creation of SMS second factor.
func (sf SMSSecondFactor) StartCreate(ctx context.Context, s *Service, info authapi.SecondFactorInfo, user *user.User) error {
	smsInfo, ok := info.(*authapi.SMSInfo)
	if !ok {
		return errors.New(errors.ErrorInvalidArgument, "")
	}
	phoneNumber := smsInfo.PhoneNumber

	err := s.RateLimiters.ContactRateLimiter.Check(fmt.Sprintf("second_factor/sms/create/%d/%s", user.ID, phoneNumber))
	if err != nil {
		return errors.New(errors.ErrorResourceExhausted, "")
	}
	err = s.RateLimiters.SecondFactorRateLimiter.Increment(fmt.Sprintf("second_factor/sms/create/%d/%s", user.ID, phoneNumber))
	if err != nil {
		return errors.New(errors.ErrorResourceExhausted, "")
	}
	closedLoopCodePartialKey := s.UserStore.GetClosedLoopCodePartialKeyBySecondFactorValue(phoneNumber)
	verificationExpiry := viper.GetDuration("contact_verification_expiry_for_phone")
	verificationCode, err := s.UserStore.CreateClosedLoopCode(ctx, closedLoopCodePartialKey, verificationExpiry)
	if err != nil {
		return err
	}

	err = s.SMSService.SendVerificationSMS(
		ctx,
		user.DisplayName(),
		phoneNumber,
		verificationCode.Code,
	)
	if err != nil {
		return err
	}

	return nil
}

// Create creates a SMS second factor.
func (sf SMSSecondFactor) Create(ctx context.Context, s *Service, info authapi.SecondFactorInfo, answer string, u *user.User) (*user.SecondFactor, error) {
	smsInfo, ok := info.(*authapi.SMSInfo)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	phoneNumber := smsInfo.PhoneNumber
	secondFactor := &user.SecondFactor{
		UserID: u.ID,
		Type:   user.SecondFactorSMS,
		Content: user.SecondFactorContent{
			PhoneNumber: db.NullableString(phoneNumber),
		},
	}

	// Additional codes to be verified
	closedLoopCodePartialKey := s.UserStore.GetClosedLoopCodePartialKeyBySecondFactorValue(phoneNumber)
	_, err := s.UserStore.BurnClosedLoopCodeByCode(ctx, closedLoopCodePartialKey, answer)
	if err != nil {
		return nil, err
	}

	return s.UserStore.CreateSecondFactor(ctx, secondFactor)
}

// StartAuthenticate initiates a SMS second factor authentication.
func (sf SMSSecondFactor) StartAuthenticate(ctx context.Context, s *Service, user *user.User) error {
	err := s.RateLimiters.ContactRateLimiter.Check(fmt.Sprintf("contact/authenticate/%d/%d", user.ID, sf.ID))
	if err != nil {
		return errors.New(errors.ErrorResourceExhausted, "")
	}
	err = s.RateLimiters.ContactRateLimiter.Increment(fmt.Sprintf("contact/authenticate/%d/%d", user.ID, sf.ID))
	if err != nil {
		return errors.New(errors.ErrorResourceExhausted, "")
	}

	closedLoopCodePartialKey := s.UserStore.GetClosedLoopCodePartialKeyBySecondFactorID(sf.ID)
	authenticationExpiry := viper.GetDuration("contact_authentication_expiry_for_phone")
	authenticationCode, err := s.UserStore.CreateClosedLoopCode(ctx, closedLoopCodePartialKey, authenticationExpiry)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"id": sf.ID,
	}).Info("send authentication sms")

	err = s.SMSService.SendAuthenticationSMS(
		ctx,
		user.DisplayName(),
		sf.PhoneNumber,
		authenticationCode.Code,
	)
	if err != nil {
		return err
	}

	return nil
}

// Authenticate authenticates a SMS second factor.
func (sf SMSSecondFactor) Authenticate(ctx context.Context, s *Service, answer string, user *user.User) error {
	err := s.RateLimiters.SecondFactorRateLimiter.Check(fmt.Sprintf("authenticate_second_factor/sms/%d", user.ID))
	if err != nil {
		return err
	}
	closedLoopCodePartialKey := s.UserStore.GetClosedLoopCodePartialKeyBySecondFactorID(sf.ID)
	_, err = s.UserStore.BurnClosedLoopCodeByCode(ctx, closedLoopCodePartialKey, answer)
	if err != nil {
		s.RateLimiters.SecondFactorRateLimiter.Increment(fmt.Sprintf("authenticate_second_factor/sms/%d", user.ID))
		return err
	}
	return nil
}

// TOTPSecondFactor is a struct of TOTP second factor.
type TOTPSecondFactor struct {
	ID         int64
	Secret     string
	LastUsedAt time.Time
}

// GetID returns the id of the TOTP second factor.
func (sf TOTPSecondFactor) GetID() int64 {
	return sf.ID
}

// GetType returns the type of the TOTP second factor.
func (TOTPSecondFactor) GetType() user.SecondFactorType {
	return user.SecondFactorTOTP
}

// ParseAuthenticationSecondFactor parses the user.SecondFactor
func (TOTPSecondFactor) ParseAuthenticationSecondFactor(secondFactor user.SecondFactor) SecondFactorDelegate {
	return TOTPSecondFactor{
		ID:         secondFactor.ID,
		Secret:     secondFactor.Content.Secret.String,
		LastUsedAt: secondFactor.LastUsedAt,
	}
}

// MarshalContent marshals the content of the TOTP second factor.
func (sf TOTPSecondFactor) MarshalContent(content user.SecondFactorContent, creating bool) *authapi.SecondFactor_Content {
	return &authapi.SecondFactor_Content{
		Identifier: content.Identifier.String,
	}
}

// StartCreate initiates the creation of TOTP second factor.
func (sf TOTPSecondFactor) StartCreate(ctx context.Context, s *Service, info authapi.SecondFactorInfo, user *user.User) error {
	return errors.New(errors.ErrorInvalidArgument, "")
}

// Create creates a TOTP second factor.
func (TOTPSecondFactor) Create(ctx context.Context, s *Service, info authapi.SecondFactorInfo, answer string, u *user.User) (*user.SecondFactor, error) {
	totpInfo, ok := info.(*authapi.TOTPInfo)
	if !ok {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	secret := totpInfo.Secret
	identifier := totpInfo.Identifier
	secondFactor := &user.SecondFactor{
		UserID: u.ID,
		Type:   user.SecondFactorTOTP,
		Content: user.SecondFactorContent{
			Secret:     db.NullableString(secret),
			Identifier: db.NullableString(identifier),
		},
	}
	if !cryptoutil.ValidateTOTP(answer, &cryptoutil.TOTPAuthenticator{
		TotpSecret: secret,
	}) {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	return s.UserStore.CreateSecondFactor(ctx, secondFactor)
}

// StartAuthenticate initiates a TOTP second factor authentication, which is not an valid action.
func (sf TOTPSecondFactor) StartAuthenticate(ctx context.Context, s *Service, user *user.User) error {
	return errors.New(errors.ErrorInvalidArgument, "")
}

// Authenticate authenticates a TOTP second factor.
func (sf TOTPSecondFactor) Authenticate(ctx context.Context, s *Service, answer string, user *user.User) error {
	err := s.RateLimiters.SecondFactorRateLimiter.Check(fmt.Sprintf("authenticate_second_factor/totp/%d", user.ID))
	if err != nil {
		return errors.New(errors.ErrorUserTemporarilyBlocked, "")
	}
	valid := cryptoutil.ValidateTOTP(answer, AsUtilTOTPAuthenticator(&sf))
	if !valid {
		s.RateLimiters.SecondFactorRateLimiter.Increment(fmt.Sprintf("authenticate_second_factor/totp/%d", user.ID))
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	return nil
}

// AsUtilTOTPAuthenticator converts a *TOTPSecondFactor into *cryptoutil.TOTPAuthenticator
func AsUtilTOTPAuthenticator(in *TOTPSecondFactor) *cryptoutil.TOTPAuthenticator {
	return &cryptoutil.TOTPAuthenticator{
		TotpSecret: in.Secret,
		LastUsedAt: in.LastUsedAt,
	}
}

// BackupCodeSecondFactor is a struct of backup code second factor.
type BackupCodeSecondFactor struct {
	ID           int64
	Secret       string
	UsedCodeMask int64 // the bitmask for the used backup code
}

// GetID returns the id of the backup code second factor.
func (sf BackupCodeSecondFactor) GetID() int64 {
	return sf.ID
}

// GetType returns the type of the backup code second factor.
func (BackupCodeSecondFactor) GetType() user.SecondFactorType {
	return user.SecondFactorBackupCode
}

// ParseAuthenticationSecondFactor parses the user.SecondFactor
func (BackupCodeSecondFactor) ParseAuthenticationSecondFactor(secondFactor user.SecondFactor) SecondFactorDelegate {
	return BackupCodeSecondFactor{
		ID:           secondFactor.ID,
		Secret:       secondFactor.Content.Secret.String,
		UsedCodeMask: secondFactor.Content.UsedCodeMask.Int64,
	}
}

// MarshalContent marshals the content of the backup code second factor.
func (sf BackupCodeSecondFactor) MarshalContent(content user.SecondFactorContent, creating bool) *authapi.SecondFactor_Content {
	used := int64(0)
	codes := []string{}
	for counter := 0; counter < 10; counter++ {
		if content.UsedCodeMask.Int64&(1<<uint64(counter)) != 0 {
			used++
		}
		if creating {
			code := cryptoutil.GetBackupCodePin(content.Secret.String, uint64(counter))
			codes = append(codes, code)
		}
	}
	return &authapi.SecondFactor_Content{
		Used:  used,
		Codes: codes,
	}
}

// StartCreate initiates the creation of backup code second factor.
func (sf BackupCodeSecondFactor) StartCreate(ctx context.Context, s *Service, info authapi.SecondFactorInfo, user *user.User) error {
	return errors.New(errors.ErrorInvalidArgument, "")
}

// Create creates a backup code second factor.
func (BackupCodeSecondFactor) Create(ctx context.Context, s *Service, info authapi.SecondFactorInfo, answer string, u *user.User) (*user.SecondFactor, error) {
	secondFactors, err := s.UserStore.FindAllSecondFactorsByUserIDAndType(ctx, u.ID, user.SecondFactorBackupCode)
	if err != nil {
		return nil, err
	}
	if len(*secondFactors) > 0 {
		return nil, errors.New(errors.ErrorAlreadyExists, "")
	}
	secret := cryptoutil.RandomBackupCodeSecret()
	secondFactor := &user.SecondFactor{
		UserID: u.ID,
		Type:   user.SecondFactorBackupCode,
		Content: user.SecondFactorContent{
			Secret:       db.NullableString(secret),
			UsedCodeMask: db.NullableInt64(int64(0)),
		},
	}
	return s.UserStore.CreateSecondFactor(ctx, secondFactor)
}

// StartAuthenticate initiates a backup code second factor authentication, which is not an valid action.
func (sf BackupCodeSecondFactor) StartAuthenticate(ctx context.Context, s *Service, user *user.User) error {
	return errors.New(errors.ErrorInvalidArgument, "")
}

// Authenticate authenticates a backup code second factor.
func (sf BackupCodeSecondFactor) Authenticate(ctx context.Context, s *Service, answer string, user *user.User) error {
	err := s.RateLimiters.SecondFactorRateLimiter.Check(fmt.Sprintf("authenticate_second_factor/backup_code/%d", user.ID))
	if err != nil {
		return errors.New(errors.ErrorUserTemporarilyBlocked, "")
	}
	usedCodeMask, err := cryptoutil.ValidateBackupCodes(answer, sf.Secret, sf.UsedCodeMask, uint64(10))
	if err != nil {
		s.RateLimiters.SecondFactorRateLimiter.Increment(fmt.Sprintf("authenticate_second_factor/backup_code/%d", user.ID))
		return errors.Wrap(err, errors.ErrorUnauthenticated, "")
	}
	_, err = s.UserStore.UpdateSecondFactorUsedCodeMaskByID(ctx, sf.ID, usedCodeMask)
	if err != nil {
		return err
	}
	return nil
}

func startAuthenticateSecondFactors(ctx context.Context, s *Service, user *user.User, secondFactors []SecondFactorDelegate) error {
	if len(secondFactors) == 0 {
		return errors.New(errors.ErrorInvalidArgument, "")
	}
	var err error
	for _, secondFactor := range secondFactors {
		err = secondFactor.StartAuthenticate(ctx, s, user)
		if err == nil {
			return nil
		}
	}
	return errors.New(errors.ErrorInvalidArgument, "")
}

func authenticateSecondFactors(ctx context.Context, s *Service, user *user.User, secondFactors []SecondFactorDelegate, answer string) (SecondFactorDelegate, error) {
	if len(secondFactors) == 0 {
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}
	var err error
	for _, secondFactor := range secondFactors {
		err = secondFactor.Authenticate(ctx, s, answer, user)
		if err == nil {
			return secondFactor, nil
		}
	}
	return nil, err
}
