package audit

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

func TestAPIListAuditLogs(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/api/v2/audit_logs?user_id=1&limit=2", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(3), res["total_size"])
	assert.Len(t, res["results"], 2)
	assert.Equal(t, float64(5), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.NotEmpty(t, res["next_page_token"])
	assert.Empty(t, res["prev_page_token"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/audit_logs?user_id=1&limit=10&page_token="+res["next_page_token"].(string), nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(3), res["total_size"])
	assert.Len(t, res["results"], 1)
	assert.Equal(t, float64(1), res["results"].([]interface{})[0].(map[string]interface{})["id"])
	assert.Empty(t, res["next_page_token"])
	assert.NotEmpty(t, res["prev_page_token"])

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/audit_logs", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(7), res["total_size"])
	assert.Len(t, res["results"], 7)

	code, res, err = testutil.JSONRequest(e, http.MethodGet, "/api/v2/audit_logs?user_id=", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, float64(7), res["total_size"])
	assert.Len(t, res["results"], 7)
}
