package registration

import (
	"net/http"
	"testing"

	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/email"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/sms"
	"authcore.io/authcore/internal/template"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/internal/validator"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func storeForTest() (*user.Store, *session.Store, *email.Service, *sms.Service, func()) {
	config.InitDefaults()
	viper.Set("secret_key_base", "855edf399835e9c9deb61877c1a76bf14eed7c35a167e10ff1b7d43db4363268")
	viper.Set("base_path", "../..")
	viper.Set("applications.app.name", "app")
	viper.Set("applications.app.allowed_callback_urls", []string{"https://example.com"})
	viper.Set("applications.example-client.name", "example")
	config.InitConfig()

	testutil.FixturesSetUp()
	d := db.NewDBFromConfig()
	redis := testutil.RedisForTest()
	encryptor := testutil.EncryptorForTest()
	userStore := user.NewStore(d, redis, encryptor)
	sessionStore := session.NewStore(d, redis, userStore)
	templateStore := template.NewStore(d)
	emailService := email.NewService(templateStore)
	smsService := sms.NewService(templateStore)

	return userStore, sessionStore, emailService, smsService, func() {
		d.Close()
		viper.Reset()
		redis.FlushAll()
	}
}

func echoForTest() (*echo.Echo, func()) {
	userStore, sesisonStore, emailService, smsService, teardown := storeForTest()
	e := echo.New()
	e.Validator = validator.Validator
	APIv2(userStore, sesisonStore, emailService, smsService)(e)
	return e, teardown
}
func TestAPICreateUser(t *testing.T) {
	e, teardown := echoForTest()
	defer teardown()

	payload := map[string]interface{}{
		"username":     "test",
		"email":        "test@example.com",
		"phone_number": "+85223423423",
	}
	code, res, err := testutil.JSONRequest(e, http.MethodPost, "/api/v2/users", payload)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "test@example.com", res["user"].(map[string]interface{})["email"])
	assert.Equal(t, "+85223423423", res["user"].(map[string]interface{})["phone_number"])
	assert.Equal(t, "test", res["user"].(map[string]interface{})["preferred_username"])
	assert.NotEqual(t, "", res["refresh_token"])
}
