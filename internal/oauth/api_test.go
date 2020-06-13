package oauth

import (
	"net/http"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"authcore.io/authcore/internal/authn"
	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/internal/validator"
)

func TestMain(m *testing.M) {
	testutil.DBSetUp()
	code := m.Run()
	testutil.DBTearDown()
	os.Exit(code)
}

func echoForTest() (*echo.Echo, func()) {
	config.InitDefaults()
	viper.Set("secret_key_base", "855edf399835e9c9deb61877c1a76bf14eed7c35a167e10ff1b7d43db4363268")
	viper.Set("access_token_private_key", `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEICLBfuNNrqDL6LDLeaQFaytAGDP7hk65Q4J2c8iBumlqoAoGCCqGSM49
AwEHoUQDQgAEKY6MShC7UrSkekyczKKvZQXuxFKDRd0DEgV6r9XeDAZoYPPTvgx3
oNBTatFJjSOJ/qRrBbqvbZDiPOLpJ7vlaQ==
-----END EC PRIVATE KEY-----
	`)
	config.InitConfig()

	testutil.FixturesSetUp()
	d := db.NewDBFromConfig()
	redis := testutil.RedisForTest()
	encryptor := testutil.EncryptorForTest()
	authnStore := authn.NewStore(redis, encryptor)
	userStore := user.NewStore(d, redis, encryptor)
	sessionStore := session.NewStore(d, redis, userStore)
	tc := authn.NewTransactionController(d, authnStore, userStore, sessionStore)

	e := echo.New()
	e.Validator = validator.Validator
	API(userStore, sessionStore, tc)(e)

	return e, func() {
		d.Close()
		viper.Reset()
		redis.FlushAll()
	}
}

func TestOpenIDConfigurationEndpoint(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	// Step 1
	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/.well-known/openid-configuration", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "https://authcore.localhost/", res["issuer"])
	assert.Equal(t, "https://authcore.localhost/oauth/authorize", res["authorization_endpoint"])
	assert.Equal(t, "https://authcore.localhost/oauth/token", res["token_endpoint"])
	assert.Equal(t, "https://authcore.localhost/.well-known/jwks.json", res["jwks_uri"])
}

func TestTokenEndpoint(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	// Step 1
	req := map[string]interface{}{
		"grant_type":    "refresh_token",
		"refresh_token": "BOBREFRESHTOKEN1",
	}
	code, res, err := testutil.JSONRequest(e, http.MethodPost, "/oauth/token", req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "bearer", res["token_type"])
	assert.NotEmpty(t, res["access_token"])
	assert.NotEmpty(t, res["id_token"])
}

func TestJWKSEndpoint(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	code, res, err := testutil.JSONRequest(e, http.MethodGet, "/.well-known/jwks.json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.NotEmpty(t, res["keys"])
	assert.Len(t, res["keys"], 1)
	assert.Equal(t, "EC", res["keys"].([]interface{})[0].(map[string]interface{})["kty"])
	assert.Equal(t, "P-256", res["keys"].([]interface{})[0].(map[string]interface{})["crv"])
	assert.Equal(t, "sig", res["keys"].([]interface{})[0].(map[string]interface{})["use"])
	assert.Equal(t, "ES256", res["keys"].([]interface{})[0].(map[string]interface{})["alg"])
	assert.Equal(t, "KY6MShC7UrSkekyczKKvZQXuxFKDRd0DEgV6r9XeDAY", res["keys"].([]interface{})[0].(map[string]interface{})["x"])
	assert.Equal(t, "aGDz074Md6DQU2rRSY0jif6kawW6r22Q4jzi6Se75Wk", res["keys"].([]interface{})[0].(map[string]interface{})["y"])
	assert.Equal(t, "hN351AH04N2BBba3N6PgNcVloRohu6KkDRDMcvr5k28", res["keys"].([]interface{})[0].(map[string]interface{})["kid"])
}
