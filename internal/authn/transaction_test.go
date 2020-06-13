package authn

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	"authcore.io/authcore/internal/authn/idp"
	"authcore.io/authcore/internal/authn/verifier"
	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/email"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/sms"
	"authcore.io/authcore/internal/template"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/cryptoutil"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	testutil.DBSetUp()
	code := m.Run()
	testutil.DBTearDown()
	os.Exit(code)
}

func tcForTest() (*TransactionController, func()) {
	config.InitDefaults()
	viper.Set("secret_key_base", "855edf399835e9c9deb61877c1a76bf14eed7c35a167e10ff1b7d43db4363268")
	viper.Set("base_path", "../..")
	viper.Set("applications.app.name", "app")
	viper.Set("applications.app.allowed_callback_urls", []string{"https://example.com"})
	config.InitConfig()

	testutil.FixturesSetUp()
	d := db.NewDBFromConfig()
	redis := testutil.RedisForTest()
	encryptor := testutil.EncryptorForTest()
	store := NewStore(redis, encryptor)
	userStore := user.NewStore(d, redis, encryptor)
	sessionStore := session.NewStore(d, redis, userStore)
	templateStore := template.NewStore(d)
	smsService := sms.NewService(templateStore)
	emailService := email.NewService(templateStore)
	tc := NewTransactionController(d, store, userStore, sessionStore)
	tc.RegisterVerifier(verifier.SMSOTP, verifier.SMSOTPVerifierFactory(smsService, redis))
	tc.RegisterVerifier(verifier.ResetLink, verifier.ResetLinkVerifierFactory(smsService, emailService, redis))
	tc.RegisterIDP(new(mockIDP))

	return tc, func() {
		d.Close()
		viper.Reset()
		redis.FlushAll()
	}
}

func TestPrimaryPassword(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	salt, err := base64.StdEncoding.DecodeString("/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=")
	if !assert.NoError(t, err) {
		return
	}

	// Step 1
	state, err := tc.StartPrimary(ctx, "app", "carol@example.com", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Equal(t, int64(2), state.UserID)
	assert.Empty(t, state.PasswordVerifierState)
	assert.Equal(t, "https://example.com/", state.RedirectURI)
	assert.Equal(t, "", state.PKCEChallenge)
	assert.Equal(t, "", state.PKCEChallengeMethod)
	assert.Equal(t, "", state.AuthorizationCode)
	assert.Equal(t, []string{"password"}, state.Factors)
	assert.Equal(t, "spake2plus", state.PasswordMethod)
	assert.Equal(t, salt, state.PasswordSalt)
	assert.False(t, state.PasswordVerified)
	state.ClearFactors()

	// Check state
	state2, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, state, state2)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), salt, nil)
	assert.NoError(t, err)
	challenge, err := tc.RequestPassword(ctx, state.StateToken, message)
	assert.NoError(t, err)
	assert.NotEmpty(t, challenge)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := sk.GetConfirmation()

	// Check state
	state3, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, "PRIMARY", state3.Status)
	assert.NotEmpty(t, state3.PasswordVerifierState)
	assert.False(t, state3.PasswordVerified)

	// Step 2
	state4, err := tc.VerifyPassword(ctx, state.StateToken, confirmation)
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", state4.Status)
	assert.NotEmpty(t, state4.AuthorizationCode)
	assert.True(t, state4.PasswordVerified)

	// Check authorization code
	code, err := tc.store.GetAuthorizationCode(ctx, state4.AuthorizationCode)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), code.UserID)
	assert.True(t, code.PasswordVerified)

	// Step 3
	sess, err := tc.ExchangeSession(ctx, "app", "https://example.com/", state4.AuthorizationCode, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), sess.UserID)
	assert.Equal(t, "app", sess.ClientID.String)
}

func TestPrimaryPasswordWithPKCE(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	salt, err := base64.StdEncoding.DecodeString("/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=")
	if !assert.NoError(t, err) {
		return
	}

	hashVerifier := sha256.Sum256([]byte("test"))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hashVerifier[:])
	// Step 1
	state, err := tc.StartPrimary(ctx, "app", "carol@example.com", "https://example.com/", "S256", codeChallenge, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Equal(t, int64(2), state.UserID)
	assert.Empty(t, state.PasswordVerifierState)
	assert.Equal(t, "https://example.com/", state.RedirectURI)
	assert.Equal(t, codeChallenge, state.PKCEChallenge)
	assert.Equal(t, "S256", state.PKCEChallengeMethod)
	assert.Equal(t, "", state.AuthorizationCode)
	assert.Equal(t, []string{"password"}, state.Factors)
	assert.Equal(t, "spake2plus", state.PasswordMethod)
	assert.Equal(t, salt, state.PasswordSalt)
	state.ClearFactors()

	// Check state
	state2, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, state, state2)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), salt, nil)
	assert.NoError(t, err)
	challenge, err := tc.RequestPassword(ctx, state.StateToken, message)
	assert.NoError(t, err)
	assert.NotEmpty(t, challenge)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := sk.GetConfirmation()

	// Check state
	state3, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, "PRIMARY", state3.Status)
	assert.NotEmpty(t, state3.PasswordVerifierState)

	// Step 2
	state4, err := tc.VerifyPassword(ctx, state.StateToken, confirmation)
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", state4.Status)
	assert.NotEmpty(t, state4.AuthorizationCode)

	// Check authorization code
	code, err := tc.store.GetAuthorizationCode(ctx, state4.AuthorizationCode)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), code.UserID)

	// Step 3
	sess, err := tc.ExchangeSession(ctx, "app", "https://example.com/", state4.AuthorizationCode, "test")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), sess.UserID)
	assert.Equal(t, "app", sess.ClientID.String)
	assert.True(t, sess.LastPasswordVerifiedAt.Valid)
}

func TestPrimaryNoUser(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	_, err := tc.StartPrimary(ctx, "app", "no-user@example.com", "https://example.com/", "", "", "")
	assert.Error(t, err)
	assert.True(t, errors.IsKind(err, errors.ErrorNotFound))
}

func TestPrimaryPasswordWithClientState(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	salt, err := base64.StdEncoding.DecodeString("/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=")
	if !assert.NoError(t, err) {
		return
	}
	clientState := "random_client_state"

	// Step 1
	state, err := tc.StartPrimary(ctx, "app", "carol@example.com", "https://example.com/", "", "", clientState)
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Equal(t, int64(2), state.UserID)
	assert.Empty(t, state.PasswordVerifierState)
	assert.Equal(t, "https://example.com/", state.RedirectURI)
	assert.Equal(t, "", state.PKCEChallenge)
	assert.Equal(t, "", state.PKCEChallengeMethod)
	assert.Equal(t, clientState, state.ClientState)
	assert.Equal(t, "", state.AuthorizationCode)
	assert.Equal(t, []string{"password"}, state.Factors)
	assert.Equal(t, "spake2plus", state.PasswordMethod)
	assert.Equal(t, salt, state.PasswordSalt)
	state.ClearFactors()

	// Check state
	state2, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, state, state2)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), salt, nil)
	assert.NoError(t, err)
	challenge, err := tc.RequestPassword(ctx, state.StateToken, message)
	assert.NoError(t, err)
	assert.NotEmpty(t, challenge)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := sk.GetConfirmation()

	// Check state
	state3, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, "PRIMARY", state3.Status)
	assert.NotEmpty(t, state3.PasswordVerifierState)
	assert.Equal(t, clientState, state.ClientState)

	// Step 2
	state4, err := tc.VerifyPassword(ctx, state.StateToken, confirmation)
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", state4.Status)
	assert.NotEmpty(t, state4.AuthorizationCode)
	assert.Equal(t, clientState, state.ClientState)

	// Check authorization code
	code, err := tc.store.GetAuthorizationCode(ctx, state4.AuthorizationCode)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), code.UserID)

	// Step 3
	sess, err := tc.ExchangeSession(ctx, "app", "https://example.com/", state4.AuthorizationCode, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), sess.UserID)
	assert.Equal(t, "app", sess.ClientID.String)
}

func TestPrimaryPasswordWithInvalidPKCE(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	salt, err := base64.StdEncoding.DecodeString("/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=")
	if !assert.NoError(t, err) {
		return
	}

	codeChallenge := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" // sha256("test")
	// Invalid code challenge method
	state, err := tc.StartPrimary(ctx, "app", "carol@example.com", "https://example.com/", "", codeChallenge, "")
	assert.Error(t, err)
	assert.Nil(t, state)

	// Invalid code verifier
	// Step 1
	state, err = tc.StartPrimary(ctx, "app", "carol@example.com", "https://example.com/", "S256", codeChallenge, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Equal(t, int64(2), state.UserID)
	assert.Empty(t, state.PasswordVerifierState)
	assert.Equal(t, "https://example.com/", state.RedirectURI)
	assert.Equal(t, codeChallenge, state.PKCEChallenge)
	assert.Equal(t, "S256", state.PKCEChallengeMethod)
	assert.Equal(t, "", state.AuthorizationCode)
	assert.Equal(t, []string{"password"}, state.Factors)
	assert.Equal(t, "spake2plus", state.PasswordMethod)
	assert.Equal(t, salt, state.PasswordSalt)
	state.ClearFactors()

	// Check state
	state2, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, state, state2)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), salt, nil)
	assert.NoError(t, err)
	challenge, err := tc.RequestPassword(ctx, state.StateToken, message)
	assert.NoError(t, err)
	assert.NotEmpty(t, challenge)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := sk.GetConfirmation()

	// Check state
	state3, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, "PRIMARY", state3.Status)
	assert.NotEmpty(t, state3.PasswordVerifierState)

	// Step 2
	state4, err := tc.VerifyPassword(ctx, state.StateToken, confirmation)
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", state4.Status)
	assert.NotEmpty(t, state4.AuthorizationCode)

	// Check authorization code
	code, err := tc.store.GetAuthorizationCode(ctx, state4.AuthorizationCode)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), code.UserID)

	// Step 3
	sess, err := tc.ExchangeSession(ctx, "app", "https://example.com/", state4.AuthorizationCode, "invalid")
	assert.Error(t, err)
	assert.Nil(t, sess)
}

func TestPrimaryInvalidRequestPassword(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartPrimary(ctx, "app", "carol@example.com", "https://example.com/", "", "", "")
	assert.NoError(t, err)

	// Empty request message
	_, err = tc.RequestPassword(ctx, state.StateToken, make([]byte, 32))
	assert.Error(t, err)

	// Zero request message
	_, err = tc.RequestPassword(ctx, state.StateToken, []byte{0x00, 0x00, 0x00})
	assert.Error(t, err)

	// Check state
	state2, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, "PRIMARY", state2.Status)
}

func TestPrimaryIncorrectPassword(t *testing.T) {
	viper.Set("authentication_rate_limit_count", 2)
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	salt, err := base64.StdEncoding.DecodeString("/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=")
	if !assert.NoError(t, err) {
		return
	}

	// Step 1
	state, err := tc.StartPrimary(ctx, "app", "carol@example.com", "https://example.com/", "", "", "")
	assert.NoError(t, err)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("wrong_password"), salt, nil)
	assert.NoError(t, err)
	challenge, err := tc.RequestPassword(ctx, state.StateToken, message)
	assert.NoError(t, err)
	assert.NotEmpty(t, challenge)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := sk.GetConfirmation()

	// Step 2
	_, err = tc.VerifyPassword(ctx, state.StateToken, confirmation)
	assert.Error(t, err)
	assert.True(t, errors.IsKind(err, errors.ErrorPermissionDenied))

	// Check state
	state2, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, "PRIMARY", state2.Status)

	// Again
	_, err = tc.VerifyPassword(ctx, state.StateToken, confirmation)
	assert.Error(t, err)
	assert.True(t, errors.IsKind(err, errors.ErrorPermissionDenied))

	// Should block user
	_, err = tc.VerifyPassword(ctx, state.StateToken, confirmation)
	assert.Error(t, err)
	assert.True(t, errors.IsKind(err, errors.ErrorUserTemporarilyBlocked))

	// New state
	state3, err := tc.StartPrimary(ctx, "app", "carol@example.com", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.Equal(t, "BLOCKED", state3.Status)

	// Should block even with correct password
	cs, message, err = spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), salt, nil)
	assert.NoError(t, err)
	challenge, err = tc.RequestPassword(ctx, state3.StateToken, message)
	assert.Error(t, err)
	assert.True(t, errors.IsKind(err, errors.ErrorPermissionDenied))
}

func TestStateTokenNotFound(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	_, err := tc.VerifyPassword(ctx, "invalid", []byte{})
	assert.Error(t, err)
	assert.True(t, errors.IsKind(err, errors.ErrorPermissionDenied))
}

func TestPrimaryAndTOTP(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartPrimary(ctx, "app", "factor@example.com", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Equal(t, int64(3), state.UserID)
	assert.False(t, state.PasswordVerified)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), state.PasswordSalt, nil)
	assert.NoError(t, err)
	challenge, err := tc.RequestPassword(ctx, state.StateToken, message)
	assert.NoError(t, err)
	assert.NotEmpty(t, challenge)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := sk.GetConfirmation()

	// Step 2
	state2, err := tc.VerifyPassword(ctx, state.StateToken, confirmation)
	assert.NoError(t, err)
	assert.Equal(t, "MFA_REQUIRED", state2.Status)
	assert.Equal(t, []string{"totp", "backup_code"}, state2.Factors)
	assert.Empty(t, state2.AuthorizationCode)
	assert.True(t, state2.PasswordVerified)

	// Step 3
	code := cryptoutil.GetTOTPPin("THISISAWEAKTOTPSECRETFORTESTSXX2", time.Now())
	state3, err := tc.VerifyMFA(ctx, state.StateToken, "totp", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", state3.Status)
	assert.NotEmpty(t, state3.AuthorizationCode)
	assert.True(t, state3.PasswordVerified)

	// Check authorization code
	ac, err := tc.store.GetAuthorizationCode(ctx, state3.AuthorizationCode)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), ac.UserID)
	assert.True(t, ac.PasswordVerified)

	// Step 4
	sess, err := tc.ExchangeSession(ctx, "app", "https://example.com/", state3.AuthorizationCode, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), sess.UserID)
	assert.Equal(t, "app", sess.ClientID.String)
	assert.True(t, sess.LastPasswordVerifiedAt.Valid)
}

func TestPrimaryAndSMS(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartPrimary(ctx, "app", "smith@example.com", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Equal(t, int64(5), state.UserID)
	assert.False(t, state.PasswordVerified)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), state.PasswordSalt, nil)
	assert.NoError(t, err)
	challenge, err := tc.RequestPassword(ctx, state.StateToken, message)
	assert.NoError(t, err)
	assert.NotEmpty(t, challenge)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := sk.GetConfirmation()

	// Step 2
	state2, err := tc.VerifyPassword(ctx, state.StateToken, confirmation)
	assert.NoError(t, err)
	assert.Equal(t, "MFA_REQUIRED", state2.Status)
	assert.Empty(t, state2.AuthorizationCode)
	assert.True(t, state2.PasswordVerified)

	// Step 3
	challenge, err = tc.RequestMFA(ctx, state.StateToken, "sms_otp", []byte{})
	assert.NoError(t, err)
	assert.Empty(t, challenge)

	// Check state
	state3, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, "MFA_REQUIRED", state3.Status)
	assert.Equal(t, "sms_otp", state3.MFAMethod)
	assert.NotEmpty(t, state3.MFAVerifierState)

	smsCode := make(map[string]interface{})
	err = json.Unmarshal([]byte(state3.MFAVerifierState), &smsCode)
	assert.NoError(t, err)
	code := smsCode["code"].(string)

	// Step 4
	state4, err := tc.VerifyMFA(ctx, state.StateToken, "sms_otp", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", state4.Status)
	assert.NotEmpty(t, state4.AuthorizationCode)
	assert.True(t, state4.PasswordVerified)

	// Check authorization code
	ac, err := tc.store.GetAuthorizationCode(ctx, state4.AuthorizationCode)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), ac.UserID)
	assert.True(t, ac.PasswordVerified)

	// Step 5
	sess, err := tc.ExchangeSession(ctx, "app", "https://example.com/", state4.AuthorizationCode, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), sess.UserID)
	assert.Equal(t, "app", sess.ClientID.String)
	assert.True(t, sess.LastPasswordVerifiedAt.Valid)
}

func TestPrimaryAndBackupCode(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartPrimary(ctx, "app", "factor@example.com", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Equal(t, int64(3), state.UserID)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), state.PasswordSalt, nil)
	assert.NoError(t, err)
	challenge, err := tc.RequestPassword(ctx, state.StateToken, message)
	assert.NoError(t, err)
	assert.NotEmpty(t, challenge)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := sk.GetConfirmation()

	// Step 2
	state2, err := tc.VerifyPassword(ctx, state.StateToken, confirmation)
	assert.NoError(t, err)
	assert.Equal(t, "MFA_REQUIRED", state2.Status)
	assert.Empty(t, state2.AuthorizationCode)

	// Step 3
	code, _ := hotp.GenerateCodeCustom("THISISASECRETFORBACKUPCODETESTSX", uint64(6), hotp.ValidateOpts{
		Digits:    otp.DigitsEight,
		Algorithm: otp.AlgorithmSHA1,
	})
	state3, err := tc.VerifyMFA(ctx, state.StateToken, "backup_code", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", state3.Status)
	assert.NotEmpty(t, state3.AuthorizationCode)

	// Check authorization code
	ac, err := tc.store.GetAuthorizationCode(ctx, state3.AuthorizationCode)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), ac.UserID)

	// Reuse backup code

	// Step 1
	state4, err := tc.StartPrimary(ctx, "app", "factor@example.com", "https://example.com/", "", "", "")
	assert.NoError(t, err)

	// Step 2
	cs, message, err = spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), state.PasswordSalt, nil)
	assert.NoError(t, err)
	challenge, err = tc.RequestPassword(ctx, state4.StateToken, message)
	assert.NoError(t, err)
	sk, err = cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation = sk.GetConfirmation()

	// Step 2
	_, err = tc.VerifyPassword(ctx, state4.StateToken, confirmation)
	assert.NoError(t, err)

	// Step 3
	_, err = tc.VerifyMFA(ctx, state4.StateToken, "backup_code", []byte(code))
	assert.Error(t, err)
	assert.True(t, errors.IsKind(err, errors.ErrorPermissionDenied))

	// Check authorization code
	state5, err := tc.store.GetState(ctx, state4.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, "MFA_REQUIRED", state5.Status)
}

func TestPrimaryAndInvalidMFA(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartPrimary(ctx, "app", "factor@example.com", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Equal(t, int64(3), state.UserID)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), state.PasswordSalt, nil)
	assert.NoError(t, err)
	challenge, err := tc.RequestPassword(ctx, state.StateToken, message)
	assert.NoError(t, err)
	assert.NotEmpty(t, challenge)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := sk.GetConfirmation()

	// Step 2
	state2, err := tc.VerifyPassword(ctx, state.StateToken, confirmation)
	assert.NoError(t, err)
	assert.Equal(t, "MFA_REQUIRED", state2.Status)
	assert.Empty(t, state2.AuthorizationCode)

	// Invalid TOTP
	_, err = tc.VerifyMFA(ctx, state.StateToken, "totp", []byte{})
	assert.Error(t, err)

	// Invalid TOTP
	_, err = tc.VerifyMFA(ctx, state.StateToken, "totp", []byte{0x00, 0x00, 0x00})
	assert.Error(t, err)

	// Non-existent factor
	_, err = tc.VerifyMFA(ctx, state.StateToken, "backup_code", []byte{})
	assert.Error(t, err)

	// Invalid factor name
	_, err = tc.VerifyMFA(ctx, state.StateToken, "invalid", []byte{})
	assert.Error(t, err)

	// Check state
	state3, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, "MFA_REQUIRED", state3.Status)
}

func TestVerifyPasswordInvalidState(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartPrimary(ctx, "app", "carol@example.com", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Equal(t, int64(2), state.UserID)

	// VerifyPassword without RequestPassword
	_, err = tc.VerifyPassword(ctx, state.StateToken, []byte{})
	assert.Error(t, err)

	// Check state
	state2, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Empty(t, state2.PasswordVerifierState)
}

func TestVerifyTOTPInvalidState(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartPrimary(ctx, "app", "factor@example.com", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Equal(t, int64(3), state.UserID)

	// VerifyMFA without verifying password
	code := cryptoutil.GetTOTPPin("THISISAWEAKTOTPSECRETFORTESTSXX2", time.Now())
	_, err = tc.VerifyMFA(ctx, state.StateToken, "totp", []byte(code))
	assert.Error(t, err)

	// Check state
	state2, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, "PRIMARY", state.Status)
	assert.Empty(t, state2.PasswordVerifierState)
	assert.Empty(t, state2.MFAMethod)
	assert.Empty(t, state2.MFAVerifierState)
}

func TestSignUp(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	verifierJSON := `
	{
		"method": "spake2plus",
		"salt": "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
		"w0": "H9EeC9z9ndtqPVIz59/hWUUh8/TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ/3QJShr3sjEVFbs/L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9",
		"l": "89obToiiylZJ2bWw9neAUtD+Xvu/zhhj+HHzQveMHMUNhFZh719/tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu+dBqAh2OMVT/ifCPAu2qLqdFJQZRAzM"
	}
	`
	state, err := tc.SignUp(ctx, "app", "https://example.com/", "testsignup@example.com", "", verifierJSON, "", "en")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "SUCCESS", state.Status)
	assert.Equal(t, "https://example.com/", state.RedirectURI)
	assert.NotEmpty(t, state.AuthorizationCode)

	// Check authorization code
	code, err := tc.store.GetAuthorizationCode(ctx, state.AuthorizationCode)
	assert.NoError(t, err)
	assert.NotEmpty(t, code.UserID)

	u, err := tc.userStore.UserByID(ctx, code.UserID)
	assert.NoError(t, err)
	assert.Equal(t, "testsignup@example.com", u.Email.String)
	assert.NotEmpty(t, u.PasswordSalt())
	assert.True(t, u.IsPasswordAuthenticationEnabled())

	// Test missing fields
	_, err = tc.SignUp(ctx, "app", "https://example.com/", "", "", verifierJSON, "", "")
	assert.Error(t, err)

	// Test create account disabled
	viper.Set("sign_up_enabled", false)
	_, err = tc.SignUp(ctx, "app", "https://example.com/", "testsignup@example.com", "", verifierJSON, "", "en")
	assert.Error(t, err)
}

func TestVerifyIDPSuccess(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartIDP(ctx, "app", "mock", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "IDP", state.Status)
	assert.Equal(t, int64(0), state.UserID)
	assert.Equal(t, "https://example.com/", state.RedirectURI)
	assert.Equal(t, "", state.PKCEChallenge)
	assert.Equal(t, "", state.PKCEChallengeMethod)
	assert.Equal(t, "", state.AuthorizationCode)
	assert.Equal(t, "mock", state.IDP)
	assert.Equal(t, idp.State("mock_state"), state.IDPState)
	assert.Equal(t, "https://example.com/authorize", state.IDPAuthorizationURL)
	assert.Empty(t, state.Factors)
	state.IDPAuthorizationURL = ""

	// Check state
	state2, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, state, state2)

	// Step 2 - fail
	state3, err := tc.VerifyIDP(ctx, state.StateToken, "fail")
	assert.Error(t, err)
	assert.Equal(t, "IDP", state3.Status)
	assert.Equal(t, "mock", state3.IDP)

	// Step 2 - success
	state4, err := tc.VerifyIDP(ctx, state.StateToken, "oliver@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", state4.Status)
	assert.NotEmpty(t, state4.AuthorizationCode)

	// Check authorization code
	code, err := tc.store.GetAuthorizationCode(ctx, state4.AuthorizationCode)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), code.UserID)

	// Step 3
	sess, err := tc.ExchangeSession(ctx, "app", "https://example.com/", state4.AuthorizationCode, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(11), sess.UserID)
	assert.Equal(t, "app", sess.ClientID.String)
	assert.False(t, sess.LastPasswordVerifiedAt.Valid)
}

func TestVerifyIDPNewUser(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartIDP(ctx, "app", "mock", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "IDP", state.Status)

	// Step 2 - register
	state3, err := tc.VerifyIDP(ctx, state.StateToken, "newuser@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", state3.Status)
	assert.NotEmpty(t, state3.AuthorizationCode)

	// Check authorization code
	code, err := tc.store.GetAuthorizationCode(ctx, state3.AuthorizationCode)
	assert.NoError(t, err)
	assert.NotEmpty(t, code.UserID)

	// Check registered user
	oauthService := user.OAuthService(999)
	oauthFactor, err := tc.userStore.FindOAuthFactorByOAuthIdentity(ctx, oauthService, "newuser@example.com")
	assert.NoError(t, err)
	assert.Equal(t, code.UserID, oauthFactor.UserID)

	u, err := tc.userStore.UserByID(ctx, code.UserID)
	assert.NoError(t, err)
	assert.Equal(t, "newuser@example.com", u.Email.String)
	assert.False(t, u.IsPasswordAuthenticationEnabled())
}

func TestVerifyIDPNewUserWhenNotAllowed(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	viper.Set("sign_up_enabled", false)
	// Step 1
	state, err := tc.StartIDP(ctx, "app", "mock", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "IDP", state.Status)

	// Step 2 - register
	_, err = tc.VerifyIDP(ctx, state.StateToken, "newuser@example.com")
	assert.Error(t, err)
}

func TestVerifyIDPAlreadyExists(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartIDP(ctx, "app", "mock", "https://example.com/", "", "", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "IDP", state.Status)

	// Step 2 - register
	state3, err := tc.VerifyIDP(ctx, state.StateToken, "carol@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "IDP_ALREADY_EXISTS", state3.Status)
	assert.Equal(t, "mock", state3.IDP)
	assert.Equal(t, []string{"spake2plus"}, state3.Factors)
}

func TestIDPBinding(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartIDPBinding(ctx, 1, "app", "mock", "https://example.com/")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "IDP_BINDING", state.Status)
	assert.Equal(t, int64(1), state.UserID)
	assert.Equal(t, "app", state.ClientID)
	assert.Equal(t, "mock", state.IDP)
	assert.Equal(t, idp.State("mock_state"), state.IDPState)
	assert.Equal(t, "https://example.com/authorize", state.IDPAuthorizationURL)
	assert.Empty(t, state.Factors)
	state.IDPAuthorizationURL = ""

	// Check state
	state2, err := tc.store.GetState(ctx, state.StateToken)
	assert.NoError(t, err)
	assert.Equal(t, state, state2)

	// Step 2 - fail
	state3, err := tc.VerifyIDPBinding(ctx, state.StateToken, 1, "app", "fail")
	assert.Error(t, err)
	assert.Equal(t, "IDP_BINDING", state3.Status)
	assert.Equal(t, "mock", state3.IDP)

	// Step 2 - different user
	state4, err := tc.VerifyIDPBinding(ctx, state.StateToken, 2, "app", "fail")
	assert.Error(t, err)
	assert.Equal(t, "IDP_BINDING", state4.Status)
	assert.Equal(t, "mock", state4.IDP)

	// Step 2 - success
	state5, err := tc.VerifyIDPBinding(ctx, state.StateToken, 1, "app", "bob@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "IDP_BINDING_SUCCESS", state5.Status)
	assert.Empty(t, state5.AuthorizationCode)

	oauthService := user.OAuthService(999)
	oauthFactor, err := tc.userStore.FindOAuthFactorByOAuthIdentity(ctx, oauthService, "bob@example.com")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), oauthFactor.UserID)
	assert.Equal(t, oauthService, oauthFactor.Service)
	assert.Equal(t, "bob@example.com", oauthFactor.OAuthUserID)
}

func TestPasswordStepUp(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	salt, err := base64.StdEncoding.DecodeString("/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=")
	if !assert.NoError(t, err) {
		return
	}

	// Step 1
	state, err := tc.StartStepUp(ctx, 6)
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "STEP_UP", state.Status)
	assert.Equal(t, int64(2), state.UserID)
	assert.Empty(t, state.PasswordVerifierState)
	assert.Equal(t, []string{"password"}, state.Factors)
	assert.Equal(t, "spake2plus", state.PasswordMethod)
	assert.Equal(t, salt, state.PasswordSalt)
	assert.False(t, state.PasswordVerified)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), salt, nil)
	assert.NoError(t, err)
	challenge, err := tc.RequestPasswordStepUp(ctx, state.StateToken, 6, message)
	assert.NoError(t, err)
	assert.NotEmpty(t, challenge)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := sk.GetConfirmation()

	// Step 3
	state2, err := tc.VerifyPasswordStepUp(ctx, state.StateToken, 6, confirmation)
	assert.NoError(t, err)
	assert.Equal(t, "STEP_UP_SUCCESS", state2.Status)
	assert.Empty(t, state2.AuthorizationCode)
	assert.True(t, state2.PasswordVerified)

	sess, err := tc.sessionStore.FindSessionByInternalID(ctx, 6)
	assert.NoError(t, err)
	assert.True(t, sess.LastPasswordVerifiedAt.Valid)
}

func TestPasswordReset(t *testing.T) {
	tc, teardown := tcForTest()
	defer teardown()
	ctx := context.Background()

	// Step 1
	state, err := tc.StartPasswordReset(ctx, "app", "carol@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, state.StateToken)
	assert.Equal(t, "PASSWORD_RESET", state.Status)
	assert.Equal(t, int64(2), state.UserID)
	assert.NotEmpty(t, state.ResetLinkState)

	resetLinkState := make(map[string]interface{})
	err = json.Unmarshal([]byte(state.ResetLinkState), &resetLinkState)
	assert.NoError(t, err)
	token := resetLinkState["token"].(string)

	// Step 2 - wrong Token
	_, err = tc.VerifyPasswordReset(ctx, state.StateToken, "wrong", "")
	assert.Error(t, err)

	// Step 2 - correct token
	state2, err := tc.VerifyPasswordReset(ctx, state.StateToken, token, "")
	assert.NoError(t, err)
	assert.Equal(t, "PASSWORD_RESET", state2.Status)

	// Step 3
	verifierJSON := `
	{
		"method": "spake2plus",
		"salt": "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
		"w0": "H9EeC9z9ndtqPVIz59/hWUUh8/TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ/3QJShr3sjEVFbs/L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9",
		"l": "89obToiiylZJ2bWw9neAUtD+Xvu/zhhj+HHzQveMHMUNhFZh719/tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu+dBqAh2OMVT/ifCPAu2qLqdFJQZRAzM"
	}
	`
	// correct token
	state3, err := tc.VerifyPasswordReset(ctx, state.StateToken, token, verifierJSON)
	assert.NoError(t, err)
	assert.Equal(t, "PASSWORD_RESET_SUCCESS", state3.Status)
}

type mockIDP struct{}

func (p *mockIDP) ID() string {
	return "mock"
}

func (p *mockIDP) AuthorizationURL(stateToken string) (string, idp.State, error) {
	return "https://example.com/authorize", "mock_state", nil
}

func (p *mockIDP) Exchange(ctx context.Context, state idp.State, code string) (*idp.AuthorizationGrant, error) {
	if code == "fail" {
		return nil, errors.New(errors.ErrorUnknown, "idp error")
	}
	return &idp.AuthorizationGrant{
		AccessToken:  "mock_access_token",
		TokenType:    "bearer",
		RefreshToken: "mock_refresh_token",
		Identity: &idp.Identity{
			ID:                code,
			Name:              "mock_name",
			PreferredUsername: "mock_username",
			Email:             code,
			EmailVerified:     true,
		},
	}, nil
}
