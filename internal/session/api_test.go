package session

import (
	"context"
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

func TestAPIListUserSessions(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/1/sessions", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(3), res["total_size"])
	assert.Equal(t, float64(3), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, float64(2), res["results"].([]interface{})[1].(map[string]interface{})["id"])
	assert.Equal(t, float64(1), res["results"].([]interface{})[2].(map[string]interface{})["id"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/5/sessions", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(0), res["total_size"])
}

func TestAPIListCurrentUserSessions(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()
	store, teardownStore := storeForTest()
	defer teardownStore()

	ctx := context.Background()
	user, err := store.userStore.UserByID(ctx, 1)
	assert.NoError(t, err)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current/sessions", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, code)

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/current/sessions", nil, user)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)

	assert.Equal(t, float64(1), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, float64(2), res["results"].([]interface{})[1].(map[string]interface{})["id"])
	assert.Equal(t, float64(3), res["results"].([]interface{})[2].(map[string]interface{})["id"])
}

func TestAPIGetSession(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/sessions/1", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(1), res["id"])

	code, _, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/sessions/99", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIGetCurrentSession(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()
	store, teardownStore := storeForTest()
	defer teardownStore()

	ctx := context.Background()
	user, err := store.userStore.UserByID(ctx, 1)
	assert.NoError(t, err)
	session, err := store.FindSessionByInternalID(ctx, 3)
	assert.NoError(t, err)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/sessions/current", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, code)

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/sessions/current", nil, user, session)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "example-client", res["client_id"])
	assert.Equal(t, "1.1.1.1", res["last_seen_ip"])
}

func TestAPIDeleteSession(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, _, err := testutil.JSONRequest(e, http.MethodDelete, "/api/v2/sessions/1", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/1/sessions", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(2), res["total_size"])
	assert.Equal(t, float64(3), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, float64(2), res["results"].([]interface{})[1].(map[string]interface{})["id"])

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/sessions/99", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIDeleteCurrentSession(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()
	store, teardownStore := storeForTest()
	defer teardownStore()

	ctx := context.Background()
	user, err := store.userStore.UserByID(ctx, 1)
	assert.NoError(t, err)
	session, err := store.FindSessionByInternalID(ctx, 3)
	assert.NoError(t, err)

	code, _, err := testutil.JSONRequest(e, http.MethodDelete, "/api/v2/sessions/current", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, code)

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/sessions/current", nil, user, session)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/1/sessions", nil, user)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(2), res["total_size"])
	assert.Equal(t, float64(2), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, float64(1), res["results"].([]interface{})[1].(map[string]interface{})["id"])

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/sessions/3", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAPIDeleteCurrentUserSession(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()
	store, teardownStore := storeForTest()
	defer teardownStore()

	ctx := context.Background()
	user, err := store.userStore.UserByID(ctx, 1)
	assert.NoError(t, err)

	code, _, err := testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/current/sessions/3", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, code)

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/current/sessions/4", nil, user)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/users/current/sessions/3", nil, user)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, code)

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/users/1/sessions", nil, user)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(2), res["total_size"])
	assert.Equal(t, float64(2), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Equal(t, float64(1), res["results"].([]interface{})[1].(map[string]interface{})["id"])

	code, _, err = testutil.JSONRequest(e, http.MethodDelete, "/api/v2/sessions/3", nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}
