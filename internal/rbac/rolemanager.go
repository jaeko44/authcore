package rbac

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/user"

	"github.com/casbin/casbin/v2/rbac"
	defaultrolemanager "github.com/casbin/casbin/v2/rbac/default-role-manager"
	"github.com/casbin/casbin/v2/util"
)

const (
	userPrefix           = "u:"
	serviceAccountPrefix = "serviceaccount:"
	rolePrefix           = "r:"
)

// RoleManager is a casbin RoleManager implementation. It use the default role manager to stores
// roles inhertance.
type RoleManager struct {
	userStore    *user.Store
	sessionStore *session.Store
	drm          rbac.RoleManager // For role inhertance.
}

// NewRoleManager returns a new RoleManager
func NewRoleManager(userStore *user.Store, sessionStore *session.Store) rbac.RoleManager {
	drm := defaultrolemanager.NewRoleManager(10)
	drm.(*defaultrolemanager.RoleManager).AddMatchingFunc("keyMatch", util.KeyMatch)
	return &RoleManager{
		userStore:    userStore,
		sessionStore: sessionStore,
		drm:          drm,
	}
}

// Clear clears all stored data and resets the role manager to the initial state.
func (rm *RoleManager) Clear() error {
	return rm.drm.Clear()
}

// AddLink adds the inheritance link between role: name1 and role: name2.
// domain is not used.
func (rm *RoleManager) AddLink(name1 string, name2 string, domain ...string) error {
	return rm.drm.AddLink(name1, name2, domain...)
}

// DeleteLink deletes the inheritance link between role: name1 and role: name2.
// domain is not used.
func (rm *RoleManager) DeleteLink(name1 string, name2 string, domain ...string) error {
	return rm.drm.DeleteLink(name1, name2, domain...)
}

// HasLink determines whether role: name1 inherits role: name2.
// domain is not used.
func (rm *RoleManager) HasLink(name1 string, name2 string, domain ...string) (bool, error) {
	if len(domain) >= 1 {
		return false, errors.New(errors.ErrorUnknown, "error: domain should not be used")
	}

	// Short cut
	if name1 == name2 {
		return true, nil
	}

	var roles []string
	if strings.HasPrefix(name1, userPrefix) || strings.HasPrefix(name1, serviceAccountPrefix) {
		var err error
		roles, err = rm.GetRoles(name1)
		if err != nil {
			return false, err
		}

		for _, role := range roles {
			ok, err := rm.drm.HasLink(role, name2)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		// Fall back to defaultrolemanager to check if it is defined in policy.
	}
	return rm.drm.HasLink(name1, name2)
}

// GetRoles gets the roles that a subject inherits.
// domain is not used.
func (rm *RoleManager) GetRoles(name string, domain ...string) (roles []string, err error) {
	if len(domain) >= 1 {
		return nil, errors.New(errors.ErrorUnknown, "error: domain should not be used")
	}

	ctx := context.Background()
	if strings.HasPrefix(name, userPrefix) {
		var id int64
		id, err = strconv.ParseInt(name[2:], 10, 64)
		if err != nil {
			err = errors.Wrap(err, errors.ErrorUnknown, "invalid user id")
			return
		}
		var rs *[]user.Role
		rs, err = rm.userStore.FindAllRolesByUserID(ctx, id)
		if err != nil {
			err = errors.Wrap(err, errors.ErrorUnknown, "error getting roles by user")
			return
		}

		roles = make([]string, len(*rs))
		for i, r := range *rs {
			roles[i] = fmt.Sprintf("r:%v", r.Name)
		}
	} else if strings.HasPrefix(name, serviceAccountPrefix) {
		serviceAccounts, err := rm.sessionStore.AllServiceAccounts()
		if err != nil {
			err = errors.Wrap(err, errors.ErrorUnknown, "error getting service accounts")
		}
		for _, sa := range serviceAccounts {
			if sa.SubjectString() == name {
				roles = make([]string, len(sa.Roles))
				for i, r := range sa.Roles {
					roles[i] = fmt.Sprintf("r:%v", r)
				}
				break
			}
		}
	}

	return
}

// GetUsers gets the users that inherits a subject.
// domain is not used.
func (rm *RoleManager) GetUsers(name string, domain ...string) (users []string, err error) {
	if len(domain) >= 1 {
		return nil, errors.New(errors.ErrorUnknown, "error: domain should not be used")
	}
	ctx := context.Background()

	var role *user.Role
	role, err = rm.userStore.FindRoleByName(ctx, name[2:])
	if err != nil {
		err = errors.Wrap(err, errors.ErrorUnknown, "role not found")
		return
	}
	var us *[]user.User
	us, err = rm.userStore.AllUsersByRoleID(ctx, role.ID)
	if err != nil {
		err = errors.Wrap(err, errors.ErrorUnknown, "error getting users by role")
		return
	}

	users = make([]string, len(*us))
	for i, u := range *us {
		users[i] = fmt.Sprintf("u:%v", u.ID)
	}

	serviceAccounts, err := rm.sessionStore.AllServiceAccounts()
	if err != nil {
		err = errors.Wrap(err, errors.ErrorUnknown, "error getting service accounts")
	}
	for _, sa := range serviceAccounts {
		if sa.HasRole(name[2:]) {
			users = append(users, serviceAccountPrefix+sa.ID)
		}
	}

	return
}

// PrintRoles prints all the roles to log.
func (rm *RoleManager) PrintRoles() error {
	return rm.drm.PrintRoles()
}
