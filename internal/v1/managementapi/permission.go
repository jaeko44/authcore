package managementapi

import (
	"authcore.io/authcore/internal/rbac"
)

// PermissionAssignments is a rbac list for PermissionAssignment.
var PermissionAssignments rbac.PermissionAssignmentList

func init() {
	// Set the permission-role relation here
	PermissionAssignments = rbac.PermissionAssignmentList{
		PA(AdminRole,
			CreateUserPermission,
			GetUserPermission,
			ListUsersPermission,
			UpdateUserPermission,
			DeleteUserPermission,
			LockUserPermission,
			UnlockUserPermission,
			UpdatePasswordPermission,
			ListContactsPermission,
			CreateContactPermission,
			DeleteContactPermission,
			StartVerifyContactPermission,
			UpdatePrimaryContactPermission,
			CreateRolePermission,
			GetRolePermission,
			ListRolesPermission,
			UpdateRolePermission,
			DeleteRolePermission,
			AssignRolePermission,
			UnassignRolePermission,
			ListRoleAssignmentsPermission,
			ListPermissionAssignmentsPermission,
			ListAuditLogsPermission,
			ListSessionsPermission,
			CreateSessionPermission,
			DeleteSessionPermission,
			GetMetadataPermission,
			UpdateMetadataPermission,
			ListSecondFactorsPermission,
			ListOAuthFactorsPermission,
			DeleteOAuthFactorsPermission,
			CreateTemplatePermission,
			ListTemplatesPermission,
			GetTemplatePermission,
			ResetTemplatePermission,
		),
		PA(EditorRole,
			CreateUserPermission,
			GetUserPermission,
			ListUsersPermission,
			UpdateUserPermission,
			DeleteUserPermission,
			LockUserPermission,
			UnlockUserPermission,
			ListContactsPermission,
			CreateContactPermission,
			DeleteContactPermission,
			StartVerifyContactPermission,
			UpdatePrimaryContactPermission,
			GetRolePermission,
			ListRolesPermission,
			ListRoleAssignmentsPermission,
			ListPermissionAssignmentsPermission,
			ListAuditLogsPermission,
			ListSessionsPermission,
			DeleteSessionPermission,
			GetMetadataPermission,
			UpdateMetadataPermission,
			ListSecondFactorsPermission,
			ListOAuthFactorsPermission,
			DeleteOAuthFactorsPermission,
			CreateTemplatePermission,
			ListTemplatesPermission,
			GetTemplatePermission,
			ResetTemplatePermission,
		),
	}
}

// Roles
const (
	AdminRole  = "authcore.admin"
	EditorRole = "authcore.editor"
)

// Permissions
const (
	ListAuditLogsPermission             = "authcore.auditLogs.list"
	CreateContactPermission             = "authcore.contacts.create"
	DeleteContactPermission             = "authcore.contacts.delete"
	ListContactsPermission              = "authcore.contacts.list"
	UpdatePrimaryContactPermission      = "authcore.contacts.updatePrimary"
	StartVerifyContactPermission        = "authcore.contacts.verify"
	GetMetadataPermission               = "authcore.metadata.get"
	UpdateMetadataPermission            = "authcore.metadata.update"
	UpdatePasswordPermission            = "authcore.password.update"
	AssignRolePermission                = "authcore.roles.assign"
	CreateRolePermission                = "authcore.roles.create"
	DeleteRolePermission                = "authcore.roles.delete"
	GetRolePermission                   = "authcore.roles.get"
	ListRolesPermission                 = "authcore.roles.list"
	UnassignRolePermission              = "authcore.roles.unassign"
	UpdateRolePermission                = "authcore.roles.update"
	ListPermissionAssignmentsPermission = "authcore.rolesPermissions.list"
	ListRoleAssignmentsPermission       = "authcore.rolesUsers.list"
	ListSecondFactorsPermission         = "authcore.secondFactors.list"
	CreateUserPermission                = "authcore.users.create"
	DeleteUserPermission                = "authcore.users.delete"
	GetUserPermission                   = "authcore.users.get"
	ListUsersPermission                 = "authcore.users.list"
	LockUserPermission                  = "authcore.users.lock"
	UnlockUserPermission                = "authcore.users.unlock"
	UpdateUserPermission                = "authcore.users.update"
	ListSessionsPermission              = "authcore.sessions.list"
	CreateSessionPermission             = "authcore.sessions.create"
	DeleteSessionPermission             = "authcore.sessions.delete"
	ListOAuthFactorsPermission          = "authcore.oAuthFactors.list"
	DeleteOAuthFactorsPermission        = "authcore.oAuthFactors.delete"
	CreateTemplatePermission            = "authcore.templates.create"
	ListTemplatesPermission             = "authcore.templates.list"
	GetTemplatePermission               = "authcore.templates.get"
	ResetTemplatePermission             = "authcore.templates.reset"
)

// PA returns a rbac PermissionAssignment struct by role name and permissions.
func PA(roleName string, permissions ...rbac.Permission) rbac.PermissionAssignment {
	return rbac.PermissionAssignment{
		Role:        roleName,
		Permissions: permissions,
	}
}
