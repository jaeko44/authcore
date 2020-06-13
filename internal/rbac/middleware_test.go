package rbac

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"authcore.io/authcore/internal/config"
	dbx "authcore.io/authcore/internal/db"
	httpserver "authcore.io/authcore/internal/http"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/internal/user"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	testutil.DBSetUp()
	code := m.Run()
	testutil.DBTearDown()
	os.Exit(code)
}

func enforcerForTest() (*Enforcer, func()) {
	config.InitDefaults()
	viper.Set("base_path", "../..")
	config.InitConfig()
	testutil.FixturesSetUp()
	db := dbx.NewDBFromConfig()
	redis := testutil.RedisForTest()
	encryptor := testutil.EncryptorForTest()
	userStore := user.NewStore(db, redis, encryptor)
	sessionStore := session.NewStore(db, redis, userStore)
	enforcer := NewEnforcer(userStore, sessionStore)

	return enforcer, func() {
		db.Close()
		viper.Reset()
		redis.FlushAll()
	}
}

func TestEnforcerUser(t *testing.T) {
	enforcer, teardown := enforcerForTest()
	defer teardown()
	s := httpserver.NewServer(mockUserMiddleware, EnforcerMiddleware(nil, enforcer))
	e := s.Echo()

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest("POST", "/healthz", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	req = httptest.NewRequest("GET", "/api/v2/users", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEnforcerGuest(t *testing.T) {
	enforcer, teardown := enforcerForTest()
	defer teardown()
	s := httpserver.NewServer(EnforcerMiddleware(nil, enforcer))
	e := s.Echo()

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest("POST", "/healthz", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestEnfocerMissingPolicy(t *testing.T) {
	enforcer, teardown := enforcerForTest()
	defer teardown()
	s := httpserver.NewServer(mockUserMiddleware, EnforcerMiddleware(nil, enforcer))
	e := s.Echo()

	req := httptest.NewRequest("GET", "/undefined_policy", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func mockUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("user", &user.User{ID: 1})
		c.Set("subject", "u:1")
		return next(c)
	}
}
