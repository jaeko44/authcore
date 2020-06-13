package template

import (
	"net/http"
	"testing"

	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/internal/validator"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func echoForTest() (*echo.Echo, func()) {
	store, teardown := storeForTest()
	e := echo.New()
	e.Validator = validator.Validator
	APIv2(store)(e)
	return e, teardown
}

func TestAPIListAvaliableLanguage(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Nil(t, res["total_size"])
	assert.Equal(t, "en", res["results"].([]interface{})[0].(string))
	assert.Equal(t, "zh-HK", res["results"].([]interface{})[1].(string))
}

func TestAPIListTemplates(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/email/en", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Nil(t, res["total_size"])
	assert.Equal(t, "VerificationMail", res["results"].([]interface{})[0].(map[string]interface{})["name"])
	assert.Equal(t, "ResetPasswordAuthenticationMail", res["results"].([]interface{})[1].(map[string]interface{})["name"])
	assert.Equal(t, "en", res["results"].([]interface{})[1].(map[string]interface{})["language"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/sms/zh-HK", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Nil(t, res["total_size"])
	assert.Equal(t, "AuthenticationSMS", res["results"].([]interface{})[0].(map[string]interface{})["name"])
	assert.Equal(t, "VerificationSMS", res["results"].([]interface{})[1].(map[string]interface{})["name"])
	assert.Equal(t, "ResetPasswordAuthenticationSMS", res["results"].([]interface{})[2].(map[string]interface{})["name"])
	assert.Equal(t, "zh-HK", res["results"].([]interface{})[1].(map[string]interface{})["language"])

	code, _, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/email/unknown", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	code, _, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/unknown/unknown", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestAPIGetTemplate(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/sms/en/AuthenticationSMS", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "Your authentication code for {application_name} is {code}.", res["text"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/email/en/VerificationMail", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "Please verify your email address", res["subject"])
	assert.NotEmpty(t, res["html"])
	assert.NotEmpty(t, res["text"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/email/en/UnknownEmail", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/email/unknown/VerificationMail", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/unknown/email/VerificationMail", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestAPIUpdateTemplate(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	payload := map[string]interface{}{
		"subject": "a",
		"html":    "b",
		"text":    "c",
	}
	code, _, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/templates/email/en/ResetPasswordAuthenticationMail", payload)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/email/en/ResetPasswordAuthenticationMail", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "a", res["subject"])
	assert.Equal(t, "b", res["html"])
	assert.Equal(t, "c", res["text"])

	code, _, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/templates/email/en/unknown", payload)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	code, _, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/templates/sms/unknown/ResetPasswordAuthenticationSMS", payload)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	code, _, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/templates/unknown/en/ResetPasswordAuthenticationSMS", payload)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestAPIResetTemplate(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	payload := map[string]interface{}{
		"text": "a",
	}
	code, _, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/templates/sms/en/ResetPasswordAuthenticationSMS", payload)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/sms/en/ResetPasswordAuthenticationSMS", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "a", res["text"])

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/templates/sms/en/ResetPasswordAuthenticationSMS", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/templates/sms/en/ResetPasswordAuthenticationSMS", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "Please visit {reset_password_link} to continue your reset password process for {application_name}.", res["text"])

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/templates/sms/en/unknown", payload)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/templates/sms/unknown/ResetPasswordAuthenticationSMS", payload)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/templates/unknown/en/ResetPasswordAuthenticationSMS", payload)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)
}
