package registration

import (
	"context"
	"os"
	"testing"

	"authcore.io/authcore/internal/config"
	dbx "authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/email"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/sms"
	"authcore.io/authcore/internal/template"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/nulls"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestMain(m *testing.M) {
	testutil.MigrationsDir = "../../../db/migrations"
	testutil.FixturesDir = "../../../db/fixtures"
	testutil.DBSetUp()
	code := m.Run()
	testutil.DBTearDown()
	os.Exit(code)
}

type testApp struct {
	db           *dbx.DB
	userStore    *user.Store
	sessionStore *session.Store
	emailService *email.Service
	smsService   *sms.Service
}

func appForTest() (*testApp, func()) {
	config.InitDefaults()
	viper.Set("require_user_email_or_phone", false)
	viper.Set("require_user_phone", false)
	viper.Set("require_user_email", false)
	viper.Set("require_user_username", false)
	viper.Set("secret_key_base", "855edf399835e9c9deb61877c1a76bf14eed7c35a167e10ff1b7d43db4363268")
	viper.Set("base_path", "../..")
	viper.Set("applications.example-client.name", "example")
	config.InitConfig()

	testutil.FixturesSetUp()
	redis := testutil.RedisForTest()
	encryptor := testutil.EncryptorForTest()
	db := dbx.NewDBFromConfig()
	templateStore := template.NewStore(db)
	app := new(testApp)
	app.db = db
	app.userStore = user.NewStore(app.db, redis, encryptor)
	app.sessionStore = session.NewStore(app.db, redis, app.userStore)
	app.emailService = email.NewService(templateStore)
	app.smsService = sms.NewService(templateStore)

	return app, func() {
		db.Close()
		viper.Reset()
		redis.FlushAll()
	}
}

// Create user in the server.
func TestRegisterUser(t *testing.T) {
	app, teardown := appForTest()
	defer teardown()

	u := &user.User{
		Username:       dbx.NullableString("alice"),
		Email:          dbx.NullableString("alice@example.com"),
		Phone:          dbx.NullableString("+85212345678"),
		DisplayNameOld: "Alice",
		Language:       dbx.NullableString("en"),
	}

	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	m["grpcgateway-user-agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	session, err := RegisterUser(ctx, app.userStore, app.sessionStore, app.emailService, app.smsService, u, "example-client", false, true)

	if assert.NoError(t, err) {
		assert.NotNil(t, u)
		assert.NotNil(t, session)
		assert.Len(t, session.RefreshToken, 43)
		assert.NotEqual(t, "", u.ID)
		assert.Equal(t, nulls.NewString("alice"), u.Username)
		assert.Equal(t, nulls.NewString("alice@example.com"), u.Email)
		assert.Equal(t, nulls.NewString("+85212345678"), u.Phone)
		assert.NotZero(t, u.CreatedAt)
		assert.NotZero(t, u.UpdatedAt)
	}
}

func TestRegisterUserWithEmailAsHandle(t *testing.T) {
	app, teardown := appForTest()
	defer teardown()
	viper.Set("require_user_email", true)

	u := &user.User{
		Username:       dbx.NullableString(""),
		Email:          dbx.NullableString("alice@example.com"),
		Phone:          dbx.NullableString(""),
		DisplayNameOld: "Alice",
		Language:       dbx.NullableString("en"),
	}

	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	m["grpcgateway-user-agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	session, err := RegisterUser(ctx, app.userStore, app.sessionStore, app.emailService, app.smsService, u, "example-client", false, true)

	if assert.NoError(t, err) {
		assert.NotNil(t, u)
		assert.NotNil(t, session)
		assert.Len(t, session.RefreshToken, 43)
		assert.NotEqual(t, "", u.ID)
		assert.Equal(t, "", u.Username.String)
		assert.Equal(t, nulls.NewString("alice@example.com"), u.Email)
		assert.Equal(t, "", u.Phone.String)
		assert.NotZero(t, u.CreatedAt)
		assert.NotZero(t, u.UpdatedAt)
	}
}

func TestRegisterUserWithPhoneAsHandle(t *testing.T) {
	app, teardown := appForTest()
	defer teardown()
	viper.Set("require_user_phone", true)

	u := &user.User{
		Username:       dbx.NullableString(""),
		Email:          dbx.NullableString(""),
		Phone:          dbx.NullableString("+85212345678"),
		DisplayNameOld: "Alice",
		Language:       dbx.NullableString("en"),
	}
	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	m["grpcgateway-user-agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	session, err := RegisterUser(ctx, app.userStore, app.sessionStore, app.emailService, app.smsService, u, "example-client", false, true)

	if assert.NoError(t, err) {
		assert.NotNil(t, u)
		assert.NotNil(t, session)
		assert.Len(t, session.RefreshToken, 43)
		assert.NotEqual(t, "", u.ID)
		assert.Equal(t, "", u.Username.String)
		assert.Equal(t, "", u.Email.String)
		assert.Equal(t, nulls.NewString("+85212345678"), u.Phone)
		assert.NotZero(t, u.CreatedAt)
		assert.NotZero(t, u.UpdatedAt)
	}
}

func TestRegisterUserWithUsernameAsHandle(t *testing.T) {
	app, teardown := appForTest()
	defer teardown()
	viper.Set("require_user_username", true)

	u := &user.User{
		Username:       dbx.NullableString("alice"),
		Email:          dbx.NullableString(""),
		Phone:          dbx.NullableString(""),
		DisplayNameOld: "Alice",
		Language:       dbx.NullableString("en"),
	}
	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	m["grpcgateway-user-agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	session, err := RegisterUser(ctx, app.userStore, app.sessionStore, app.emailService, app.smsService, u, "example-client", false, true)

	if assert.NoError(t, err) {
		assert.NotNil(t, u)
		assert.NotNil(t, session)
		assert.Len(t, session.RefreshToken, 43)
		assert.NotEqual(t, "", u.ID)
		assert.Equal(t, nulls.NewString("alice"), u.Username)
		assert.Equal(t, "", u.Email.String)
		assert.Equal(t, "", u.Phone.String)
		assert.NotZero(t, u.CreatedAt)
		assert.NotZero(t, u.UpdatedAt)
	}
}

func TestRegisterUserWithoutPhoneRequirement(t *testing.T) {
	app, teardown := appForTest()
	defer teardown()
	viper.Set("require_user_phone", true)

	u := &user.User{
		Username:       dbx.NullableString("Alice"),
		Email:          dbx.NullableString("alice@example.com"),
		Phone:          dbx.NullableString(""),
		DisplayNameOld: "Alice",
		Language:       dbx.NullableString("en"),
	}
	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	m["grpcgateway-user-agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	_, err := RegisterUser(ctx, app.userStore, app.sessionStore, app.emailService, app.smsService, u, "example-client", false, true)

	assert.Error(t, err)
}

func TestRegisterUserWithoutContactRequirement(t *testing.T) {
	app, teardown := appForTest()
	defer teardown()
	viper.Set("require_user_email_or_phone", true)

	u := &user.User{
		Username:       dbx.NullableString(""),
		Email:          dbx.NullableString(""),
		Phone:          dbx.NullableString(""),
		DisplayNameOld: "Alice",
		Language:       dbx.NullableString("en"),
	}
	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	m["grpcgateway-user-agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	_, err := RegisterUser(ctx, app.userStore, app.sessionStore, app.emailService, app.smsService, u, "example-client", false, true)

	if assert.Error(t, err) {
		ie := err.(*errors.Error)
		assert.NotNil(t, ie.FieldViolations())
		assert.Len(t, ie.FieldViolations(), 2)
		assert.Equal(t, ie.FieldViolations()[0].Field, "phone")
		assert.Equal(t, ie.FieldViolations()[1].Field, "email")
	}
}

func TestRegisterUserWithoutSetting(t *testing.T) {
	app, teardown := appForTest()
	defer teardown()

	u := &user.User{
		Username:       dbx.NullableString(""),
		Email:          dbx.NullableString(""),
		Phone:          dbx.NullableString(""),
		DisplayNameOld: "Alice",
		Language:       dbx.NullableString("en"),
	}
	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	m["grpcgateway-user-agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	_, err := RegisterUser(ctx, app.userStore, app.sessionStore, app.emailService, app.smsService, u, "example-client", false, true)

	if assert.Error(t, err) {
		ie := err.(*errors.Error)
		assert.NotNil(t, ie.FieldViolations())
		assert.Len(t, ie.FieldViolations(), 1)
		assert.Equal(t, ie.FieldViolations()[0].Field, "email")
	}
}

func TestRegisterUserWithoutSession(t *testing.T) {
	app, teardown := appForTest()
	defer teardown()

	u := &user.User{
		Username:       dbx.NullableString("alice"),
		Email:          dbx.NullableString("alice@example.com"),
		Phone:          dbx.NullableString("+85212345678"),
		DisplayNameOld: "Alice",
		Language:       dbx.NullableString("en"),
	}
	ctx := context.Background()
	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	m["grpcgateway-user-agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	session, err := RegisterUser(ctx, app.userStore, app.sessionStore, app.emailService, app.smsService, u, "example-client", false, false)

	if assert.NoError(t, err) {
		assert.NotNil(t, u)
		assert.Nil(t, session)
		assert.NotEqual(t, "", u.ID)
		assert.Equal(t, nulls.NewString("alice"), u.Username)
		assert.Equal(t, nulls.NewString("alice@example.com"), u.Email)
		assert.Equal(t, nulls.NewString("+85212345678"), u.Phone)
		assert.NotZero(t, u.CreatedAt)
		assert.NotZero(t, u.UpdatedAt)
	}
}
