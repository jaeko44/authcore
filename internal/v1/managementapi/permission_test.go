package managementapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"authcore.io/authcore/internal/v1/authentication"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/authapi"
	"authcore.io/authcore/pkg/api/managementapi"
)

// TestAdminPermission test every API to see if it fits the access control defined at permission.go
// for authcore.admin permission group
// Order by swagger doc
func TestAdminPermission(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	assert := assert.New(t)

	// Test API access for authcore.admin
	adminUser, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(err) {
		return
	}

	ctx := user.NewContextWithCurrentUser(context.Background(), adminUser)

	// ListAuditLogs, authcore.auditLogs.list
	req := &managementapi.ListAuditLogsRequest{
		PageSize: 2,
	}

	_, err = srv.ListAuditLogs(ctx, req)
	assert.NoError(err)

	// DeleteOAuthFactor, authcore.oAuthFactors.delete
	req5 := &managementapi.DeleteOAuthFactorRequest{
		Id: "3",
	}
	_, err = srv.DeleteOAuthFactor(ctx, req5)
	assert.NoError(err)

	// ListRoles, authcore.roles.list
	req6 := &managementapi.ListRolesRequest{}
	_, err = srv.ListRoles(ctx, req6)
	assert.NoError(err)

	// CreateRole, authcore.roles.create
	req7 := &managementapi.CreateRoleRequest{
		Name: "snowdrop.editor",
	}
	_, err = srv.CreateRole(ctx, req7)
	assert.NoError(err)

	// DeleteRole, authcore.roles.delete
	req8 := &managementapi.DeleteRoleRequest{
		RoleId: "3",
	}
	_, err = srv.DeleteRole(ctx, req8)
	assert.NoError(err)

	// ListPermissionAssignment, authcore.rolesPermissions.list
	req9 := &managementapi.ListPermissionAssignmentsRequest{
		RoleId: "1",
	}
	_, err = srv.ListPermissionAssignments(ctx, req9)
	assert.NoError(err)

	// ListSessions, authcore.sessions.list
	req10 := &managementapi.ListSessionsRequest{
		UserId:   "1",
		PageSize: 2,
	}
	_, err = srv.ListSessions(ctx, req10)
	assert.NoError(err)

	// CreateSession, authcore.sessions.create
	req11 := &managementapi.CreateSessionRequest{
		UserId:   "1",
		DeviceId: "0",
	}
	_, err = srv.CreateSession(ctx, req11)
	assert.NoError(err)

	// DeleteSession, authcore.sessions.delete
	req12 := &managementapi.DeleteSessionRequest{
		SessionId: "1",
	}
	_, err = srv.DeleteSession(ctx, req12)
	assert.NoError(err)

	// ListUsers, authcore.users.list
	req13 := &managementapi.ListUsersRequest{
		PageSize: 2,
	}
	_, err = srv.ListUsers(ctx, req13)
	assert.NoError(err)

	// CreateUser, authcore.users.create
	req14 := &managementapi.CreateUserRequest{
		Username:    "alice",
		Email:       "alice@example.com",
		Phone:       "+85212345678",
		DisplayName: "Alice",
	}
	_, err = srv.CreateUser(ctx, req14)
	assert.NoError(err)

	// ListCurrentUserPermissions, nil
	req15 := &managementapi.ListCurrentUserPermissionsRequest{}
	_, err = srv.ListCurrentUserPermissions(ctx, req15)
	assert.NoError(err)

	// ChangePassword, authcore.password.update
	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	if !assert.NoError(err) {
		return
	}

	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte("new_password"), clientIdentity, serverIdentity, spake2)
	if !assert.NoError(err) {
		return
	}
	req16 := &managementapi.ChangePasswordRequest{
		UserId: "1",
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
	}
	_, err = srv.ChangePassword(ctx, req16)
	assert.NoError(err)

	// GetUser, authcore.users.get
	req17 := &managementapi.GetUserRequest{
		UserId: "1",
	}
	user, err := srv.GetUser(ctx, req17)
	if !assert.NoError(err) {
		return
	}

	// UpdateUser, authcore.users.update
	user.DisplayName = "bob_updated"
	req18 := &managementapi.UpdateUserRequest{
		UserId: user.Id,
		User:   user,
	}
	_, err = srv.UpdateUser(ctx, req18)
	assert.NoError(err)

	// GetMetadata, authcore.metadata.get
	req21 := &managementapi.GetMetadataRequest{
		UserId: "1",
	}
	_, err = srv.GetMetadata(ctx, req21)
	assert.NoError(err)

	// UpdateMetadata, authcore.metadata.update
	req22 := &managementapi.UpdateMetadataRequest{
		UserId:       "1",
		UserMetadata: `{"favourite_links":["https://github.com","https://google.com","https://blocksq.com"]}`,
		AppMetadata:  `{"kyc_status":true}`,
	}
	_, err = srv.UpdateMetadata(ctx, req22)
	assert.NoError(err)

	// ListOAuthFactor, authcore.oAuthFactors.list
	req23 := &managementapi.ListOAuthFactorsRequest{
		UserId: "10",
	}
	_, err = srv.ListOAuthFactors(ctx, req23)
	assert.NoError(err)

	// ListRoleAssignments, authcore.rolesUsers.list
	req24 := &managementapi.ListRoleAssignmentsRequest{
		UserId: "1",
	}
	_, err = srv.ListRoleAssignments(ctx, req24)
	assert.NoError(err)

	// AssignRole, authcore.roles.assign
	req25 := &managementapi.AssignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err = srv.AssignRole(ctx, req25)
	if !assert.NoError(err) {
		return
	}

	// UnassignRole, authcore.roles.unassign
	req26 := &managementapi.UnassignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err = srv.UnassignRole(ctx, req26)
	assert.NoError(err)

	// ListSecondFactors, authcore.secondFactors.list
	req27 := &managementapi.ListSecondFactorsRequest{
		UserId: "3",
	}
	_, err = srv.ListSecondFactors(ctx, req27)
	assert.NoError(err)
}

// Test API access for authcore.editor
func TestEditorPermission(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	assert := assert.New(t)

	// Test API access for authcore.editor
	editorUser, err := srv.UserStore.UserByID(context.Background(), 2)
	if !assert.NoError(err) {
		return
	}

	ctx := user.NewContextWithCurrentUser(context.Background(), editorUser)

	// ListAuditLogs, authcore.auditLogs.list
	req := &managementapi.ListAuditLogsRequest{
		PageSize: 2,
	}

	_, err = srv.ListAuditLogs(ctx, req)
	assert.NoError(err)

	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	// DeleteOAuthFactor, authcore.oAuthFactors.delete
	req5 := &managementapi.DeleteOAuthFactorRequest{
		Id: "3",
	}
	_, err = srv.DeleteOAuthFactor(ctx, req5)
	assert.NoError(err)

	// ListRoles, authcore.roles.list
	req6 := &managementapi.ListRolesRequest{}
	_, err = srv.ListRoles(ctx, req6)
	assert.NoError(err)

	// CreateRole, authcore.roles.create
	req7 := &managementapi.CreateRoleRequest{
		Name: "snowdrop.editor",
	}
	_, err = srv.CreateRole(ctx, req7)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// DeleteRole, authcore.roles.delete
	req8 := &managementapi.DeleteRoleRequest{
		RoleId: "3",
	}
	_, err = srv.DeleteRole(ctx, req8)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// ListPermissionAssignment, authcore.rolesPermissions.list
	req9 := &managementapi.ListPermissionAssignmentsRequest{
		RoleId: "1",
	}
	_, err = srv.ListPermissionAssignments(ctx, req9)
	assert.NoError(err)

	// ListSessions, authcore.sessions.list
	req10 := &managementapi.ListSessionsRequest{
		UserId:   "1",
		PageSize: 2,
	}
	_, err = srv.ListSessions(ctx, req10)
	assert.NoError(err)

	// CreateSession, authcore.sessions.create
	req11 := &managementapi.CreateSessionRequest{
		UserId:   "1",
		DeviceId: "0",
	}
	_, err = srv.CreateSession(ctx, req11)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}
	// DeleteSession, authcore.sessions.delete
	req12 := &managementapi.DeleteSessionRequest{
		SessionId: "1",
	}
	_, err = srv.DeleteSession(ctx, req12)
	assert.NoError(err)

	// ListUsers, authcore.users.list
	req13 := &managementapi.ListUsersRequest{
		PageSize: 2,
	}
	_, err = srv.ListUsers(ctx, req13)
	assert.NoError(err)

	// CreateUser, authcore.users.create
	req14 := &managementapi.CreateUserRequest{
		Username:    "alice",
		Email:       "alice@example.com",
		Phone:       "+85212345678",
		DisplayName: "Alice",
	}
	_, err = srv.CreateUser(ctx, req14)
	assert.NoError(err)

	// ListCurrentUserPermissions, nil
	req15 := &managementapi.ListCurrentUserPermissionsRequest{}
	_, err = srv.ListCurrentUserPermissions(ctx, req15)
	assert.NoError(err)

	// ChangePassword, authcore.password.update
	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	if !assert.NoError(err) {
		return
	}

	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte("new_password"), clientIdentity, serverIdentity, spake2)
	if !assert.NoError(err) {
		return
	}
	req16 := &managementapi.ChangePasswordRequest{
		UserId: "1",
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
	}
	_, err = srv.ChangePassword(ctx, req16)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// GetUser, authcore.users.get
	req17 := &managementapi.GetUserRequest{
		UserId: "1",
	}
	user, err := srv.GetUser(ctx, req17)
	if !assert.NoError(err) {
		return
	}

	// UpdateUser, authcore.users.update
	user.DisplayName = "bob_updated"
	req18 := &managementapi.UpdateUserRequest{
		UserId: user.Id,
		User:   user,
	}
	_, err = srv.UpdateUser(ctx, req18)
	assert.NoError(err)

	// GetMetadata, authcore.metadata.get
	req21 := &managementapi.GetMetadataRequest{
		UserId: "1",
	}
	_, err = srv.GetMetadata(ctx, req21)
	assert.NoError(err)

	// UpdateMetadata, authcore.metadata.update
	req22 := &managementapi.UpdateMetadataRequest{
		UserId:       "1",
		UserMetadata: `{"favourite_links":["https://github.com","https://google.com","https://blocksq.com"]}`,
		AppMetadata:  `{"kyc_status":true}`,
	}
	_, err = srv.UpdateMetadata(ctx, req22)
	assert.NoError(err)

	// ListOAuthFactor, authcore.oAuthFactors.list
	req23 := &managementapi.ListOAuthFactorsRequest{
		UserId: "10",
	}
	_, err = srv.ListOAuthFactors(ctx, req23)
	assert.NoError(err)

	// ListRoleAssignments, authcore.rolesUsers.list
	req24 := &managementapi.ListRoleAssignmentsRequest{
		UserId: "1",
	}
	_, err = srv.ListRoleAssignments(ctx, req24)
	assert.NoError(err)

	// AssignRole, authcore.roles.assign
	req25 := &managementapi.AssignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err = srv.AssignRole(ctx, req25)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// UnassignRole, authcore.roles.unassign
	req26 := &managementapi.UnassignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err = srv.UnassignRole(ctx, req26)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// ListSecondFactors, authcore.secondFactors.list
	req27 := &managementapi.ListSecondFactorsRequest{
		UserId: "3",
	}
	_, err = srv.ListSecondFactors(ctx, req27)
	assert.NoError(err)
}

// Test API access for normal user
func TestNormalUserPermission(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	assert := assert.New(t)

	// Test API access for normal user
	u, err := srv.UserStore.UserByID(context.Background(), 3)
	if !assert.NoError(err) {
		return
	}

	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	// ListAuditLogs, authcore.auditLogs.list
	req := &managementapi.ListAuditLogsRequest{
		PageSize: 2,
	}

	_, err = srv.ListAuditLogs(ctx, req)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	m := make(map[string]string)
	m["grpcgateway-origin"] = "http://0.0.0.0:8000"
	ctx = metadata.NewIncomingContext(ctx, metadata.New(m))

	// DeleteOAuthFactor, authcore.oAuthFactors.delete
	req5 := &managementapi.DeleteOAuthFactorRequest{
		Id: "3",
	}
	_, err = srv.DeleteOAuthFactor(ctx, req5)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// ListRoles, authcore.roles.list
	req6 := &managementapi.ListRolesRequest{}
	_, err = srv.ListRoles(ctx, req6)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// CreateRole, authcore.roles.create
	req7 := &managementapi.CreateRoleRequest{
		Name: "snowdrop.editor",
	}
	_, err = srv.CreateRole(ctx, req7)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// DeleteRole, authcore.roles.delete
	req8 := &managementapi.DeleteRoleRequest{
		RoleId: "3",
	}
	_, err = srv.DeleteRole(ctx, req8)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// ListPermissionAssignment, authcore.rolesPermissions.list
	req9 := &managementapi.ListPermissionAssignmentsRequest{
		RoleId: "1",
	}
	_, err = srv.ListPermissionAssignments(ctx, req9)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// ListSessions, authcore.sessions.list
	req10 := &managementapi.ListSessionsRequest{
		UserId:   "1",
		PageSize: 2,
	}
	_, err = srv.ListSessions(ctx, req10)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// CreateSession, authcore.sessions.create
	req11 := &managementapi.CreateSessionRequest{
		UserId:   "1",
		DeviceId: "0",
	}
	_, err = srv.CreateSession(ctx, req11)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}
	// DeleteSession, authcore.sessions.delete
	req12 := &managementapi.DeleteSessionRequest{
		SessionId: "1",
	}
	_, err = srv.DeleteSession(ctx, req12)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// ListUsers, authcore.users.list
	req13 := &managementapi.ListUsersRequest{
		PageSize: 2,
	}
	_, err = srv.ListUsers(ctx, req13)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// CreateUser, authcore.users.create
	req14 := &managementapi.CreateUserRequest{
		Username:    "alice",
		Email:       "alice@example.com",
		Phone:       "+85212345678",
		DisplayName: "Alice",
	}
	_, err = srv.CreateUser(ctx, req14)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// ListCurrentUserPermissions, nil
	req15 := &managementapi.ListCurrentUserPermissionsRequest{}
	_, err = srv.ListCurrentUserPermissions(ctx, req15)
	assert.NoError(err)

	// ChangePassword, authcore.password.update
	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	if !assert.NoError(err) {
		return
	}

	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte("new_password"), clientIdentity, serverIdentity, spake2)
	if !assert.NoError(err) {
		return
	}
	req16 := &managementapi.ChangePasswordRequest{
		UserId: "1",
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
	}
	_, err = srv.ChangePassword(ctx, req16)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// GetUser, authcore.users.get
	req17 := &managementapi.GetUserRequest{
		UserId: "1",
	}
	_, err = srv.GetUser(ctx, req17)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// UpdateUser, authcore.users.update
	req18 := &managementapi.UpdateUserRequest{
		UserId: "1",
		User: &authapi.User{
			DisplayName: "bob_updated",
		},
	}
	_, err = srv.UpdateUser(ctx, req18)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// GetMetadata, authcore.metadata.get
	req21 := &managementapi.GetMetadataRequest{
		UserId: "1",
	}
	_, err = srv.GetMetadata(ctx, req21)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// UpdateMetadata, authcore.metadata.update
	req22 := &managementapi.UpdateMetadataRequest{
		UserId:       "1",
		UserMetadata: `{"favourite_links":["https://github.com","https://google.com","https://blocksq.com"]}`,
		AppMetadata:  `{"kyc_status":true}`,
	}
	_, err = srv.UpdateMetadata(ctx, req22)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// ListOAuthFactor, authcore.oAuthFactors.list
	req23 := &managementapi.ListOAuthFactorsRequest{
		UserId: "10",
	}
	_, err = srv.ListOAuthFactors(ctx, req23)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// ListRoleAssignments, authcore.rolesUsers.list
	req24 := &managementapi.ListRoleAssignmentsRequest{
		UserId: "1",
	}
	_, err = srv.ListRoleAssignments(ctx, req24)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// AssignRole, authcore.roles.assign
	req25 := &managementapi.AssignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err = srv.AssignRole(ctx, req25)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// UnassignRole, authcore.roles.unassign
	req26 := &managementapi.UnassignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err = srv.UnassignRole(ctx, req26)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}

	// ListSecondFactors, authcore.secondFactors.list
	req27 := &managementapi.ListSecondFactorsRequest{
		UserId: "3",
	}
	_, err = srv.ListSecondFactors(ctx, req27)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.PermissionDenied, status.Code())
		}
	}
}

// Test API access for unauthenticated user
func TestUnauthenticated(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	assert := assert.New(t)

	// Test API access for normal user
	ctx := context.Background()

	// ListAuditLogs, authcore.auditLogs.list
	req := &managementapi.ListAuditLogsRequest{
		PageSize: 2,
	}

	_, err := srv.ListAuditLogs(ctx, req)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// DeleteOAuthFactor, authcore.oAuthFactors.delete
	req5 := &managementapi.DeleteOAuthFactorRequest{
		Id: "3",
	}
	_, err = srv.DeleteOAuthFactor(ctx, req5)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// ListRoles, authcore.roles.list
	req6 := &managementapi.ListRolesRequest{}
	_, err = srv.ListRoles(ctx, req6)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// CreateRole, authcore.roles.create
	req7 := &managementapi.CreateRoleRequest{
		Name: "snowdrop.editor",
	}
	_, err = srv.CreateRole(ctx, req7)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// DeleteRole, authcore.roles.delete
	req8 := &managementapi.DeleteRoleRequest{
		RoleId: "3",
	}
	_, err = srv.DeleteRole(ctx, req8)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// ListPermissionAssignment, authcore.rolesPermissions.list
	req9 := &managementapi.ListPermissionAssignmentsRequest{
		RoleId: "1",
	}
	_, err = srv.ListPermissionAssignments(ctx, req9)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// ListSessions, authcore.sessions.list
	req10 := &managementapi.ListSessionsRequest{
		UserId:   "1",
		PageSize: 2,
	}
	_, err = srv.ListSessions(ctx, req10)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// CreateSession, authcore.sessions.create
	req11 := &managementapi.CreateSessionRequest{
		UserId:   "1",
		DeviceId: "0",
	}
	_, err = srv.CreateSession(ctx, req11)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}
	// DeleteSession, authcore.sessions.delete
	req12 := &managementapi.DeleteSessionRequest{
		SessionId: "1",
	}
	_, err = srv.DeleteSession(ctx, req12)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// ListUsers, authcore.users.list
	req13 := &managementapi.ListUsersRequest{
		PageSize: 2,
	}
	_, err = srv.ListUsers(ctx, req13)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// CreateUser, authcore.users.create
	req14 := &managementapi.CreateUserRequest{
		Username:    "alice",
		Email:       "alice@example.com",
		Phone:       "+85212345678",
		DisplayName: "Alice",
	}
	_, err = srv.CreateUser(ctx, req14)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// ListCurrentUserPermissions, nil
	req15 := &managementapi.ListCurrentUserPermissionsRequest{}
	_, err = srv.ListCurrentUserPermissions(ctx, req15)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// ChangePassword, authcore.password.update
	clientIdentity, serverIdentity := authentication.GetIdentity()
	spake2, err := authentication.NewSPAKE2Plus()
	if !assert.NoError(err) {
		return
	}

	salt, verifierW0, verifierL, err := authentication.GenerateSPAKE2Verifier([]byte("new_password"), clientIdentity, serverIdentity, spake2)
	if !assert.NoError(err) {
		return
	}
	req16 := &managementapi.ChangePasswordRequest{
		UserId: "1",
		PasswordVerifier: &authapi.PasswordVerifier{
			Salt:       salt,
			VerifierW0: verifierW0,
			VerifierL:  verifierL,
		},
	}
	_, err = srv.ChangePassword(ctx, req16)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// GetUser, authcore.users.get
	req17 := &managementapi.GetUserRequest{
		UserId: "1",
	}
	_, err = srv.GetUser(ctx, req17)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// UpdateUser, authcore.users.update
	req18 := &managementapi.UpdateUserRequest{
		UserId: "1",
		User: &authapi.User{
			DisplayName: "bob_updated",
		},
	}
	_, err = srv.UpdateUser(ctx, req18)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// GetMetadata, authcore.metadata.get
	req21 := &managementapi.GetMetadataRequest{
		UserId: "1",
	}
	_, err = srv.GetMetadata(ctx, req21)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// UpdateMetadata, authcore.metadata.update
	req22 := &managementapi.UpdateMetadataRequest{
		UserId:       "1",
		UserMetadata: `{"favourite_links":["https://github.com","https://google.com","https://blocksq.com"]}`,
		AppMetadata:  `{"kyc_status":true}`,
	}
	_, err = srv.UpdateMetadata(ctx, req22)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// ListOAuthFactor, authcore.oAuthFactors.list
	req23 := &managementapi.ListOAuthFactorsRequest{
		UserId: "10",
	}
	_, err = srv.ListOAuthFactors(ctx, req23)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// ListRoleAssignments, authcore.rolesUsers.list
	req24 := &managementapi.ListRoleAssignmentsRequest{
		UserId: "1",
	}
	_, err = srv.ListRoleAssignments(ctx, req24)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// AssignRole, authcore.roles.assign
	req25 := &managementapi.AssignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err = srv.AssignRole(ctx, req25)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// UnassignRole, authcore.roles.unassign
	req26 := &managementapi.UnassignRoleRequest{
		UserId: "3",
		RoleId: "2",
	}
	_, err = srv.UnassignRole(ctx, req26)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}

	// ListSecondFactors, authcore.secondFactors.list
	req27 := &managementapi.ListSecondFactorsRequest{
		UserId: "3",
	}
	_, err = srv.ListSecondFactors(ctx, req27)
	if assert.Error(err) {
		status, ok := status.FromError(err)
		if assert.True(ok) {
			assert.Equal(codes.Unauthenticated, status.Code())
		}
	}
}
