package user

import (
	"context"
	"net/http"
	"testing"
	"time"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/internal/validator"
	"authcore.io/authcore/pkg/cryptoutil"

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

func TestAPIListUsers(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(13), res["total_size"])
	assert.Len(t, res["results"], 13)
	assert.Equal(t, float64(1), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Empty(t, res["next_page_token"])
	assert.Empty(t, res["prev_page_token"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users?sort_by=created_at%20desc", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(13), res["total_size"])
	assert.Len(t, res["results"], 13)
	assert.Equal(t, float64(99), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Empty(t, res["next_page_token"])
	assert.Empty(t, res["prev_page_token"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users?email=benny@example.com", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(1), res["total_size"])
	assert.Len(t, res["results"], 1)
	assert.Equal(t, float64(6), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Empty(t, res["next_page_token"])
	assert.Empty(t, res["prev_page_token"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users?email=benny@example.com", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(1), res["total_size"])
	assert.Len(t, res["results"], 1)
	assert.Equal(t, float64(6), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Empty(t, res["next_page_token"])
	assert.Empty(t, res["prev_page_token"])
}

func TestAPIGetUser(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/1", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(1), res["id"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/999", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIListUsersPaginated(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(13), res["total_size"])
	assert.Len(t, res["results"], 13)
	assert.Equal(t, float64(1), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Empty(t, res["next_page_token"])
	assert.Empty(t, res["prev_page_token"])
	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users?limit=10", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(13), res["total_size"])
	assert.Len(t, res["results"], 10)
	assert.Equal(t, float64(1), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.NotEmpty(t, res["next_page_token"])
	assert.Empty(t, res["prev_page_token"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users?limit=10&page_token="+res["next_page_token"].(string), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(13), res["total_size"])
	assert.Len(t, res["results"], 3)
	assert.Equal(t, float64(11), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Empty(t, res["next_page_token"])
	assert.NotEmpty(t, res["prev_page_token"])
}

func TestAPIDeleteUser(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	me := &User{ID: 1}
	code, _, err := testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/2", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	// not exist
	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/999", nil, me)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)

	// delete self
	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/1", nil, me)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestAPIUpdateUser(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodPut, "/api/v2/users/1", map[string]interface{}{"email": "alice@example.com"})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "bob", res["preferred_username"])
	assert.Equal(t, "alice@example.com", res["email"])

	code, res, err = testutil.JSONRequest(e, http.MethodPut, "/api/v2/users/2", map[string]interface{}{"phone_number_verified": true})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "carol", res["preferred_username"])
	assert.Equal(t, "+85221111111", res["phone_number"])
	assert.Equal(t, true, res["phone_number_verified"])

	code, res, err = testutil.JSONRequest(e, http.MethodPut, "/api/v2/users/1", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "bob", res["preferred_username"])
	assert.Equal(t, "alice@example.com", res["email"])

	code, res, err = testutil.JSONRequest(e, http.MethodPut, "/api/v2/users/1", map[string]interface{}{"app_metadata": map[string]interface{}{"test": 1}})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, map[string]interface{}{"test": float64(1)}, res["app_metadata"])

	code, res, err = testutil.JSONRequest(e, http.MethodPut, "/api/v2/users/1", map[string]interface{}{"user_metadata": map[string]interface{}{"test": "string"}})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, map[string]interface{}{"test": "string"}, res["user_metadata"])

	code, res, err = testutil.JSONRequest(e, http.MethodPut, "/api/v2/users/1", map[string]interface{}{"is_locked": true})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, true, res["locked"])

	code, res, err = testutil.JSONRequest(e, http.MethodPut, "/api/v2/users/1", map[string]interface{}{"is_locked": false})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, false, res["locked"])
}

func TestAPIUpdateUserPassword(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	payload := map[string]interface{}{
		"salt": "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
		"w0":   "H9EeC9z9ndtqPVIz59/hWUUh8/TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ/3QJShr3sjEVFbs/L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9",
		"l":    "89obToiiylZJ2bWw9neAUtD+Xvu/zhhj+HHzQveMHMUNhFZh719/tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu+dBqAh2OMVT/ifCPAu2qLqdFJQZRAzM",
	}

	code, _, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/users/1/password", payload)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	code, _, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/users/999/password", payload)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIGetUserRoles(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/1/roles", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Nil(t, res["total_size"])
	assert.Equal(t, "authcore.admin", res["results"].([]interface{})[0].(map[string]interface{})["name"])
	assert.Equal(t, "authcore.editor", res["results"].([]interface{})[1].(map[string]interface{})["name"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/3/roles", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 0)
}

func TestAPIUpdateUserRoles(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/users/3/roles", map[string]interface{}{"role_id": 2})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Nil(t, res["total_size"])
	assert.Equal(t, "authcore.editor", res["results"].([]interface{})[0].(map[string]interface{})["name"])

	code, res, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/users/3/roles", map[string]interface{}{"role_id": 99})
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIListUserIDP(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/10/idp", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Nil(t, res["total_size"])
	assert.Equal(t, float64(2), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, float64(3), res["results"].([]interface{})[1].(map[string]interface{})["id"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/1/idp", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 0)
}

func TestAPIListCurrentUserIDP(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	me := &User{ID: 10}

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current/idp", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "facebook", res["results"].([]interface{})[0].(map[string]interface{})["service"])
	assert.Equal(t, "google", res["results"].([]interface{})[1].(map[string]interface{})["service"])

	me = &User{ID: 1}

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current/idp", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 0)
}

func TestAPIDeleteUserIDP(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, _, err := testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/10/idp/facebook", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/10/idp", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(3), res["results"].([]interface{})[0].(map[string]interface{})["id"])

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/10/idp/facebook", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIDeleteCurrentUserIDP(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	me := &User{ID: 10}
	code, _, err := testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/current/idp/facebook", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current/idp", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "google", res["results"].([]interface{})[0].(map[string]interface{})["service"])

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/current/idp/facebook", nil, me)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIGetIDP(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/idp/2", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(2), res["id"])
	assert.Equal(t, "facebook", res["service"])

	code, _, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/idp/99", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIListUserMFA(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/3/mfa", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 2)
	assert.Equal(t, float64(2), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, "totp", res["results"].([]interface{})[0].(map[string]interface{})["type"])
	assert.Equal(t, "2019-01-09T00:00:00Z", res["results"].([]interface{})[0].(map[string]interface{})["last_used_at"])
	assert.Equal(t, "", res["results"].([]interface{})[0].(map[string]interface{})["value"])
	assert.Equal(t, float64(6), res["results"].([]interface{})[1].(map[string]interface{})["id"])
	assert.Equal(t, "backup_code", res["results"].([]interface{})[1].(map[string]interface{})["type"])
	assert.Equal(t, "2019-07-07T07:07:07Z", res["results"].([]interface{})[1].(map[string]interface{})["last_used_at"])
	assert.Equal(t, "", res["results"].([]interface{})[1].(map[string]interface{})["value"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/5/mfa", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 1)
	assert.Equal(t, float64(1), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, "sms_otp", res["results"].([]interface{})[0].(map[string]interface{})["type"])
	assert.Equal(t, "2019-06-13T00:00:00Z", res["results"].([]interface{})[0].(map[string]interface{})["last_used_at"])
	assert.Equal(t, "+85298765432", res["results"].([]interface{})[0].(map[string]interface{})["value"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/1/mfa", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 0)
}

func TestAPIDeleteMFA(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/3/mfa", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 2)
	assert.Equal(t, float64(2), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, "totp", res["results"].([]interface{})[0].(map[string]interface{})["type"])
	assert.Equal(t, "2019-01-09T00:00:00Z", res["results"].([]interface{})[0].(map[string]interface{})["last_used_at"])
	assert.Equal(t, "", res["results"].([]interface{})[0].(map[string]interface{})["value"])
	assert.Equal(t, float64(6), res["results"].([]interface{})[1].(map[string]interface{})["id"])
	assert.Equal(t, "backup_code", res["results"].([]interface{})[1].(map[string]interface{})["type"])
	assert.Equal(t, "2019-07-07T07:07:07Z", res["results"].([]interface{})[1].(map[string]interface{})["last_used_at"])
	assert.Equal(t, "", res["results"].([]interface{})[1].(map[string]interface{})["value"])

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/mfa/2", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/3/mfa", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 1)
	assert.Equal(t, float64(6), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, "backup_code", res["results"].([]interface{})[0].(map[string]interface{})["type"])
	assert.Equal(t, "2019-07-07T07:07:07Z", res["results"].([]interface{})[0].(map[string]interface{})["last_used_at"])
	assert.Equal(t, "", res["results"].([]interface{})[0].(map[string]interface{})["value"])

	// not exist
	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/mfa/99", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIListCurrentUserMFA(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	me := &User{ID: 3}

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current/mfa", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 2)
	assert.Equal(t, float64(2), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, "totp", res["results"].([]interface{})[0].(map[string]interface{})["type"])
	assert.Equal(t, "2019-01-09T00:00:00Z", res["results"].([]interface{})[0].(map[string]interface{})["last_used_at"])
	assert.Equal(t, "", res["results"].([]interface{})[0].(map[string]interface{})["value"])
	assert.Equal(t, float64(6), res["results"].([]interface{})[1].(map[string]interface{})["id"])
	assert.Equal(t, "backup_code", res["results"].([]interface{})[1].(map[string]interface{})["type"])
	assert.Equal(t, "2019-07-07T07:07:07Z", res["results"].([]interface{})[1].(map[string]interface{})["last_used_at"])
	assert.Equal(t, "", res["results"].([]interface{})[1].(map[string]interface{})["value"])

	me = &User{ID: 5}

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current/mfa", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 1)
	assert.Equal(t, float64(1), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, "sms_otp", res["results"].([]interface{})[0].(map[string]interface{})["type"])
	assert.Equal(t, "2019-06-13T00:00:00Z", res["results"].([]interface{})[0].(map[string]interface{})["last_used_at"])
	assert.Equal(t, "+85298765432", res["results"].([]interface{})[0].(map[string]interface{})["value"])

	me = &User{ID: 1}

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current/mfa", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 0)
}

func TestAPICreateCurrentUserMFA(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	me := &User{ID: 1}
	req := map[string]interface{}{
		"type":   "totp",
		"secret": "THISISATOTPSECRETXXXXXXXXXXXXXXX",
	}

	// Invalid verifier
	code, _, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/users/current/mfa", req, me)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	// Valid verifier
	totpCode := cryptoutil.GetTOTPPin("THISISATOTPSECRETXXXXXXXXXXXXXXX", time.Now())
	req = map[string]interface{}{
		"type":     "totp",
		"secret":   "THISISATOTPSECRETXXXXXXXXXXXXXXX",
		"verifier": []byte(totpCode),
	}
	code, _, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/users/current/mfa", req, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current/mfa", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 1)
	assert.Equal(t, "totp", res["results"].([]interface{})[0].(map[string]interface{})["type"])
}

func TestAPIDeleteCurrentUserMFA(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	me := &User{ID: 3}

	code, _, err := testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/current/mfa/2", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	// Another user's MFA
	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/current/mfa/1", nil, me)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIDeleteRole(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, _, err := testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/1/roles/2", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/1/roles", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, res["results"], 1)

	// nonexisting roles
	code, res, err = testutil.JSONRequest(e, http.MethodPost, "/api/v2/users/3/roles", map[string]interface{}{"role_id": 99})
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIGetCurrentUser(t *testing.T) {
	store, teardown := storeForTest()
	e := echo.New()
	e.Validator = validator.Validator
	APIv2(store)(e)
	defer teardown()

	ctx := context.Background()
	me, err := store.UserByID(ctx, 1)
	assert.NoError(t, err)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(1), res["id"])
	assert.Equal(t, "Bob", res["name"])
	assert.Equal(t, "bob", res["preferred_username"])
	assert.Equal(t, "bob@example.com", res["email"])
	assert.True(t, res["email_verified"].(bool))
	assert.Equal(t, "+85223456789", res["phone_number"])
	assert.True(t, res["phone_number_verified"].(bool))
	assert.Equal(t, map[string]interface{}{"favourite_links": []interface{}{"https://github.com", "https://blocksq.com"}}, res["user_metadata"])
	assert.Equal(t, "authcore.admin", (res["roles"].([]interface{})[0].(map[string]interface{})["name"]))
	assert.Equal(t, "authcore.editor", (res["roles"].([]interface{})[1].(map[string]interface{})["name"]))
	assert.Equal(t, "zh-HK", res["language"])

	// Check fallback case for language field
	me, err = store.UserByID(ctx, 2)
	assert.NoError(t, err)

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current", nil, me)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "en", res["language"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current", nil)
	assert.Error(t, err)
	assert.True(t, errors.IsKind(err, errors.ErrorUnauthenticated))
}

func TestAPIUpdateCurrentUserPassword(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	payload := map[string]interface{}{
		"salt": "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
		"w0":   "H9EeC9z9ndtqPVIz59/hWUUh8/TFdowJApvxHkbRhTZeTsrue0cxUgqUkZ/3QJShr3sjEVFbs/L5Ca3LFIHbPlpWULzMUxmbZSVDQkLSQMdxxxNP1CH9",
		"l":    "89obToiiylZJ2bWw9neAUtD+Xvu/zhhj+HHzQveMHMUNhFZh719/tYgBRvp2LRflO6Rko9q7bUCCRgz4mSBYSibpmCo9y8GoFvWBarSUu+dBqAh2OMVT/ifCPAu2qLqdFJQZRAzM",
	}

	// Success
	me := &User{ID: 2}
	sess := &mockSession{updateCurrentUserPasswordAllowed: true}
	code, res, err := testutil.JSONRequest(e, http.MethodPut, "/api/v2/users/current/password", payload, me, sess)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, res["is_password_set"].(bool))

	// Require step-up authentication
	sess = &mockSession{updateCurrentUserPasswordAllowed: false}
	code, _, err = testutil.JSONRequest(e, http.MethodPut, "/api/v2/users/current/password", payload, me, sess)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, code)

	// Set new password
	me = &User{ID: 1}
	sess = &mockSession{updateCurrentUserPasswordAllowed: false}
	code, res, err = testutil.JSONRequest(e, http.MethodPut, "/api/v2/users/current/password", payload, me, sess)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, res["is_password_set"].(bool))
}

type mockSession struct {
	updateCurrentUserPasswordAllowed bool
}

func (s *mockSession) UpdateCurrentUserPasswordAllowed() bool {
	return s.updateCurrentUserPasswordAllowed
}
