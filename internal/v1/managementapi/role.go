package managementapi

import (
	"context"
	"strconv"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/rbac"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/managementapi"
)

// ListRoles lists the roles.
func (s *Service) ListRoles(ctx context.Context, in *managementapi.ListRolesRequest) (*managementapi.ListRolesResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, ListRolesPermission)
	if err != nil {
		return nil, err
	}

	roles, err := s.UserStore.FindAllRoles(ctx)
	if err != nil {
		return nil, err
	}

	var pbRoles []*managementapi.Role
	for _, role := range *roles {
		pbRole, err := MarshalRole(&role)
		if err != nil {
			return nil, err
		}
		pbRoles = append(pbRoles, pbRole)
	}

	return &managementapi.ListRolesResponse{
		Roles: pbRoles,
	}, nil
}

// CreateRole creates a non-system role.
func (s *Service) CreateRole(ctx context.Context, in *managementapi.CreateRoleRequest) (*managementapi.Role, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, CreateRolePermission)
	if err != nil {
		return nil, err
	}

	role, err := s.UserStore.CreateRole(ctx, &user.Role{Name: in.Name})
	if err != nil {
		return nil, err
	}

	pbRole, err := MarshalRole(role)
	if err != nil {
		return nil, err
	}

	return pbRole, nil
}

// DeleteRole removes a non-system role.
func (s *Service) DeleteRole(ctx context.Context, in *managementapi.DeleteRoleRequest) (*managementapi.DeleteRoleResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, DeleteRolePermission)
	if err != nil {
		return nil, err
	}

	roleID, err := strconv.ParseInt(in.RoleId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	err = s.UserStore.DeleteRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return &managementapi.DeleteRoleResponse{}, nil
}

// AssignRole assigns a role to an user.
func (s *Service) AssignRole(ctx context.Context, in *managementapi.AssignRoleRequest) (*managementapi.AssignRoleResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, AssignRolePermission)
	if err != nil {
		return nil, err
	}

	roleID, err := strconv.ParseInt(in.RoleId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	userID, err := strconv.ParseInt(in.UserId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	err = s.UserStore.AssignRole(ctx, &user.RoleUser{
		RoleID: roleID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	return &managementapi.AssignRoleResponse{}, nil
}

// UnassignRole unassigns a role to an user.
func (s *Service) UnassignRole(ctx context.Context, in *managementapi.UnassignRoleRequest) (*managementapi.UnassignRoleResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, UnassignRolePermission)
	if err != nil {
		return nil, err
	}

	roleID, err := strconv.ParseInt(in.RoleId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	userID, err := strconv.ParseInt(in.UserId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	err = s.UserStore.UnassignByRoleIDAndUserID(ctx, roleID, userID)
	if err != nil {
		return nil, err
	}

	return &managementapi.UnassignRoleResponse{}, nil
}

// ListRoleAssignments lists role assignment (role-user relation) of an user.
func (s *Service) ListRoleAssignments(ctx context.Context, in *managementapi.ListRoleAssignmentsRequest) (*managementapi.ListRoleAssignmentsResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, ListRoleAssignmentsPermission)
	if err != nil {
		return nil, err
	}

	id, err := strconv.ParseInt(in.UserId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	roles, err := s.UserStore.FindAllRolesByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	var pbRoles []*managementapi.Role
	for _, role := range *roles {
		pbRole, err := MarshalRole(&role)
		if err != nil {
			return nil, err
		}
		pbRoles = append(pbRoles, pbRole)
	}

	return &managementapi.ListRoleAssignmentsResponse{
		Roles: pbRoles,
	}, nil
}

// ListPermissionAssignments lists permission assignment (role-permission relation) of a role
func (s *Service) ListPermissionAssignments(ctx context.Context, in *managementapi.ListPermissionAssignmentsRequest) (*managementapi.ListPermissionAssignmentsResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	err := s.authorize(ctx, ListPermissionAssignmentsPermission)
	if err != nil {
		return nil, err
	}

	roleID, err := strconv.ParseInt(in.RoleId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	role, err := s.UserStore.FindRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	permissions := []rbac.Permission{}
	for _, permissionAssignment := range PermissionAssignments {
		if permissionAssignment.Role == role.Name {
			permissions = permissionAssignment.Permissions
		}
	}

	var pbPermissions []*managementapi.Permission
	for _, permission := range permissions {
		pbPermission, err := MarshalPermission(permission)
		if err != nil {
			return nil, err
		}
		pbPermissions = append(pbPermissions, pbPermission)
	}

	return &managementapi.ListPermissionAssignmentsResponse{
		Permissions: pbPermissions,
	}, nil
}

// ListCurrentUserPermissions lists the permissions for the current user
func (s *Service) ListCurrentUserPermissions(ctx context.Context, in *managementapi.ListCurrentUserPermissionsRequest) (*managementapi.ListCurrentUserPermissionsResponse, error) {
	currentUser, ok := user.CurrentUserFromContext(ctx)
	if !ok || currentUser == nil {
		return nil, errors.New(errors.ErrorUnauthenticated, "")
	}

	roles, err := s.UserStore.FindAllRolesByUserID(ctx, currentUser.ID)
	if err != nil {
		return nil, err
	}

	// TODO: optimise it.  https://gitlab.com/blocksq/kitty/issues/239
	permissions := []rbac.Permission{}

	for _, role := range *roles {
		// find role
		for _, permissionAssignment := range PermissionAssignments {
			if permissionAssignment.Role == role.Name {
				for _, permission := range permissionAssignment.Permissions {
					overlap := false
					for _, foundPermission := range permissions {
						if foundPermission == permission {
							overlap = true
							break
						}
					}
					if !overlap {
						permissions = append(permissions, permission)
					}
				}
				break
			}
		}
	}

	var pbPermissions []*managementapi.Permission
	for _, permission := range permissions {
		pbPermission, err := MarshalPermission(permission)
		if err != nil {
			return nil, err
		}
		pbPermissions = append(pbPermissions, pbPermission)
	}

	return &managementapi.ListCurrentUserPermissionsResponse{
		Permissions: pbPermissions,
	}, nil
}

// MarshalRole marshals a *db.Role into Protobuf message.
func MarshalRole(in *user.Role) (*managementapi.Role, error) {
	return &managementapi.Role{
		Id:         int64(in.ID),
		Name:       in.Name,
		SystemRole: in.IsSystemRole,
	}, nil
}

// MarshalPermission marshals a rbac.Permission into Protobuf message
// TODO: Refactor rbac.Permission into db layer https://gitlab.com/blocksq/kitty/issues/237
func MarshalPermission(in rbac.Permission) (*managementapi.Permission, error) {
	return &managementapi.Permission{
		Name: string(in),
	}, nil
}
