package rbac

import (
	"path/filepath"

	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/internal/session"

	"github.com/casbin/casbin/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Enforcer is responsible for authorization enforcement.
type Enforcer struct {
	*casbin.Enforcer
	userStore *user.Store
}

// NewEnforcer configures a new casbin.Enforcer.
func NewEnforcer(userStore *user.Store, sessionStore *session.Store) *Enforcer {
	basePath := viper.GetString("base_path")
	modelPath := filepath.Join(basePath, "policies/rbac_model.conf")
	policyPath := filepath.Join(basePath, "policies/rbac_policy.csv")
	e, err := casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		log.Fatalf("error initializing enforcer: %v", err)
	}
	e.SetRoleManager(NewRoleManager(userStore, sessionStore))
	e.LoadPolicy()
	return &Enforcer{
		Enforcer:  e,
		userStore: userStore,
	}
}
