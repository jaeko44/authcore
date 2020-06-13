package authapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/cryptoutil"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
)

// StartPasswordAuthn get login parameters for user (e.g. salt, scrypt params etc).
func (s *Service) StartPasswordAuthn(ctx context.Context, in *authapi.StartPasswordAuthnRequest) (*authapi.AuthenticationState, error) {
	if in.UserHandle == "" {
		return nil, errors.New(errors.ErrorInvalidArgument, "user handle should not be empty")
	}
	clientID := in.ClientId

	// TODO: If not found, return some salt derived from some server secret
	user, err := s.UserStore.UserByHandle(ctx, in.UserHandle)
	if err != nil {
		return nil, err
	}

	if user.IsCurrentlyLocked() {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	if !user.IsPasswordAuthenticationEnabled() {
		return nil, errors.New(errors.ErrorInvalidArgument, "user has not enabled password authentication")
	}

	challenges := []string{"PASSWORD"}

	// TODO: handle the device ID

	codeChallenge := in.CodeChallenge
	codeChallengeMethod := in.CodeChallengeMethod
	successRedirectURL := in.SuccessRedirectUrl
	if codeChallenge != "" && codeChallengeMethod != "S256" {
		return nil, errors.New(errors.ErrorInvalidArgument, "invalid code challenge method")
	}
	authState, err := s.AuthenticationService.CreateAuthenticationState(ctx, clientID, user.ID, 0, challenges, codeChallenge, successRedirectURL)
	if err != nil {
		return nil, err
	}

	pbChallenges := []authapi.AuthenticationState_ChallengeType{}
	for _, challengeItem := range challenges {
		pbChallenges = append(pbChallenges, authapi.AuthenticationState_ChallengeType(
			authapi.AuthenticationState_ChallengeType_value[challengeItem],
		))
	}

	return &authapi.AuthenticationState{
		TemporaryToken: authState.TemporaryToken,
		Challenges:     pbChallenges,
		PasswordSalt:   user.PasswordSalt(),
	}, nil
}

// PasswordAuthnKeyExchange continues a authentication flow by starting key exchange.
func (s *Service) PasswordAuthnKeyExchange(ctx context.Context, in *authapi.PasswordAuthnKeyExchangeRequest) (*authapi.AuthenticationState, error) {
	challenges := []string{"PASSWORD"}

	authState, err := s.AuthenticationService.FindAuthenticationStateByTemporaryToken(ctx, in.TemporaryToken)
	if err != nil {
		return nil, err
	}

	user, err := s.UserStore.UserByID(ctx, authState.UserID)
	if err != nil {
		return nil, err
	}

	passwordChallenge, err := s.AuthenticationService.NewPasswordChallengeWithUser(ctx, user, in.Message)
	if err != nil {
		return nil, err
	}

	pbChallenges := []authapi.AuthenticationState_ChallengeType{}
	for _, challengeItem := range challenges {
		pbChallenges = append(pbChallenges, authapi.AuthenticationState_ChallengeType(
			authapi.AuthenticationState_ChallengeType_value[challengeItem],
		))
	}
	return &authapi.AuthenticationState{
		TemporaryToken:    authState.TemporaryToken,
		Challenges:        pbChallenges,
		PasswordChallenge: passwordChallenge,
	}, nil
}

// FinishPasswordAuthn completes the password authentication and provides challenge responses to update the state of an authentication flow.
func (s *Service) FinishPasswordAuthn(ctx context.Context, in *authapi.FinishPasswordAuthnRequest) (*authapi.AuthenticationState, error) {
	token := in.TemporaryToken

	authState, err := s.AuthenticationService.FindAuthenticationStateByTemporaryToken(ctx, token)
	if err != nil {
		return nil, err
	}

	skipAuthenticateSecondFactor := false
	var user *user.User
	switch response := in.Response.(type) {
	case *authapi.FinishPasswordAuthnRequest_PasswordResponse:
		// Authenticate method: PASSWORD
		if !containsChallenge(authState.Challenges, "PASSWORD") {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}
		passwordResponse := response.PasswordResponse
		user, err = s.UserStore.UserByID(ctx, authState.UserID)
		if err != nil {
			return nil, err
		}
		err = s.RateLimiters.AuthenticationRateLimiter.Check(fmt.Sprintf("password/authenticate/%d", user.ID))
		if err != nil {
			return nil, errors.New(errors.ErrorUserTemporarilyBlocked, "")
		}
		err = s.AuthenticationService.VerifyPasswordResponseWithUser(ctx, user, passwordResponse.Token, passwordResponse.Confirmation)
		if err != nil {
			s.RateLimiters.AuthenticationRateLimiter.Increment(fmt.Sprintf("password/authenticate/%d", user.ID))
			return nil, err
		}
		log.WithFields(log.Fields{
			"user_id": user.PublicID(),
		}).Info("password authentication succeded")
	default:
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	// Check if the users need to authenticate with second factors
	if !skipAuthenticateSecondFactor {
		secondFactors, err := s.UserStore.FindAllSecondFactorsByUserID(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		if len(*secondFactors) > 0 {
			// Not authenticated
			challenges := []string{}
			pbChallenges := []authapi.AuthenticationState_ChallengeType{}
			for _, secondFactor := range *secondFactors {
				secondFactorType, err := secondFactor.Type.StringV1()
				if err != nil {
					return nil, errors.Wrap(err, errors.ErrorUnknown, "")
				}
				challenges = append(challenges, secondFactorType)
				pbChallenges = append(pbChallenges, authapi.AuthenticationState_ChallengeType(
					authapi.AuthenticationState_ChallengeType_value[secondFactorType],
				))
			}

			_, err = s.AuthenticationService.UpdateAuthenticationStateChallengesByTemporaryToken(ctx, token, challenges)
			if err != nil {
				return nil, err
			}

			return &authapi.AuthenticationState{
				TemporaryToken: token,
				Authenticated:  false,
				Challenges:     pbChallenges,
			}, nil
		}
	}
	// Authenticated
	authorizationToken, err := s.AuthenticationService.CreateAuthorizationToken(ctx, user.ID, authState.ClientID, authState.PKCEChallenge)
	if err != nil {
		return nil, err
	}
	err = s.AuthenticationService.DeleteAuthenticationStateByTemporaryToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &authapi.AuthenticationState{
		Authenticated:       true,
		AuthorizationToken:  authorizationToken.AuthorizationToken,
		AuthenticatedUserId: user.PublicID(),
		RedirectUrl:         authState.SuccessRedirectURL,
	}, nil
}

// StartAuthenticateSecondFactor initiates the second factor authentication.
func (s *Service) StartAuthenticateSecondFactor(ctx context.Context, in *authapi.StartAuthenticateSecondFactorRequest) (*authapi.StartAuthenticateSecondFactorResponse, error) {
	token := in.TemporaryToken
	authState, err := s.AuthenticationService.FindAuthenticationStateByTemporaryToken(ctx, token)
	if err != nil {
		return nil, err
	}
	user, err := s.UserStore.UserByID(ctx, authState.UserID)
	if err != nil {
		return nil, err
	}
	secondFactor := ChallengeTypeToSecondFactorDelegate(in.Challenge)
	secondFactors, err := s.UserStore.FindAllSecondFactorsByUserIDAndType(ctx, user.ID, secondFactor.GetType())
	if err != nil {
		return nil, err
	}

	var cSecondFactors []SecondFactorDelegate
	for _, secondFactor := range *secondFactors {
		sfDelegate := FactorTypeToSecondFactorDelegate(secondFactor.Type)
		cSecondFactor := sfDelegate.ParseAuthenticationSecondFactor(secondFactor)
		cSecondFactors = append(cSecondFactors, cSecondFactor)
	}
	err = startAuthenticateSecondFactors(ctx, s, user, cSecondFactors)
	if err != nil {
		return nil, err
	}

	return &authapi.StartAuthenticateSecondFactorResponse{}, nil
}

// AuthenticateSecondFactor provides challenge responses to update the state of an authentication flow.
func (s *Service) AuthenticateSecondFactor(ctx context.Context, in *authapi.AuthenticateSecondFactorRequest) (*authapi.AuthenticationState, error) {
	token := in.TemporaryToken

	authState, err := s.AuthenticationService.FindAuthenticationStateByTemporaryToken(ctx, token)
	if err != nil {
		return nil, err
	}
	user, err := s.UserStore.UserByID(ctx, authState.UserID)
	if err != nil {
		return nil, err
	}
	secondFactor := ChallengeTypeToSecondFactorDelegate(in.Challenge)
	answer := in.Answer
	secondFactors, err := s.UserStore.FindAllSecondFactorsByUserIDAndType(ctx, user.ID, secondFactor.GetType())
	if err != nil {
		return nil, err
	}

	var cSecondFactors []SecondFactorDelegate
	for _, secondFactor := range *secondFactors {
		sfDelegate := FactorTypeToSecondFactorDelegate(secondFactor.Type)
		cSecondFactor := sfDelegate.ParseAuthenticationSecondFactor(secondFactor)
		cSecondFactors = append(cSecondFactors, cSecondFactor)
	}
	authenticatedSecondFactor, err := authenticateSecondFactors(ctx, s, user, cSecondFactors, answer)
	if err != nil {
		return nil, err
	}
	_, err = s.UserStore.UpdateSecondFactorLastUsedAtByID(ctx, authenticatedSecondFactor.GetID())
	if err != nil {
		return nil, err
	}

	// Authenticated
	authorizationToken, err := s.AuthenticationService.CreateAuthorizationToken(ctx, user.ID, authState.ClientID, authState.PKCEChallenge)
	if err != nil {
		return nil, err
	}
	err = s.AuthenticationService.DeleteAuthenticationStateByTemporaryToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &authapi.AuthenticationState{
		Authenticated:       true,
		AuthorizationToken:  authorizationToken.AuthorizationToken,
		AuthenticatedUserId: user.PublicID(),
		RedirectUrl:         authState.SuccessRedirectURL,
	}, nil
}

// StartResetPasswordAuthentication starts a new authentication flow to reset password.
func (s *Service) StartResetPasswordAuthentication(ctx context.Context, in *authapi.StartResetPasswordAuthenticationRequest) (*authapi.ResetPasswordAuthenticationState, error) {
	if in.UserHandle == "" {
		return nil, errors.New(errors.ErrorInvalidArgument, "user handle should not be empty")
	}
	incomingCtx, _ := metadata.FromIncomingContext(ctx)
	origin := incomingCtx["grpcgateway-origin"][0]

	u, err := s.UserStore.UserByHandle(ctx, in.UserHandle)
	if err != nil {
		return nil, err
	}
	u, err = s.UserStore.IncreaseResetPasswordCount(ctx, u)
	if err != nil {
		return nil, err
	}

	challenges := []string{"CONTACT_TOKEN"}
	clientID := in.ClientId
	resetPasswordAuthState, err := s.AuthenticationService.CreateResetPasswordAuthenticationState(ctx, clientID, u.ID, 0, challenges)
	if err != nil {
		return nil, err
	}
	closedLoopCodeExpiry := viper.GetDuration("contact_reset_password_authentication_expiry")
	resetPasswordRedirectLink := viper.GetString("reset_password_redirect_link")
	// Only replace the link with origin when it contains %s at the beginning
	if strings.HasPrefix(resetPasswordRedirectLink, "%s") {
		resetPasswordRedirectLink = fmt.Sprintf(resetPasswordRedirectLink, origin)
	}
	closedLoopCode, err := s.createClosedLoopCodeForResetPassword(ctx, closedLoopCodeExpiry, resetPasswordAuthState.TemporaryToken)
	if err != nil {
		return nil, err
	}

	if u.Email.Valid && in.UserHandle == u.Email.String {
		log.WithFields(log.Fields{
			"id": u.ID,
		}).Info("send reset password email")
		err = s.EmailService.SendResetPasswordAuthenticationMail(
			ctx,
			origin,
			u.DisplayName(),
			u.Email.String, // email address
			u.Language.String,
			closedLoopCode.Token,
			resetPasswordRedirectLink,
		)
		if err != nil {
			return nil, err
		}
	} else if u.Phone.Valid && in.UserHandle == u.Phone.String {
		log.WithFields(log.Fields{
			"id": u.ID,
		}).Info("send reset password sms")
		err = s.SMSService.SendResetPasswordAuthenticationSMS(
			ctx,
			origin,
			u.DisplayName(),
			u.Phone.String, // phone
			closedLoopCode.Token,
			resetPasswordRedirectLink,
		)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(errors.ErrorInvalidArgument, "could not reset password with this handle")
	}

	pbChallenges := []authapi.ResetPasswordAuthenticationState_ChallengeType{}
	for _, challengeItem := range challenges {
		pbChallenges = append(pbChallenges, authapi.ResetPasswordAuthenticationState_ChallengeType(
			authapi.ResetPasswordAuthenticationState_ChallengeType_value[challengeItem],
		))
	}
	return &authapi.ResetPasswordAuthenticationState{
		TemporaryToken: resetPasswordAuthState.TemporaryToken,
		Challenges:     pbChallenges,
	}, nil
}

// AuthenticateResetPassword provides challenge responses to update the state of an authentication flow for reset password.
func (s *Service) AuthenticateResetPassword(ctx context.Context, in *authapi.AuthenticateResetPasswordRequest) (*authapi.ResetPasswordAuthenticationState, error) {
	// Gets the temporary token
	temporaryToken := in.TemporaryToken
	var inputToken string
	switch response := in.Response.(type) {
	case *authapi.AuthenticateResetPasswordRequest_ContactToken:
		inputToken = response.ContactToken.Token
		// Finds the reset password token from temporary token.
		resetPasswordToken, err := s.AuthenticationService.FindResetPasswordTokenByTemporaryToken(ctx, inputToken)
		if err == nil {
			u, err := s.UserStore.UserByID(ctx, resetPasswordToken.UserID)
			if err != nil {
				return nil, err
			}
			return &authapi.ResetPasswordAuthenticationState{
				Authenticated:       true,
				ResetPasswordToken:  resetPasswordToken.ResetPasswordToken,
				ResetPasswordUserId: u.PublicID(),
			}, nil
		}

		closedLoopCode, err := s.findClosedLoopCodeForResetPasswordByToken(ctx, inputToken)
		if err == nil {
			temporaryToken = closedLoopCode.TemporaryToken
		}
	}

	resetPasswordAuthState, err := s.AuthenticationService.FindResetPasswordAuthenticationStateByTemporaryToken(ctx, temporaryToken)
	if err != nil {
		return nil, err
	}
	u, err := s.UserStore.UserByID(ctx, resetPasswordAuthState.UserID)
	if err != nil {
		return nil, err
	}

	switch response := in.Response.(type) {
	case *authapi.AuthenticateResetPasswordRequest_ContactToken:
		// Authenticate method: CONTACT_TOKEN
		if !containsChallenge(resetPasswordAuthState.Challenges, "CONTACT_TOKEN") {
			return nil, errors.New(errors.ErrorInvalidArgument, "")
		}
		token := response.ContactToken.Token
		_, err = s.burnClosedLoopCodeForResetPasswordByToken(ctx, token)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(errors.ErrorInvalidArgument, "")
	}

	authenticated := true
	challenges := []string{}

	// Check if the users need to authenticate with another challenge.
	// There are no extra steps at the moment.

	if !authenticated {
		// Not authenticated
		pbChallenges := []authapi.ResetPasswordAuthenticationState_ChallengeType{}
		for _, challengeItem := range challenges {
			pbChallenges = append(pbChallenges, authapi.ResetPasswordAuthenticationState_ChallengeType(
				authapi.ResetPasswordAuthenticationState_ChallengeType_value[challengeItem],
			))
		}

		_, err = s.AuthenticationService.UpdateResetPasswordStateChallengesByTemporaryToken(ctx, temporaryToken, challenges)
		if err != nil {
			return nil, err
		}

		return &authapi.ResetPasswordAuthenticationState{
			TemporaryToken: temporaryToken,
			Authenticated:  false,
			Challenges:     pbChallenges,
		}, nil
	}
	// Authenticated
	resetPasswordToken, err := s.AuthenticationService.CreateResetPasswordToken(ctx, u.ID, inputToken)
	if err != nil {
		return nil, err
	}
	err = s.AuthenticationService.DeleteResetPasswordAuthenticationStateByTemporaryToken(ctx, temporaryToken)
	if err != nil {
		return nil, err
	}

	return &authapi.ResetPasswordAuthenticationState{
		Authenticated:       true,
		ResetPasswordToken:  resetPasswordToken.ResetPasswordToken,
		ResetPasswordUserId: u.PublicID(),
	}, nil
}

// AuthenticateResetPasswordSecondFactor provides challenge responses to update the state of an authentication flow for reset password.
func (s *Service) AuthenticateResetPasswordSecondFactor(ctx context.Context, in *authapi.AuthenticateResetPasswordSecondFactorRequest) (*authapi.ResetPasswordAuthenticationState, error) {
	return &authapi.ResetPasswordAuthenticationState{}, nil
}

// CreateAuthorizationToken creates an authorization token for the current user.
func (s *Service) CreateAuthorizationToken(ctx context.Context, in *authapi.CreateAuthorizationTokenRequest) (*authapi.AuthorizationToken, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}
	codeChallenge := in.CodeChallenge
	codeChallengeMethod := in.CodeChallengeMethod
	if codeChallenge != "" && codeChallengeMethod != "S256" {
		return nil, errors.New(errors.ErrorInvalidArgument, "invalid code challenge method")
	}
	clientID := in.ClientId
	authorizationToken, err := s.AuthenticationService.CreateAuthorizationToken(ctx, currentUser.ID, clientID, codeChallenge)
	if err != nil {
		return nil, err
	}
	// TODO: set the device id to a valid id
	_, err = s.SessionStore.CreateSession(ctx, currentUser.ID, int64(0), authorizationToken.ClientID, authorizationToken.AuthorizationToken, false)
	if err != nil {
		return nil, err
	}
	return &authapi.AuthorizationToken{
		AuthorizationToken: authorizationToken.AuthorizationToken,
	}, nil
}

// containsChallenge checks if a challenge is in the allow list of challenges.
func containsChallenge(challenges []string, challenge string) bool {
	for _, challengeAllowed := range challenges {
		if challenge == challengeAllowed {
			return true
		}
	}
	return false
}

// ClosedLoopCodeForResetPassword represents the code for reset password.
type ClosedLoopCodeForResetPassword struct {
	CodeSentAt     time.Time
	CodeExpireAt   time.Time
	Token          string
	TemporaryToken string
}

func (s *Service) createClosedLoopCodeForResetPassword(ctx context.Context, expiry time.Duration, temporaryToken string) (*ClosedLoopCodeForResetPassword, error) {
	// random string = token & key for redis
	// stores the temporary token as well

	codeSentAt := time.Now()
	codeExpireAt := codeSentAt.Add(expiry)

	token := cryptoutil.RandomToken32()

	closedLoopCodeForPassword := &ClosedLoopCodeForResetPassword{
		CodeSentAt:     codeSentAt,
		CodeExpireAt:   codeExpireAt,
		Token:          token,
		TemporaryToken: temporaryToken,
	}

	closedLoopCodeForPasswordJSON, err := json.Marshal(closedLoopCodeForPassword)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	closedLoopCodeForPasswordKey := fmt.Sprintf("closed_loop_code_for_password/token/%s", token)
	err = s.Redis.Set(closedLoopCodeForPasswordKey, closedLoopCodeForPasswordJSON, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	// solely for the test cases.
	closedLoopCodeForPasswordPointerKey := fmt.Sprintf("closed_loop_code_for_password/pointer/%s", temporaryToken)
	err = s.Redis.Set(closedLoopCodeForPasswordPointerKey, token, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return closedLoopCodeForPassword, nil
}

func (s *Service) findClosedLoopCodeForResetPasswordByToken(ctx context.Context, token string) (*ClosedLoopCodeForResetPassword, error) {
	closedLoopCodeForPasswordKey := fmt.Sprintf("closed_loop_code_for_password/token/%s", token)
	closedLoopCodeForPasswordJSON, err := s.Redis.Get(closedLoopCodeForPasswordKey).Result()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorNotFound, "")
	}

	closedLoopCodeForPassword := &ClosedLoopCodeForResetPassword{}
	err = json.Unmarshal([]byte(closedLoopCodeForPasswordJSON), closedLoopCodeForPassword)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return closedLoopCodeForPassword, nil
}

func (s *Service) burnClosedLoopCodeForResetPasswordByToken(ctx context.Context, token string) (*ClosedLoopCodeForResetPassword, error) {
	closedLoopCodeForPassword, err := s.findClosedLoopCodeForResetPasswordByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	closedLoopCodeForPasswordKey := fmt.Sprintf("closed_loop_code_for_password/token/%s", token)
	s.Redis.Del(closedLoopCodeForPasswordKey)
	if closedLoopCodeForPassword.CodeExpireAt.Before(time.Now()) {
		return nil, errors.New(errors.ErrorDeadlineExceeded, "")
	}
	return closedLoopCodeForPassword, nil
}
