package authn

import (
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"authcore.io/authcore/internal/audit"
	"authcore.io/authcore/internal/authn/verifier"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/internal/validator"
	"authcore.io/authcore/pkg/cryptoutil"
	"authcore.io/authcore/pkg/nulls"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func echoForTest() (*echo.Echo, func()) {
	tc, teardown := tcForTest()
	e := echo.New()
	e.Validator = validator.Validator
	APIv2(tc, audit.NewLoggingAuditor())(e)
	return e, teardown
}

func TestAPIAuthnPrimary(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	// Step 1
	req := map[string]interface{}{
		"client_id":    "app",
		"handle":       "factor@example.com",
		"redirect_uri": "https://example.com/",
	}
	code, res, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "PRIMARY", res["status"])
	stateToken := res["state_token"]
	assert.NotEmpty(t, stateToken)
	passwordSalt, err := base64.StdEncoding.DecodeString(res["password_salt"].(string))
	assert.NoError(t, err)
	assert.NotEmpty(t, passwordSalt)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), passwordSalt, nil)
	assert.NoError(t, err)

	req = map[string]interface{}{
		"state_token": stateToken,
		"message":     message,
	}
	code, res, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/password", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.NotEmpty(t, res["challenge"])
	challenge, err := base64.StdEncoding.DecodeString(res["challenge"].(string))
	assert.NoError(t, err)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := base64.StdEncoding.EncodeToString(sk.GetConfirmation())

	// Step 3
	req = map[string]interface{}{
		"state_token": stateToken,
		"verifier":    confirmation,
	}
	code, res, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/password/verify", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "MFA_REQUIRED", res["status"])
	assert.Equal(t, []interface{}{"totp", "backup_code"}, res["factors"])

	// Step 4
	totp := cryptoutil.GetTOTPPin("THISISAWEAKTOTPSECRETFORTESTSXX2", time.Now())
	req = map[string]interface{}{
		"state_token": stateToken,
		"verifier":    []byte(totp),
	}
	code, res, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/mfa/totp/verify", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "SUCCESS", res["status"])
	assert.NotEmpty(t, res["authorization_code"])
}

func TestAPIAuthnInvalidRedirectURI(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	req := map[string]interface{}{
		"client_id":    "app",
		"handle":       "factor@example.com",
		"redirect_uri": "https://google.com/",
	}
	_, _, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn", req)
	assert.Error(t, err)
	assert.Equal(t, "redirect_uri https://google.com/ is not allowed", err.Error())
}

func TestAPISignUp(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	// email
	req := map[string]interface{}{
		"client_id":    "app",
		"redirect_uri": "https://example.com/",
		"email":        "testsignup@example.com",
		"password_verifier": map[string]interface{}{
			"method": "spake2plus",
			"salt":   "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
			"w0":     "H9EeC9z9ndtqPVIz59/hWUUh8/TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ/3QJShr3sjEVFbs/L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9",
			"l":      "89obToiiylZJ2bWw9neAUtD+Xvu/zhhj+HHzQveMHMUNhFZh719/tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu+dBqAh2OMVT/ifCPAu2qLqdFJQZRAzM",
		},
		"name":     "test",
		"language": "en",
	}
	code, res, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/signup", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "SUCCESS", res["status"])
	assert.NotEmpty(t, res["authorization_code"])

	// phone
	req = map[string]interface{}{
		"client_id":    "app",
		"redirect_uri": "https://example.com/",
		"phone":        "+85222222222",
		"password_verifier": map[string]interface{}{
			"method": "spake2plus",
			"salt":   "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
			"w0":     "H9EeC9z9ndtqPVIz59/hWUUh8/TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ/3QJShr3sjEVFbs/L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9",
			"l":      "89obToiiylZJ2bWw9neAUtD+Xvu/zhhj+HHzQveMHMUNhFZh719/tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu+dBqAh2OMVT/ifCPAu2qLqdFJQZRAzM",
		},
		"name":     "test",
		"language": "en",
	}
	code, res, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/signup", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "SUCCESS", res["status"])
	assert.NotEmpty(t, res["authorization_code"])

	// email empty
	req = map[string]interface{}{
		"client_id":    "app",
		"redirect_uri": "https://example.com/",
		"email":        "",
		"password_verifier": map[string]interface{}{
			"method": "spake2plus",
			"salt":   "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
			"w0":     "H9EeC9z9ndtqPVIz59/hWUUh8/TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ/3QJShr3sjEVFbs/L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9",
			"l":      "89obToiiylZJ2bWw9neAUtD+Xvu/zhhj+HHzQveMHMUNhFZh719/tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu+dBqAh2OMVT/ifCPAu2qLqdFJQZRAzM",
		},
		"name":     "test",
		"language": "en",
	}
	code, _, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/signup", req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	// phone empty
	req = map[string]interface{}{
		"client_id":    "app",
		"redirect_uri": "https://example.com/",
		"phone":        "",
		"password_verifier": map[string]interface{}{
			"method": "spake2plus",
			"salt":   "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
			"w0":     "H9EeC9z9ndtqPVIz59/hWUUh8/TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ/3QJShr3sjEVFbs/L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9",
			"l":      "89obToiiylZJ2bWw9neAUtD+Xvu/zhhj+HHzQveMHMUNhFZh719/tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu+dBqAh2OMVT/ifCPAu2qLqdFJQZRAzM",
		},
		"name":     "test",
		"language": "en",
	}
	code, _, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/signup", req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	// no email / phone
	req = map[string]interface{}{
		"client_id":    "app",
		"redirect_uri": "https://example.com/",
		"password_verifier": map[string]interface{}{
			"method": "spake2plus",
			"salt":   "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
			"w0":     "H9EeC9z9ndtqPVIz59/hWUUh8/TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ/3QJShr3sjEVFbs/L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9",
			"l":      "89obToiiylZJ2bWw9neAUtD+Xvu/zhhj+HHzQveMHMUNhFZh719/tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu+dBqAh2OMVT/ifCPAu2qLqdFJQZRAzM",
		},
		"name":     "test",
		"language": "en",
	}
	code, _, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/signup", req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	req = map[string]interface{}{
		"client_id":    "app",
		"redirect_uri": "https://example.com/",
		"email":        "",
		"phone":        "",
		"password_verifier": map[string]interface{}{
			"method": "spake2plus",
			"salt":   "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
			"w0":     "H9EeC9z9ndtqPVIz59/hWUUh8/TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ/3QJShr3sjEVFbs/L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9",
			"l":      "89obToiiylZJ2bWw9neAUtD+Xvu/zhhj+HHzQveMHMUNhFZh719/tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu+dBqAh2OMVT/ifCPAu2qLqdFJQZRAzM",
		},
		"name":     "test",
		"language": "en",
	}
	code, _, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/signup", req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	// invalid email
	req = map[string]interface{}{
		"client_id":    "app",
		"redirect_uri": "https://example.com/",
		"email":        "invalid",
		"password_verifier": map[string]interface{}{
			"method": "spake2plus",
			"salt":   "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
			"w0":     "H9EeC9z9ndtqPVIz59/hWUUh8/TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ/3QJShr3sjEVFbs/L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9",
			"l":      "89obToiiylZJ2bWw9neAUtD+Xvu/zhhj+HHzQveMHMUNhFZh719/tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu+dBqAh2OMVT/ifCPAu2qLqdFJQZRAzM",
		},
		"name":     "test",
		"language": "en",
	}
	code, _, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/signup", req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	// test create account blocked
	viper.Set("sign_up_enabled", false)
	req["email"] = "testsignupfail@example.com"
	code, _, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/signup", req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, code)
}

func TestAPIAuthnIDP(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	// Step 1
	req := map[string]interface{}{
		"client_id":    "app",
		"redirect_uri": "https://example.com/",
	}
	code, res, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/idp/mock", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "IDP", res["status"])
	stateToken := res["state_token"]
	assert.NotEmpty(t, stateToken)
	assert.Equal(t, "https://example.com/authorize", res["idp_authorization_url"])
	assert.Equal(t, "https://example.com/", res["redirect_uri"])

	// Step 2 - success
	req = map[string]interface{}{
		"state_token": stateToken,
		"code":        "oliver@example.com",
	}
	code, res, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/idp/mock/verify", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "SUCCESS", res["status"])
	assert.NotEmpty(t, res["authorization_code"])
	assert.Equal(t, "https://example.com/", res["redirect_uri"])
}

func TestAPIPasswordReset(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	// Step 1
	req := map[string]interface{}{
		"handle":    "carol@example.com",
		"client_id": "app",
	}
	code, res, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/password_reset", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "PASSWORD_RESET", res["status"])
	stateToken := res["state_token"]
	assert.NotEmpty(t, stateToken)

	// Step 2 - wrong token
	req = map[string]interface{}{
		"state_token": stateToken,
		"reset_token": "wrong_token",
	}
	code, res, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/password_reset/verify", req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, code)
}

func TestAPIAuthnIDPBinding(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	sess := &session.Session{ID: 1, UserID: 1, ClientID: nulls.NewString("app")}
	me := &user.User{ID: 1}

	// Step 1
	req := map[string]interface{}{
		"redirect_uri": "https://example.com/",
	}
	code, res, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/idp_binding/mock", req, me, sess)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "IDP_BINDING", res["status"])
	stateToken := res["state_token"]
	assert.NotEmpty(t, stateToken)
	assert.Equal(t, "https://example.com/authorize", res["idp_authorization_url"])
	assert.Equal(t, "https://example.com/", res["redirect_uri"])

	// Step 2 - success
	req = map[string]interface{}{
		"state_token": stateToken,
		"code":        "bob@example.com",
	}
	code, res, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/idp_binding/mock/verify", req, me, sess)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "IDP_BINDING_SUCCESS", res["status"])
	assert.Empty(t, res["authorization_code"])
	assert.Equal(t, "https://example.com/", res["redirect_uri"])
}

func TestAPIStepUp(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	sess := &session.Session{ID: 6, UserID: 2, ClientID: nulls.NewString("app")}
	me := &user.User{ID: 2}

	// Step 1
	code, res, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/step_up", nil, me, sess)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "STEP_UP", res["status"])
	stateToken := res["state_token"]
	assert.NotEmpty(t, stateToken)
	passwordSalt, err := base64.StdEncoding.DecodeString(res["password_salt"].(string))
	assert.NoError(t, err)
	assert.NotEmpty(t, passwordSalt)

	// Step 2
	spake2, err := verifier.NewSPAKE2Plus()
	assert.NoError(t, err)
	clientIdentity := []byte("authcoreuser")
	serverIdentity := []byte("authcore")
	cs, message, err := spake2.StartClient(clientIdentity, serverIdentity, []byte("password"), passwordSalt, nil)
	assert.NoError(t, err)

	req := map[string]interface{}{
		"state_token": stateToken,
		"message":     message,
	}
	code, res, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/step_up/password", req, me, sess)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.NotEmpty(t, res["challenge"])
	challenge, err := base64.StdEncoding.DecodeString(res["challenge"].(string))
	assert.NoError(t, err)
	sk, err := cs.Finish(challenge)
	assert.NoError(t, err)
	confirmation := base64.StdEncoding.EncodeToString(sk.GetConfirmation())

	// Step 3
	req = map[string]interface{}{
		"state_token": stateToken,
		"verifier":    confirmation,
	}
	code, res, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/step_up/password/verify", req, me, sess)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "STEP_UP_SUCCESS", res["status"])
}

func TestAPIGetState(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	// Step 1
	req := map[string]interface{}{
		"client_id":    "app",
		"handle":       "factor@example.com",
		"redirect_uri": "https://example.com/",
	}
	code, res1, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)

	// Clear volatile fields
	res1["factors"] = nil
	res1["password_salt"] = nil
	res1["password_method"] = ""

	// Get state
	req = map[string]interface{}{
		"state_token": res1["state_token"],
	}
	code, res2, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/authn/get_state", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, res1, res2)
}
