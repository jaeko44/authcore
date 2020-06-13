package settings

import (
	"net/http"
	"testing"

	"authcore.io/authcore/internal/testutil"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func echoForTest() *echo.Echo {
	e := echo.New()
	APIv2()(e)
	return e
}

func TestAPIPreferences(t *testing.T) {
	e := echoForTest()

	viper.SetDefault("applications.test.name", "Test")

	code, res, error := testutil.JSONRequest(e, http.MethodGet, "/api/v2/preferences?clientId=test", map[string]interface{}{})
	assert.NoError(t, error)
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, res, "app_hosts")
	assert.Contains(t, res, "analytics_token")
	assert.Contains(t, res, "redirect_fallback_url")
	assert.Contains(t, res, "sign_up_enabled")

	preferences := res["preferences"].(map[string]interface{})
	assert.Contains(t, preferences, "company")
	assert.Contains(t, preferences, "logo")
	assert.Contains(t, preferences, "idp_list")

	code, _, error = testutil.JSONRequest(e, http.MethodGet, "/api/v2/preferences", nil)
	assert.Error(t, error)
	assert.Equal(t, http.StatusBadRequest, code)
}
