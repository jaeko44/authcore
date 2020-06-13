package rbac

import (
	"context"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
)

// UserRoleResolver is the struct for resolving roles.
// Includes RoleDB for resolving roles.
type UserRoleResolver struct {
	userStore *user.Store
}

// NewUserRoleResolver returns a new UserRoleResolver instance.
func NewUserRoleResolver(userStore *user.Store) *UserRoleResolver {
	return &UserRoleResolver{
		userStore: userStore,
	}
}

// ResolveRole resolves role from context.
func (r *UserRoleResolver) ResolveRole(ctx context.Context) ([]Role, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}
	roles, err := r.userStore.FindAllRolesByUserID(ctx, currentUser.ID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	rbacRoles := []Role{}
	for _, role := range *roles {
		rbacRoles = append(rbacRoles, role.Name)
	}
	return rbacRoles, nil
}
