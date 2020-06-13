package authapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/internal/v1/authentication"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/cryptoutil"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestStartPasswordAuthn(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Get Password Parameters
	email := "carol@example.com"

	req := &authapi.StartPasswordAuthnRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)

	if assert.Nil(t, err) {
		assert.NotEmpty(t, res.PasswordSalt)
		assert.Len(t, res.PasswordSalt, 32)
		assert.Equal(t, "_Jb4pAuatq5rrwdNRGRqW-PhlqzNR1pYtp1N5YWEn7s", base64.RawURLEncoding.EncodeToString(res.PasswordSalt))
		assert.Equal(t, "", res.AuthenticatedUserId)
		assert.Len(t, res.TemporaryToken, 43)
	} else {
		return
	}
}

func TestStartPasswordAuthnWithoutClientID(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Get Password Parameters
	email := "carol@example.com"

	req := &authapi.StartPasswordAuthnRequest{
		UserHandle: email,
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)

	// should not error as fallback to default client id
	if assert.Nil(t, err) {
		assert.NotEmpty(t, res.PasswordSalt)
		assert.Len(t, res.PasswordSalt, 32)
		assert.Equal(t, "_Jb4pAuatq5rrwdNRGRqW-PhlqzNR1pYtp1N5YWEn7s", base64.RawURLEncoding.EncodeToString(res.PasswordSalt))
		assert.Equal(t, "", res.AuthenticatedUserId)
		assert.Len(t, res.TemporaryToken, 43)
	} else {
		return
	}
}

func TestPasswordAuthn(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Get Password Parameters
	email := "carol@example.com"

	req := &authapi.StartPasswordAuthnRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.Nil(t, err)

	// 2. Do key exchange
	password := []byte("password")
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	clientIdentity, serverIdentity := authentication.GetIdentity()

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, password, res.PasswordSalt, nil)

	req2 := &authapi.PasswordAuthnKeyExchangeRequest{
		Message:        message,
		TemporaryToken: res.TemporaryToken,
	}

	res2, err := srv.PasswordAuthnKeyExchange(context.Background(), req2)

	if assert.Nil(t, err) {
		assert.Equal(t, "", res2.AuthenticatedUserId)
		assert.Len(t, res2.TemporaryToken, 43)
		assert.Len(t, res2.Challenges, 1)
		assert.Equal(t, authapi.AuthenticationState_PASSWORD, res2.Challenges[0])
		assert.NotEmpty(t, res2.PasswordChallenge.Token)
		assert.NotEmpty(t, res2.PasswordChallenge.Message)

	} else {
		return
	}

	challengeToken := res2.PasswordChallenge.Token
	message = res2.PasswordChallenge.Message

	// 3. Finish confirmation
	secret, err := state.Finish(res2.PasswordChallenge.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	req3 := &authapi.FinishPasswordAuthnRequest{
		TemporaryToken: res2.TemporaryToken,
		Response: &authapi.FinishPasswordAuthnRequest_PasswordResponse{
			PasswordResponse: &authapi.PasswordResponse{
				Token:        challengeToken,
				Confirmation: confirmation,
			},
		},
	}

	res3, err := srv.FinishPasswordAuthn(context.Background(), req3)

	if assert.Nil(t, err) {
		assert.Len(t, res3.AuthorizationToken, 43)
		assert.Equal(t, true, res3.Authenticated)
		assert.Equal(t, "2", res3.AuthenticatedUserId)
		assert.Len(t, res3.Challenges, 0)
		assert.Nil(t, res3.PasswordChallenge)
	}

	req4 := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN,
		Token:     res3.AuthorizationToken,
	}
	res4, err := srv.CreateAccessToken(context.Background(), req4)

	if assert.Nil(t, err) {
		assert.NotEmpty(t, res4.AccessToken)
		assert.NotEmpty(t, res4.RefreshToken)
		assert.NotEqual(t, res3.AuthorizationToken, res4.RefreshToken)
		assert.Equal(t, authapi.AccessToken_BEARER, res4.TokenType)
		assert.Equal(t, int64(28800), res4.ExpiresIn)
	}

	// Calling update again after creating access token should fail

	_, err = srv.FinishPasswordAuthn(context.Background(), req3)
	assert.Error(t, err)

	// Calling CreateAccessToken again should fail

	_, err = srv.CreateAccessToken(context.Background(), req4)
	assert.Error(t, err)
}

func TestStartPasswordAuthnNoUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	email := "no-user@example.com"

	req := &authapi.StartPasswordAuthnRequest{
		UserHandle: email,
	}

	_, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.Error(t, err)
}

func TestPasswordAuthnKeyExchangeMissingTemporaryKey(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	req := &authapi.PasswordAuthnKeyExchangeRequest{}

	_, err := srv.PasswordAuthnKeyExchange(context.Background(), req)
	assert.Error(t, err)
}

func TestFinishPasswordAuthnMissingTemporaryKey(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	req := &authapi.FinishPasswordAuthnRequest{}

	_, err := srv.FinishPasswordAuthn(context.Background(), req)
	assert.Error(t, err)
}

func TestPasswordAuthenticationMissingStart(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Get Password Parameters
	email := "carol@example.com"

	req := &authapi.StartPasswordAuthnRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)

	password := []byte("password")
	spake2, err := authentication.NewSPAKE2Plus()
	clientIdentity, serverIdentity := authentication.GetIdentity()

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, password, res.PasswordSalt, nil)

	req2 := &authapi.PasswordAuthnKeyExchangeRequest{
		Message:        message,
		TemporaryToken: res.TemporaryToken,
	}

	res2, err := srv.PasswordAuthnKeyExchange(context.Background(), req2)

	if assert.Nil(t, err) {
		assert.Equal(t, "", res2.AuthenticatedUserId)
		assert.Len(t, res2.TemporaryToken, 43)
		assert.Len(t, res2.Challenges, 1)
		assert.Equal(t, authapi.AuthenticationState_PASSWORD, res2.Challenges[0])
		assert.NotEmpty(t, res2.PasswordChallenge.Token)
		assert.NotEmpty(t, res2.PasswordChallenge.Message)
	} else {
		return
	}

	challengeToken := res2.PasswordChallenge.Token
	message = res2.PasswordChallenge.Message

	// 3. Finish confirmation with wrong / no temporary token
	secret, err := state.Finish(res2.PasswordChallenge.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	req3 := &authapi.FinishPasswordAuthnRequest{
		TemporaryToken: "XXX",
		Response: &authapi.FinishPasswordAuthnRequest_PasswordResponse{
			PasswordResponse: &authapi.PasswordResponse{
				Token:        challengeToken,
				Confirmation: confirmation,
			},
		},
	}
	_, err = srv.FinishPasswordAuthn(context.Background(), req3)
	assert.Error(t, err)
}

// Users should be able to sign in if they authenticate with an active TOTP authenticator
func TestTOTPAuthentication(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Get Password Parameters
	email := "factor@example.com"

	req := &authapi.StartPasswordAuthnRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.Nil(t, err)

	// 2. Do key exchange
	password := []byte("password")
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	clientIdentity, serverIdentity := authentication.GetIdentity()

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, password, res.PasswordSalt, nil)

	req2 := &authapi.PasswordAuthnKeyExchangeRequest{
		Message:        message,
		TemporaryToken: res.TemporaryToken,
	}

	res2, err := srv.PasswordAuthnKeyExchange(context.Background(), req2)
	assert.Nil(t, err)

	challengeToken := res2.PasswordChallenge.Token
	message = res2.PasswordChallenge.Message

	// 3. Finish confirmation
	secret, err := state.Finish(res2.PasswordChallenge.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	req3 := &authapi.FinishPasswordAuthnRequest{
		TemporaryToken: res2.TemporaryToken,
		Response: &authapi.FinishPasswordAuthnRequest_PasswordResponse{
			PasswordResponse: &authapi.PasswordResponse{
				Token:        challengeToken,
				Confirmation: confirmation,
			},
		},
	}

	res3, err := srv.FinishPasswordAuthn(context.Background(), req3)

	if assert.Nil(t, err) {
		assert.Len(t, res3.TemporaryToken, 43)
		assert.Equal(t, false, res3.Authenticated)
	}

	// 4. Tries to create access token - Fails however as the user is not authenticated yet
	req4 := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN,
		Token:     res2.TemporaryToken,
	}
	_, err = srv.CreateAccessToken(context.Background(), req4)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}

	// 5. Authenticates with TOTP
	req5 := &authapi.AuthenticateSecondFactorRequest{
		TemporaryToken: res2.TemporaryToken,
		Challenge:      authapi.AuthenticationState_TIME_BASED_ONE_TIME_PASSWORD,
		Answer:         cryptoutil.GetTOTPPin("THISISAWEAKTOTPSECRETFORTESTSXX2", time.Now()),
	}
	res5, err := srv.AuthenticateSecondFactor(context.Background(), req5)

	if assert.Nil(t, err) {
		assert.Len(t, res5.AuthorizationToken, 43)
		assert.Equal(t, true, res5.Authenticated)
		assert.Equal(t, "3", res5.AuthenticatedUserId)
		assert.Len(t, res5.Challenges, 0)
	}

	// 6. Creates access token
	req6 := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN,
		Token:     res5.AuthorizationToken,
	}
	res6, err := srv.CreateAccessToken(context.Background(), req6)

	if assert.Nil(t, err) {
		assert.NotEmpty(t, res6.AccessToken)
		assert.NotEmpty(t, res6.RefreshToken)
		assert.NotEqual(t, res5.TemporaryToken, res6.RefreshToken)
		assert.Equal(t, authapi.AccessToken_BEARER, res6.TokenType)
		assert.Equal(t, int64(28800), res6.ExpiresIn)
	}

	// 7. Tries to authenticate again - Fails as the user is authenticated already
	_, err = srv.FinishPasswordAuthn(context.Background(), req3)
	assert.Error(t, err)

	// 8. Tries to create access token again - Fails as the user has created access token already
	_, err = srv.CreateAccessToken(context.Background(), req6)
	assert.Error(t, err)
}

// Users should not be able to sign in if they authenticate with an inactive TOTP authenticator
func TestTOTPAuthenticationUsingInactiveAuthenticator(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	email := "factor@example.com"
	inactiveTOTPSecret := "THISISAWEAKTOTPSECRETFORTESTSXX3"

	// 1. Get Password Parameters
	req := &authapi.StartPasswordAuthnRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.Nil(t, err)

	// 2. Do key exchange
	password := []byte("password")
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	clientIdentity, serverIdentity := authentication.GetIdentity()

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, password, res.PasswordSalt, nil)

	req2 := &authapi.PasswordAuthnKeyExchangeRequest{
		Message:        message,
		TemporaryToken: res.TemporaryToken,
	}

	res2, err := srv.PasswordAuthnKeyExchange(context.Background(), req2)
	assert.Nil(t, err)

	challengeToken := res2.PasswordChallenge.Token
	message = res2.PasswordChallenge.Message

	// 3. Finish confirmation
	secret, err := state.Finish(res2.PasswordChallenge.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	req3 := &authapi.FinishPasswordAuthnRequest{
		TemporaryToken: res2.TemporaryToken,
		Response: &authapi.FinishPasswordAuthnRequest_PasswordResponse{
			PasswordResponse: &authapi.PasswordResponse{
				Token:        challengeToken,
				Confirmation: confirmation,
			},
		},
	}

	_, err = srv.FinishPasswordAuthn(context.Background(), req3)

	assert.Nil(t, err)

	// 4. Tries to create access token - Fails however as the user is not authenticated yet
	req4 := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN,
		Token:     res2.TemporaryToken,
	}
	_, err = srv.CreateAccessToken(context.Background(), req4)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}

	// 5. Authenticates with TOTP with an inactive TOTP token
	req5 := &authapi.AuthenticateSecondFactorRequest{
		TemporaryToken: res2.TemporaryToken,
		Challenge:      authapi.AuthenticationState_TIME_BASED_ONE_TIME_PASSWORD,
		Answer:         cryptoutil.GetTOTPPin(inactiveTOTPSecret, time.Now()),
	}
	_, err = srv.AuthenticateSecondFactor(context.Background(), req5)

	assert.Error(t, err)
}

// A user should not be able to start authentication (get password parameters) with a non-primary email address
func TestStartPasswordAuthnWithNonprimaryContacts(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	email := "bob_the_third@example.com"

	req := &authapi.StartPasswordAuthnRequest{
		UserHandle: email,
	}

	_, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.Error(t, err)
}

// Users should be able to sign in if they authenticate with SMS
func TestSMSAuthentication(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	userID := int64(5)
	email := "smith@example.com"

	// 1. Get Password Parameters
	req := &authapi.StartPasswordAuthnRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.Nil(t, err)

	// 2. Do key exchange
	password := []byte("password")
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	clientIdentity, serverIdentity := authentication.GetIdentity()

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, password, res.PasswordSalt, nil)

	req2 := &authapi.PasswordAuthnKeyExchangeRequest{
		Message:        message,
		TemporaryToken: res.TemporaryToken,
	}

	res2, err := srv.PasswordAuthnKeyExchange(context.Background(), req2)
	assert.Nil(t, err)

	challengeToken := res2.PasswordChallenge.Token
	message = res2.PasswordChallenge.Message

	// 3. Finish confirmation
	secret, err := state.Finish(res2.PasswordChallenge.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	req3 := &authapi.FinishPasswordAuthnRequest{
		TemporaryToken: res2.TemporaryToken,
		Response: &authapi.FinishPasswordAuthnRequest_PasswordResponse{
			PasswordResponse: &authapi.PasswordResponse{
				Token:        challengeToken,
				Confirmation: confirmation,
			},
		},
	}

	res3, err := srv.FinishPasswordAuthn(context.Background(), req3)

	if assert.Nil(t, err) {
		assert.Len(t, res3.TemporaryToken, 43)
		assert.Equal(t, false, res3.Authenticated)
	}

	// 4. Tries to create access token - Fails however as the user is not authenticated yet
	req4 := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN,
		Token:     res2.TemporaryToken,
	}
	_, err = srv.CreateAccessToken(context.Background(), req4)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}

	// 5. Generates a SMS code
	req5 := &authapi.StartAuthenticateSecondFactorRequest{
		TemporaryToken: res2.TemporaryToken,
		Challenge:      authapi.AuthenticationState_SMS_CODE,
	}
	_, err = srv.StartAuthenticateSecondFactor(context.Background(), req5)

	secondFactors, err := srv.UserStore.FindAllSecondFactorsByUserIDAndType(context.Background(), userID, user.SecondFactorSMS)
	assert.Nil(t, err)
	secondFactor := (*secondFactors)[0]

	closedLoopCodePartialKey := srv.UserStore.GetClosedLoopCodePartialKeyBySecondFactorID(secondFactor.ID)
	closedLoopCode, err := srv.UserStore.FindClosedLoopCodeByKey(context.Background(), closedLoopCodePartialKey)
	assert.Nil(t, err)

	// 6. Authenticates with SMS
	req6 := &authapi.AuthenticateSecondFactorRequest{
		TemporaryToken: res2.TemporaryToken,
		Challenge:      authapi.AuthenticationState_SMS_CODE,
		Answer:         closedLoopCode.Code,
	}
	res6, err := srv.AuthenticateSecondFactor(context.Background(), req6)

	if assert.Nil(t, err) {
		assert.Len(t, res6.AuthorizationToken, 43)
		assert.Equal(t, true, res6.Authenticated)
		assert.Equal(t, "5", res6.AuthenticatedUserId)
		assert.Len(t, res6.Challenges, 0)
	}

	// 7. Creates access token
	req7 := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN,
		Token:     res6.AuthorizationToken,
	}
	res7, err := srv.CreateAccessToken(context.Background(), req7)

	if assert.Nil(t, err) {
		assert.NotEmpty(t, res7.AccessToken)
		assert.NotEmpty(t, res7.RefreshToken)
		assert.NotEqual(t, res6.AuthorizationToken, res7.RefreshToken)
		assert.Equal(t, authapi.AccessToken_BEARER, res7.TokenType)
		assert.Equal(t, int64(28800), res7.ExpiresIn)
	}
}

// Users should not be able to sign in if there are too much fail attempts when authenticating with SMS
func TestSMSAuthenticationTooMuchFailAttempts(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	viper.Set("closed_loop_max_attempts", "1")

	userID := int64(5)
	email := "smith@example.com"

	// 1. Get Password Parameters
	req := &authapi.StartPasswordAuthnRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.Nil(t, err)

	// 2. Do key exchange
	password := []byte("password")
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	clientIdentity, serverIdentity := authentication.GetIdentity()

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, password, res.PasswordSalt, nil)

	req2 := &authapi.PasswordAuthnKeyExchangeRequest{
		Message:        message,
		TemporaryToken: res.TemporaryToken,
	}

	res2, err := srv.PasswordAuthnKeyExchange(context.Background(), req2)
	assert.Nil(t, err)

	challengeToken := res2.PasswordChallenge.Token
	message = res2.PasswordChallenge.Message

	// 3. Finish confirmation
	secret, err := state.Finish(res2.PasswordChallenge.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	req3 := &authapi.FinishPasswordAuthnRequest{
		TemporaryToken: res2.TemporaryToken,
		Response: &authapi.FinishPasswordAuthnRequest_PasswordResponse{
			PasswordResponse: &authapi.PasswordResponse{
				Token:        challengeToken,
				Confirmation: confirmation,
			},
		},
	}

	res3, err := srv.FinishPasswordAuthn(context.Background(), req3)

	if assert.Nil(t, err) {
		assert.Len(t, res3.TemporaryToken, 43)
		assert.Equal(t, false, res3.Authenticated)
	}

	// 4. Generates a SMS code
	req4 := &authapi.StartAuthenticateSecondFactorRequest{
		TemporaryToken: res3.TemporaryToken,
		Challenge:      authapi.AuthenticationState_SMS_CODE,
	}
	_, err = srv.StartAuthenticateSecondFactor(context.Background(), req4)
	assert.Nil(t, err)

	secondFactors, err := srv.UserStore.FindAllSecondFactorsByUserIDAndType(context.Background(), userID, user.SecondFactorSMS)
	assert.Nil(t, err)
	secondFactor := (*secondFactors)[0]

	closedLoopCodePartialKey := srv.UserStore.GetClosedLoopCodePartialKeyBySecondFactorID(secondFactor.ID)
	closedLoopCode, err := srv.UserStore.FindClosedLoopCodeByKey(context.Background(), closedLoopCodePartialKey)
	assert.Nil(t, err)

	// 5. Input a wrong verification code
	req5 := &authapi.AuthenticateSecondFactorRequest{
		TemporaryToken: res2.TemporaryToken,
		Challenge:      authapi.AuthenticationState_SMS_CODE,
		Answer:         "0000000000",
	}
	_, err = srv.AuthenticateSecondFactor(context.Background(), req5)

	assert.Error(t, err)

	// 6. Authenticates with SMS (fails as there are too much invalid requests)
	req6 := &authapi.AuthenticateSecondFactorRequest{
		TemporaryToken: res2.TemporaryToken,
		Challenge:      authapi.AuthenticationState_SMS_CODE,
		Answer:         closedLoopCode.Code,
	}
	_, err = srv.AuthenticateSecondFactor(context.Background(), req6)

	assert.Error(t, err)
}

// Users should not be able to sign in if the SMS authentication details is expired
func TestSMSAuthenticationAfterExpiry(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	viper.Set("contact_authentication_expiry_for_phone", "1s")

	userID := int64(5)
	email := "smith@example.com"

	// 1. Get Password Parameters
	req := &authapi.StartPasswordAuthnRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.Nil(t, err)

	// 2. Do key exchange
	password := []byte("password")
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	clientIdentity, serverIdentity := authentication.GetIdentity()

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, password, res.PasswordSalt, nil)

	req2 := &authapi.PasswordAuthnKeyExchangeRequest{
		Message:        message,
		TemporaryToken: res.TemporaryToken,
	}

	res2, err := srv.PasswordAuthnKeyExchange(context.Background(), req2)
	assert.Nil(t, err)

	challengeToken := res2.PasswordChallenge.Token
	message = res2.PasswordChallenge.Message

	// 3. Finish confirmation
	secret, err := state.Finish(res2.PasswordChallenge.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	req3 := &authapi.FinishPasswordAuthnRequest{
		TemporaryToken: res2.TemporaryToken,
		Response: &authapi.FinishPasswordAuthnRequest_PasswordResponse{
			PasswordResponse: &authapi.PasswordResponse{
				Token:        challengeToken,
				Confirmation: confirmation,
			},
		},
	}

	res3, err := srv.FinishPasswordAuthn(context.Background(), req3)

	if assert.Nil(t, err) {
		assert.Len(t, res3.TemporaryToken, 43)
		assert.Equal(t, false, res3.Authenticated)
	}

	// 4. Generates a SMS code
	req4 := &authapi.StartAuthenticateSecondFactorRequest{
		TemporaryToken: res2.TemporaryToken,
		Challenge:      authapi.AuthenticationState_SMS_CODE,
	}
	_, err = srv.StartAuthenticateSecondFactor(context.Background(), req4)
	assert.Nil(t, err)

	secondFactors, err := srv.UserStore.FindAllSecondFactorsByUserIDAndType(context.Background(), userID, user.SecondFactorSMS)
	assert.Nil(t, err)
	secondFactor := (*secondFactors)[0]

	closedLoopCodePartialKey := srv.UserStore.GetClosedLoopCodePartialKeyBySecondFactorID(secondFactor.ID)
	closedLoopCode, err := srv.UserStore.FindClosedLoopCodeByKey(context.Background(), closedLoopCodePartialKey)
	assert.Nil(t, err)

	// 5. Authenticates with SMS
	req5 := &authapi.AuthenticateSecondFactorRequest{
		TemporaryToken: res2.TemporaryToken,
		Challenge:      authapi.AuthenticationState_SMS_CODE,
		Answer:         closedLoopCode.Code,
	}
	time.Sleep(2 * time.Second)
	_, err = srv.AuthenticateSecondFactor(context.Background(), req5)

	assert.Error(t, err)
}

// Users should be able to sign in if they authenticate with backup code
func TestBackupCodeAuthentication(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	email := "factor@example.com"

	// 1. Get Password Parameters
	req := &authapi.StartPasswordAuthnRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.Nil(t, err)

	// 2. Do key exchange
	password := []byte("password")
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	clientIdentity, serverIdentity := authentication.GetIdentity()

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, password, res.PasswordSalt, nil)

	req2 := &authapi.PasswordAuthnKeyExchangeRequest{
		Message:        message,
		TemporaryToken: res.TemporaryToken,
	}

	res2, err := srv.PasswordAuthnKeyExchange(context.Background(), req2)
	assert.Nil(t, err)

	challengeToken := res2.PasswordChallenge.Token
	message = res2.PasswordChallenge.Message

	// 3. Finish confirmation
	secret, err := state.Finish(res2.PasswordChallenge.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	req3 := &authapi.FinishPasswordAuthnRequest{
		TemporaryToken: res2.TemporaryToken,
		Response: &authapi.FinishPasswordAuthnRequest_PasswordResponse{
			PasswordResponse: &authapi.PasswordResponse{
				Token:        challengeToken,
				Confirmation: confirmation,
			},
		},
	}

	res3, err := srv.FinishPasswordAuthn(context.Background(), req3)

	if assert.Nil(t, err) {
		assert.Len(t, res3.TemporaryToken, 43)
		assert.Equal(t, false, res3.Authenticated)
	}

	// 4. Tries to create access token - Fails however as the user is not authenticated yet
	req4 := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN,
		Token:     res3.TemporaryToken,
	}
	_, err = srv.CreateAccessToken(context.Background(), req4)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}

	// 5. Authenticates with backup code
	answer, err := hotp.GenerateCodeCustom("THISISASECRETFORBACKUPCODETESTSX", uint64(6), hotp.ValidateOpts{
		Digits:    otp.DigitsEight,
		Algorithm: otp.AlgorithmSHA1,
	})
	assert.NoError(t, err)
	req5 := &authapi.AuthenticateSecondFactorRequest{
		TemporaryToken: res2.TemporaryToken,
		Challenge:      authapi.AuthenticationState_BACKUP_CODE,
		Answer:         answer,
	}
	res5, err := srv.AuthenticateSecondFactor(context.Background(), req5)

	if assert.Nil(t, err) {
		assert.Len(t, res5.AuthorizationToken, 43)
		assert.Equal(t, true, res5.Authenticated)
		assert.Equal(t, "3", res5.AuthenticatedUserId)
		assert.Len(t, res5.Challenges, 0)
	}

	// 6. Creates access token
	req6 := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN,
		Token:     res5.AuthorizationToken,
	}
	res6, err := srv.CreateAccessToken(context.Background(), req6)

	if assert.Nil(t, err) {
		assert.NotEmpty(t, res6.AccessToken)
		assert.NotEmpty(t, res6.RefreshToken)
		assert.NotEqual(t, res5.AuthorizationToken, res6.RefreshToken)
		assert.Equal(t, authapi.AccessToken_BEARER, res6.TokenType)
		assert.Equal(t, int64(28800), res6.ExpiresIn)
	}
}

// Users should not be able to initiate authentication if they are currently locked
func TestAuthenticateLockedUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	email := "benny@example.com"

	// 1. Get Password Parameters
	req := &authapi.StartPasswordAuthnRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}

	_, err := srv.StartPasswordAuthn(context.Background(), req)
	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

// Users should be able to reset password with a happy path
func TestAuthenticateResetPassword(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	email := "bob@example.com"

	// 1. starts the reset request
	req := &authapi.StartResetPasswordAuthenticationRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}
	res, err := srv.StartResetPasswordAuthentication(ctx, req)
	assert.Nil(t, err)
	assert.Len(t, res.TemporaryToken, 43)
	assert.Len(t, res.Challenges, 1)

	// 2. inputs the email verification code
	closedLoopCodePointerKey := fmt.Sprintf("closed_loop_code_for_password/pointer/%s", res.TemporaryToken)
	closedLoopCodeToken, err := srv.Redis.Get(closedLoopCodePointerKey).Result()

	req2 := &authapi.AuthenticateResetPasswordRequest{
		Response: &authapi.AuthenticateResetPasswordRequest_ContactToken{
			ContactToken: &authapi.ContactToken{
				Token: closedLoopCodeToken,
			},
		},
	}
	res2, err := srv.AuthenticateResetPassword(ctx, req2)
	assert.Nil(t, err)
	assert.True(t, res2.Authenticated)
	assert.Len(t, res2.ResetPasswordToken, 43)

	// 3. resets the password
	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte("new_password"), clientIdentity, serverIdentity, spake2)
	assert.Nil(t, err)

	req3 := &authapi.ResetPasswordRequest{
		Token: res2.ResetPasswordToken,
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
	}
	_, err = srv.ResetPassword(ctx, req3)
	assert.Nil(t, err)

	// 4. could not resets the password again as the request is expired
	req4 := &authapi.ResetPasswordRequest{
		Token: res2.ResetPasswordToken,
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
	}
	_, err = srv.ResetPassword(ctx, req4)
	assert.NotNil(t, err)
}

func TestStartResetPasswordWithoutClientID(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	email := "bob@example.com"

	// 1. starts the reset request
	req := &authapi.StartResetPasswordAuthenticationRequest{
		UserHandle: email,
	}
	res, err := srv.StartResetPasswordAuthentication(ctx, req)

	// should not error as fallback to default client id
	if assert.NoError(t, err) {
		assert.Len(t, res.TemporaryToken, 43)
		assert.Len(t, res.Challenges, 1)
	}
}

func TestPasswordAuthnWithPKCE(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Get Password Parameters
	email := "carol@example.com"

	req := &authapi.StartPasswordAuthnRequest{
		ClientId:            "authcore-io",
		UserHandle:          email,
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "S256",
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.Nil(t, err)

	// 2. Do key exchange
	password := []byte("password")
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	clientIdentity, serverIdentity := authentication.GetIdentity()

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, password, res.PasswordSalt, nil)

	req2 := &authapi.PasswordAuthnKeyExchangeRequest{
		Message:        message,
		TemporaryToken: res.TemporaryToken,
	}

	res2, err := srv.PasswordAuthnKeyExchange(context.Background(), req2)

	if assert.Nil(t, err) {
		assert.Equal(t, "", res2.AuthenticatedUserId)
		assert.Len(t, res2.TemporaryToken, 43)
		assert.Len(t, res2.Challenges, 1)
		assert.Equal(t, authapi.AuthenticationState_PASSWORD, res2.Challenges[0])
		assert.NotEmpty(t, res2.PasswordChallenge.Token)
		assert.NotEmpty(t, res2.PasswordChallenge.Message)
	} else {
		return
	}

	challengeToken := res2.PasswordChallenge.Token
	message = res2.PasswordChallenge.Message

	// 3. Finish confirmation
	secret, err := state.Finish(res2.PasswordChallenge.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	req3 := &authapi.FinishPasswordAuthnRequest{
		TemporaryToken: res2.TemporaryToken,
		Response: &authapi.FinishPasswordAuthnRequest_PasswordResponse{
			PasswordResponse: &authapi.PasswordResponse{
				Token:        challengeToken,
				Confirmation: confirmation,
			},
		},
	}
	res3, err := srv.FinishPasswordAuthn(context.Background(), req3)
	if assert.Nil(t, err) {
		assert.Len(t, res3.AuthorizationToken, 43)
		assert.Equal(t, true, res3.Authenticated)
		assert.Equal(t, "2", res3.AuthenticatedUserId)
		assert.Len(t, res3.Challenges, 0)
		assert.Nil(t, res3.PasswordChallenge)
	}

	req4 := &authapi.CreateAccessTokenRequest{
		GrantType:    authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN,
		Token:        res3.AuthorizationToken,
		CodeVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
	}
	_, err = srv.CreateAccessToken(context.Background(), req4)
	assert.Nil(t, err)
}

func TestPasswordAuthnWithPKCEWithWrongVerifier(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Get Password Parameters
	email := "carol@example.com"

	req := &authapi.StartPasswordAuthnRequest{
		ClientId:            "authcore-io",
		UserHandle:          email,
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "S256",
	}

	res, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.Nil(t, err)

	// 2. Do key exchange
	password := []byte("password")
	spake2, err := authentication.NewSPAKE2Plus()
	assert.Nil(t, err)

	clientIdentity, serverIdentity := authentication.GetIdentity()

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, password, res.PasswordSalt, nil)

	req2 := &authapi.PasswordAuthnKeyExchangeRequest{
		Message:        message,
		TemporaryToken: res.TemporaryToken,
	}

	res2, err := srv.PasswordAuthnKeyExchange(context.Background(), req2)

	if assert.Nil(t, err) {
		assert.Equal(t, "", res2.AuthenticatedUserId)
		assert.Len(t, res2.TemporaryToken, 43)
		assert.Len(t, res2.Challenges, 1)
		assert.Equal(t, authapi.AuthenticationState_PASSWORD, res2.Challenges[0])
		assert.NotEmpty(t, res2.PasswordChallenge.Token)
		assert.NotEmpty(t, res2.PasswordChallenge.Message)
	} else {
		return
	}

	challengeToken := res2.PasswordChallenge.Token
	message = res2.PasswordChallenge.Message

	// 3. Finish confirmation
	secret, err := state.Finish(res2.PasswordChallenge.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	req3 := &authapi.FinishPasswordAuthnRequest{
		TemporaryToken: res2.TemporaryToken,
		Response: &authapi.FinishPasswordAuthnRequest_PasswordResponse{
			PasswordResponse: &authapi.PasswordResponse{
				Token:        challengeToken,
				Confirmation: confirmation,
			},
		},
	}
	res3, err := srv.FinishPasswordAuthn(context.Background(), req3)
	if assert.Nil(t, err) {
		assert.Len(t, res3.AuthorizationToken, 43)
		assert.Equal(t, true, res3.Authenticated)
		assert.Equal(t, "2", res3.AuthenticatedUserId)
		assert.Len(t, res3.Challenges, 0)
		assert.Nil(t, res3.PasswordChallenge)
	}

	// 4. The request should fail as the code verifier is incorrect
	req4 := &authapi.CreateAccessTokenRequest{
		GrantType:    authapi.CreateAccessTokenRequest_AUTHORIZATION_TOKEN,
		Token:        res3.AuthorizationToken,
		CodeVerifier: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}
	_, err = srv.CreateAccessToken(context.Background(), req4)
	assert.Error(t, err)
}

// The authorization token should not become refresh token after it is expired.
// https://gitlab.com/blocksq/authcore/issues/204
func TestPasswordAuthnWithAuthrTokenExpired(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	viper.Set("authorization_token_expires_in", "1s")

	// 1. Get Password Parameters
	email := "carol@example.com"

	req := &authapi.StartPasswordAuthnRequest{
		ClientId:   "authcore-io",
		UserHandle: email,
	}
	res, err := srv.StartPasswordAuthn(context.Background(), req)
	assert.NoError(t, err)

	// 2. Do key exchange
	password := []byte("password")
	spake2, err := authentication.NewSPAKE2Plus()
	assert.NoError(t, err)

	clientIdentity, serverIdentity := authentication.GetIdentity()

	state, message, err := spake2.StartClient(clientIdentity, serverIdentity, password, res.PasswordSalt, nil)
	assert.NoError(t, err)

	req2 := &authapi.PasswordAuthnKeyExchangeRequest{
		Message:        message,
		TemporaryToken: res.TemporaryToken,
	}
	res2, err := srv.PasswordAuthnKeyExchange(context.Background(), req2)
	assert.NoError(t, err)

	challengeToken := res2.PasswordChallenge.Token
	message = res2.PasswordChallenge.Message

	// 3. Finish confirmation
	secret, err := state.Finish(res2.PasswordChallenge.Message)
	assert.Nil(t, err)

	confirmation := secret.GetConfirmation()

	req3 := &authapi.FinishPasswordAuthnRequest{
		TemporaryToken: res2.TemporaryToken,
		Response: &authapi.FinishPasswordAuthnRequest_PasswordResponse{
			PasswordResponse: &authapi.PasswordResponse{
				Token:        challengeToken,
				Confirmation: confirmation,
			},
		},
	}
	res3, err := srv.FinishPasswordAuthn(context.Background(), req3)
	assert.NoError(t, err)

	// Sleep until the authorization token is expired
	time.Sleep(2 * time.Second)

	// 4. Create an access token from refresh token (#204)
	req4 := &authapi.CreateAccessTokenRequest{
		GrantType: authapi.CreateAccessTokenRequest_REFRESH_TOKEN,
		Token:     res3.AuthorizationToken,
	}
	_, err = srv.CreateAccessToken(context.Background(), req4)
	assert.Error(t, err)
}
