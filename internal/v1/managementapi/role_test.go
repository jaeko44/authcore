package managementapi

import (
	"context"
	"testing"

	// "authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/managementapi"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListRoles(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List roles
	req := &managementapi.ListRolesRequest{}
	res, err := srv.ListRoles(ctx, req)

	if assert.Nil(t, err) {
		assert.Len(t, res.Roles, 3)

		role1 := res.Roles[0]
		assert.Equal(t, int64(1), role1.Id)
		assert.Equal(t, "authcore.admin", role1.Name)
		assert.True(t, role1.SystemRole)

		role2 := res.Roles[1]
		assert.Equal(t, int64(2), role2.Id)
		assert.Equal(t, "authcore.editor", role2.Name)
		assert.True(t, role2.SystemRole)

		role3 := res.Roles[2]
		assert.Equal(t, int64(3), role3.Id)
		assert.Equal(t, "snowdrop.admin", role3.Name)
		assert.False(t, role3.SystemRole)
	}
}

func TestListRolesWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. List roles while the account is not authenticated
	req := &managementapi.ListRolesRequest{}
	_, err := srv.ListRoles(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestListRolesWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List roles while the account is not authorized
	req := &managementapi.ListRolesRequest{}
	_, err = srv.ListRoles(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestCreateRoleAPI(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Create role
	req := &managementapi.CreateRoleRequest{
		Name: "snowdrop.editor",
	}
	res, err := srv.CreateRole(ctx, req)

	if assert.Nil(t, err) {
		assert.Equal(t, "snowdrop.editor", res.Name)
		assert.False(t, res.SystemRole)
	}

	// 2. Create role with the same name again
	req2 := &managementapi.CreateRoleRequest{
		Name: "snowdrop.editor",
	}
	_, err = srv.CreateRole(ctx, req2)

	assert.Error(t, err)
}

func TestCreateRoleWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Create role while the account is not authenticated
	req := &managementapi.CreateRoleRequest{
		Name: "snowdrop.admin",
	}
	_, err := srv.CreateRole(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestCreateRoleWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Create role while the account is not authorized
	req := &managementapi.CreateRoleRequest{
		Name: "snowdrop.admin",
	}
	_, err = srv.CreateRole(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestDeleteRole(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List roles
	req := &managementapi.ListRolesRequest{}
	res, err := srv.ListRoles(ctx, req)

	assert.Nil(t, err)

	// 2. Delete role
	req2 := &managementapi.DeleteRoleRequest{
		RoleId: "3",
	}
	_, err = srv.DeleteRole(ctx, req2)

	assert.Nil(t, err)

	// 3. List roles
	req3 := &managementapi.ListRolesRequest{}
	res3, err := srv.ListRoles(ctx, req3)

	assert.Nil(t, err)
	assert.Len(t, res3.Roles, len(res.Roles)-1)
}

func TestDeleteRoleWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Delete role while the account is not authenticated
	req := &managementapi.DeleteRoleRequest{
		RoleId: "3",
	}
	_, err := srv.DeleteRole(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestDeleteRoleWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Delete role while the account is not authorized
	req := &managementapi.DeleteRoleRequest{
		RoleId: "3",
	}
	_, err = srv.DeleteRole(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestAssignRoleAPI(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Assign role
	req := &managementapi.AssignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err = srv.AssignRole(ctx, req)

	assert.Nil(t, err)

	// 2. Assign the same role to the same user again
	req2 := &managementapi.AssignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err = srv.AssignRole(ctx, req2)

	assert.Error(t, err)
}

func TestAssignRoleWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Assign role while the account is not authenticated
	req := &managementapi.AssignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err := srv.AssignRole(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestAssignRoleWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Assign role while the account is not authorized
	req := &managementapi.AssignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err = srv.AssignRole(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestUnassignRole(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Unassign role
	req := &managementapi.UnassignRoleRequest{
		UserId: "2",
		RoleId: "2",
	}
	_, err = srv.UnassignRole(ctx, req)

	assert.Nil(t, err)
}

func TestUnassignRoleWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. Unassign role while the account is not authenticated
	req := &managementapi.UnassignRoleRequest{
		UserId: "2",
		RoleId: "2",
	}
	_, err := srv.UnassignRole(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestUnassignRoleWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. Unassign role while the account is not authorized
	req := &managementapi.UnassignRoleRequest{
		UserId: "2",
		RoleId: "2",
	}
	_, err = srv.UnassignRole(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestListRoleAssignments(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List role assignments
	req := &managementapi.ListRoleAssignmentsRequest{
		UserId: "1",
	}
	res, err := srv.ListRoleAssignments(ctx, req)

	if assert.Nil(t, err) {
		roles := res.Roles
		assert.Len(t, roles, 2)

		role1 := roles[0]
		assert.Equal(t, "authcore.admin", role1.Name)

		role2 := roles[1]
		assert.Equal(t, "authcore.editor", role2.Name)
	}
}

func TestListRoleAssignmentsWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. List role assignments while the account is not authenticated
	req := &managementapi.ListRoleAssignmentsRequest{
		UserId: "1",
	}
	_, err := srv.ListRoleAssignments(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestListRoleAssignmentsWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List role assignments while the account is not authorized
	req := &managementapi.ListRoleAssignmentsRequest{
		UserId: "1",
	}
	_, err = srv.ListRoleAssignments(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestListPermissionAssignments(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List permission assignments of authcore.admin
	req := &managementapi.ListPermissionAssignmentsRequest{
		RoleId: "1",
	}
	res, err := srv.ListPermissionAssignments(ctx, req)

	if assert.Nil(t, err) {
		// See some examples
		assert.Equal(t, "authcore.users.create", res.Permissions[0].Name)
		assert.Equal(t, "authcore.users.get", res.Permissions[1].Name)
		assert.Equal(t, "authcore.users.list", res.Permissions[2].Name)
	}

	// 2. List permission assignments of authcore.editor
	req2 := &managementapi.ListPermissionAssignmentsRequest{
		RoleId: "2",
	}
	res2, err := srv.ListPermissionAssignments(ctx, req2)

	if assert.Nil(t, err) {
		// Permissions of authcore.admin should be greater than authcore.editor
		assert.True(t, len(res.Permissions) > len(res2.Permissions))
	}
}

func TestListPermissionAssignmentsWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. List permission assignments while the account is not authenticated
	req := &managementapi.ListPermissionAssignmentsRequest{
		RoleId: "1",
	}
	_, err := srv.ListPermissionAssignments(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}

func TestListPermissionAssignmentsWhenUnauthorized(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List permission assignments while the account is not authorized
	req := &managementapi.ListPermissionAssignmentsRequest{
		RoleId: "1",
	}
	_, err = srv.ListPermissionAssignments(ctx, req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.PermissionDenied, status.Code())
		}
	}
}

func TestListCurrentUserPermissions(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// 1. List the permissions of the current user
	req := &managementapi.ListCurrentUserPermissionsRequest{}
	res, err := srv.ListCurrentUserPermissions(ctx, req)

	if assert.Nil(t, err) {
		assert.Equal(t, "authcore.users.create", res.Permissions[0].Name)
		assert.Equal(t, "authcore.users.get", res.Permissions[1].Name)
		assert.Equal(t, "authcore.users.list", res.Permissions[2].Name)
	}
}

func TestListCurrentUserPermissionsWhenUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	// 1. List the permissions of the current user while the user is not authenticated
	req := &managementapi.ListCurrentUserPermissionsRequest{}
	_, err := srv.ListCurrentUserPermissions(context.Background(), req)

	if assert.Error(t, err) {
		status, ok := status.FromError(err)
		if assert.True(t, ok) {
			assert.Equal(t, codes.Unauthenticated, status.Code())
		}
	}
}
