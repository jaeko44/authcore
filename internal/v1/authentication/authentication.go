package authentication

import (
	"context"
	"encoding/json"

	"authcore.io/authcore/internal/clientapp"
	"authcore.io/authcore/internal/errors"

	"authcore.io/authcore/pkg/cryptoutil"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

// AuthenticationStateKeyPrefix and AuthorizationTokenKeyPrefix are the Redis key prefixs
const (
	AuthenticationStateKeyPrefix                = "authentication_state/"
	ResetPasswordAuthenticationStateKeyPrefix   = "reset_password_authentication_state/"
	AuthorizationTokenKeyPrefix                 = "authorization_token/"
	ResetPasswordTokenKeyPrefix                 = "reset_password_token/"
	ResetPasswordTokenToTemporaryTokenMapPrefix = "reset_password_token_to_temporary_token_map/"
	TemporaryTokenToResetPasswordTokenMapPrefix = "temporary_token_to_reset_password_token_map/"
)

// CreateAuthenticationState creates and saves an unauthenticated State.
func (srv *Service) CreateAuthenticationState(ctx context.Context, clientID string, userID, deviceID int64, challenges []string, pkceChallenge, successRedirectURL string) (*State, error) {
	clientApp, err := clientapp.GetByClientID(clientID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	authState := &State{
		ClientID:           clientApp.ID,
		UserID:             userID,
		DeviceID:           deviceID,
		TemporaryToken:     cryptoutil.RandomToken32(),
		Challenges:         challenges,
		PKCEChallenge:      pkceChallenge,
		SuccessRedirectURL: successRedirectURL,
	}
	err = authState.Validate()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	authStateJSON, err := json.Marshal(authState)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	expiry := viper.GetDuration("authentication_time_limit")
	err = srv.Redis.Set(AuthenticationStateKeyPrefix+authState.TemporaryToken, authStateJSON, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return authState, nil
}

// FindAuthenticationStateByTemporaryToken lookups an unauthenticated session with a temporary token.
func (srv *Service) FindAuthenticationStateByTemporaryToken(ctx context.Context, temporaryToken string) (*State, error) {
	authStateJSON, err := srv.Redis.Get(AuthenticationStateKeyPrefix + temporaryToken).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	authState := &State{}

	err = json.Unmarshal([]byte(authStateJSON), authState)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return authState, nil
}

// UpdateAuthenticationStateChallengesByTemporaryToken updates the challenges of an authentication state by its temporary token
func (srv *Service) UpdateAuthenticationStateChallengesByTemporaryToken(ctx context.Context, temporaryToken string, challenges []string) (*State, error) {
	authStateJSON, err := srv.Redis.Get(AuthenticationStateKeyPrefix + temporaryToken).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	authState := &State{}

	err = json.Unmarshal([]byte(authStateJSON), authState)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	authState.Challenges = challenges

	newAuthStateJSON, err := json.Marshal(authState)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	expiry := viper.GetDuration("authentication_time_limit")

	err = srv.Redis.Set(AuthenticationStateKeyPrefix+authState.TemporaryToken, newAuthStateJSON, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return authState, nil
}

// UpdateUserIDByTemporaryToken updates the user id of an authentication state by its temporary token
func (srv *Service) UpdateUserIDByTemporaryToken(ctx context.Context, temporaryToken string, userID int64) (*State, error) {
	authStateJSON, err := srv.Redis.Get(AuthenticationStateKeyPrefix + temporaryToken).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	authState := &State{}

	err = json.Unmarshal([]byte(authStateJSON), authState)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	if !authState.IsOAuth() {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	authState.UserID = userID

	newAuthStateJSON, err := json.Marshal(authState)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	expiry := viper.GetDuration("authentication_time_limit")

	err = srv.Redis.Set(AuthenticationStateKeyPrefix+authState.TemporaryToken, newAuthStateJSON, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return authState, nil
}

// DeleteAuthenticationStateByTemporaryToken deletes authentication state with a temporary token.
func (srv *Service) DeleteAuthenticationStateByTemporaryToken(ctx context.Context, temporaryToken string) error {
	err := srv.Redis.Del(AuthenticationStateKeyPrefix + temporaryToken).Err()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// CreateAuthorizationToken creates and saves an authorization token.
func (srv *Service) CreateAuthorizationToken(ctx context.Context, userID int64, clientID, codeChallenge string) (*AuthorizationToken, error) {
	clientApp, err := clientapp.GetByClientID(clientID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	authorizationToken := &AuthorizationToken{
		ClientID:           clientApp.ID,
		UserID:             userID,
		AuthorizationToken: cryptoutil.RandomToken32(),
		CodeChallenge:      codeChallenge,
	}

	err = authorizationToken.Validate()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	authorizationTokenJSON, err := json.Marshal(authorizationToken)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	expiry := viper.GetDuration("authorization_token_expires_in")

	err = srv.Redis.Set(AuthorizationTokenKeyPrefix+authorizationToken.AuthorizationToken, authorizationTokenJSON, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return authorizationToken, nil
}

// FindAuthorizationToken lookups an authorization token
func (srv *Service) FindAuthorizationToken(ctx context.Context, sAuthorizationToken string) (*AuthorizationToken, error) {
	authorizationTokenJSON, err := srv.Redis.Get(AuthorizationTokenKeyPrefix + sAuthorizationToken).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	authorizationToken := &AuthorizationToken{}
	err = json.Unmarshal([]byte(authorizationTokenJSON), authorizationToken)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return authorizationToken, nil
}

// DeleteAuthorizationToken deletes an authorization token
func (srv *Service) DeleteAuthorizationToken(ctx context.Context, authorizationToken string) error {
	err := srv.Redis.Del(AuthorizationTokenKeyPrefix + authorizationToken).Err()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// CreateResetPasswordAuthenticationState creates and saves an unauthenticated State.
func (srv *Service) CreateResetPasswordAuthenticationState(ctx context.Context, clientID string, userID int64, deviceID int64, challenges []string) (*State, error) {
	clientApp, err := clientapp.GetByClientID(clientID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	resetPasswordAuthState := &State{
		ClientID:       clientApp.ID,
		UserID:         userID,
		DeviceID:       deviceID,
		TemporaryToken: cryptoutil.RandomToken32(),
		Challenges:     challenges,
	}

	err = resetPasswordAuthState.Validate()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	resetPasswordAuthStateJSON, err := json.Marshal(resetPasswordAuthState)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	expiry := viper.GetDuration("authenticate_reset_password_time_limit")

	err = srv.Redis.Set(ResetPasswordAuthenticationStateKeyPrefix+resetPasswordAuthState.TemporaryToken, resetPasswordAuthStateJSON, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return resetPasswordAuthState, nil
}

// FindResetPasswordAuthenticationStateByTemporaryToken lookups an unauthenticated session for reset password with a temporary token.
func (srv *Service) FindResetPasswordAuthenticationStateByTemporaryToken(ctx context.Context, temporaryToken string) (*State, error) {
	resetPasswordAuthStateJSON, err := srv.Redis.Get(ResetPasswordAuthenticationStateKeyPrefix + temporaryToken).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	resetPasswordAuthState := &State{}

	err = json.Unmarshal([]byte(resetPasswordAuthStateJSON), resetPasswordAuthState)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return resetPasswordAuthState, nil
}

// UpdateResetPasswordStateChallengesByTemporaryToken updates the challenges of an reset password state by its temporary token
func (srv *Service) UpdateResetPasswordStateChallengesByTemporaryToken(ctx context.Context, temporaryToken string, challenges []string) (*State, error) {
	resetPasswordAuthStateJSON, err := srv.Redis.Get(ResetPasswordAuthenticationStateKeyPrefix + temporaryToken).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	resetPasswordAuthState := &State{}

	err = json.Unmarshal([]byte(resetPasswordAuthStateJSON), resetPasswordAuthState)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	resetPasswordAuthState.Challenges = challenges

	newResetPasswordAuthStateJSON, err := json.Marshal(resetPasswordAuthState)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	expiry := viper.GetDuration("authenticate_reset_password_time_limit")

	err = srv.Redis.Set(ResetPasswordAuthenticationStateKeyPrefix+resetPasswordAuthState.TemporaryToken, newResetPasswordAuthStateJSON, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return resetPasswordAuthState, nil
}

// DeleteResetPasswordAuthenticationStateByTemporaryToken deletes authentication state for reset password with a temporary token.
func (srv *Service) DeleteResetPasswordAuthenticationStateByTemporaryToken(ctx context.Context, temporaryToken string) error {
	err := srv.Redis.Del(ResetPasswordAuthenticationStateKeyPrefix + temporaryToken).Err()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}

// CreateResetPasswordToken creates and saves an authorization token for reset password.
// Temporary token is involved for late-comers to lookup.
func (srv *Service) CreateResetPasswordToken(ctx context.Context, userID int64, temporaryToken string) (*ResetPasswordToken, error) {
	resetPasswordToken := &ResetPasswordToken{
		UserID:             userID,
		ResetPasswordToken: cryptoutil.RandomToken32(),
	}

	err := resetPasswordToken.Validate()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	resetPasswordTokenJSON, err := json.Marshal(resetPasswordToken)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	expiry := viper.GetDuration("authorization_token_expires_in")

	err = srv.Redis.Set(ResetPasswordTokenKeyPrefix+resetPasswordToken.ResetPasswordToken, resetPasswordTokenJSON, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	err = srv.Redis.Set(TemporaryTokenToResetPasswordTokenMapPrefix+temporaryToken, resetPasswordToken.ResetPasswordToken, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	err = srv.Redis.Set(ResetPasswordTokenToTemporaryTokenMapPrefix+resetPasswordToken.ResetPasswordToken, temporaryToken, expiry).Err()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return resetPasswordToken, nil
}

// FindResetPasswordTokenByTemporaryToken finds a reset password token from a temporary token.
func (srv *Service) FindResetPasswordTokenByTemporaryToken(ctx context.Context, temporaryToken string) (*ResetPasswordToken, error) {
	sResetPasswordToken, err := srv.Redis.Get(TemporaryTokenToResetPasswordTokenMapPrefix + temporaryToken).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return srv.FindResetPasswordToken(ctx, sResetPasswordToken)
}

// FindResetPasswordToken lookups an authorization token for reset password
func (srv *Service) FindResetPasswordToken(ctx context.Context, sResetPasswordToken string) (*ResetPasswordToken, error) {
	resetPasswordTokenJSON, err := srv.Redis.Get(ResetPasswordTokenKeyPrefix + sResetPasswordToken).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	resetPasswordToken := &ResetPasswordToken{}

	err = json.Unmarshal([]byte(resetPasswordTokenJSON), resetPasswordToken)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return resetPasswordToken, nil
}

// DeleteResetPasswordToken deletes an authorization token for reset password and the corresponding temporary token.
func (srv *Service) DeleteResetPasswordToken(ctx context.Context, resetPasswordToken string) error {
	temporaryToken, err := srv.Redis.Get(ResetPasswordTokenToTemporaryTokenMapPrefix + resetPasswordToken).Result()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	err = srv.Redis.Del(ResetPasswordTokenKeyPrefix + resetPasswordToken).Err()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	err = srv.Redis.Del(TemporaryTokenToResetPasswordTokenMapPrefix + temporaryToken).Err()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	err = srv.Redis.Del(ResetPasswordTokenToTemporaryTokenMapPrefix + resetPasswordToken).Err()
	if err != nil {
		return errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return nil
}
