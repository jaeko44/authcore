package rbac

import (
	"context"

	"authcore.io/authcore/internal/errors"
)

// RoleResolver provides a hook to retrieve an array of roles.
// type RoleResolver func(ctx context.Context) ([]Role, error)
type RoleResolver interface {
	ResolveRole(ctx context.Context) ([]Role, error)
}

// Service represents a RBAC service.
type Service struct {
	roleResolver             RoleResolver
	permissionAssignmentList PermissionAssignmentList
}

// Role represents a role.
type Role interface{}

// Permission represents a permission.
type Permission string

// PermissionAssignment represents a permission assignment (role-permission relation).
type PermissionAssignment struct {
	Role        string
	Permissions []Permission
}

// PermissionAssignmentList represents an array for PermissionAssignment.
type PermissionAssignmentList []PermissionAssignment

// NewService returns a new RBAC service.
func NewService(roleResolver RoleResolver, permissionAssignmentList PermissionAssignmentList) *Service {
	return &Service{
		roleResolver:             roleResolver,
		permissionAssignmentList: permissionAssignmentList,
	}
}

// Authorize returns no errors if an operation could be executed given permissions.
func (service *Service) Authorize(ctx context.Context, permissions ...Permission) error {
	roles, err := service.roleResolver.ResolveRole(ctx)
	if err != nil {
		return err
	}
	return service.authorizeByRolesAndPermissions(roles, permissions)
}

func (service *Service) authorizeByRolesAndPermissions(roles []Role, permissions []Permission) error {
	// TODO: optimise it.  https://gitlab.com/blocksq/authcore/issues/92
	permissionsMatched := make([]bool, len(permissions))
	for roleIndex := 0; roleIndex < len(roles); roleIndex++ {
		role := roles[roleIndex]
		rolePermissions, err := service.getPermissionsByRole(role)
		if err != nil {
			continue
		}
		for rolePermissionsIndex := 0; rolePermissionsIndex < len(rolePermissions); rolePermissionsIndex++ {
			rolePermission := rolePermissions[rolePermissionsIndex]
			for permissionsIndex := 0; permissionsIndex < len(permissions); permissionsIndex++ {
				permission := permissions[permissionsIndex]
				if rolePermission == permission {
					permissionsMatched[permissionsIndex] = true
				}
			}
		}
	}
	for permissionsIndex := 0; permissionsIndex < len(permissions); permissionsIndex++ {
		if !permissionsMatched[permissionsIndex] {
			return errors.New(errors.ErrorUnknown, "not authorized")
		}
	}
	return nil
}

func (service *Service) getPermissionsByRole(role Role) ([]Permission, error) {
	for index := 0; index < len(service.permissionAssignmentList); index++ {
		permissionAssignment := service.permissionAssignmentList[index]
		if permissionAssignment.Role == role {
			return permissionAssignment.Permissions, nil
		}
	}
	return nil, errors.New(errors.ErrorUnknown, "no roles matched")
}
