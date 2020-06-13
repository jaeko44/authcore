package authentication

import (
	"context"
	"os"
	"testing"

	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/rbac"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/internal/user"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	testutil.MigrationsDir = "../../../db/migrations"
	testutil.FixturesDir = "../../../db/fixtures"
	testutil.DBSetUp()
	code := m.Run()
	testutil.DBTearDown()
	os.Exit(code)
}

func ServiceForTest() (*Service, func()) {
	testutil.FixturesSetUp()
	db := db.NewDBFromConfig()
	redisClient := testutil.RedisForTest()
	config.InitDefaults()
	viper.Set("secret_key_base", "855edf399835e9c9deb61877c1a76bf14eed7c35a167e10ff1b7d43db4363268")
	viper.Set("applications.authcore-io.allowed_callback_urls", []string{
		"https://example.com/",
	})
	config.InitConfig()

	encryptor := testutil.EncryptorForTest()
	userStore := user.NewStore(db, redisClient, encryptor)
	service := NewService(redisClient, userStore)

	return service, func() {
		db.Close()
		viper.Reset()
		redisClient.FlushAll()
	}
}

type roleResolver struct{}

func (r *roleResolver) ResolveRole(ctx context.Context) ([]rbac.Role, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}
	if currentUser.ID == 1 {
		return []rbac.Role{"authcore.admin"}, nil
	} else if currentUser.ID == 2 {
		return []rbac.Role{"authcore.editor"}, nil
	}
	return []rbac.Role{}, nil
}
