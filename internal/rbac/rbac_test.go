package rbac

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type roleResolver struct{}

func (r roleResolver) ResolveRole(ctx context.Context) ([]Role, error) {
	return []Role{"authcore.editor"}, nil
}

func TestAuthorize(t *testing.T) {
	// Preparing the mock
	permissionAssignmentList := PermissionAssignmentList{
		PermissionAssignment{
			Role: "authcore.admin",
			Permissions: []Permission{
				"authcore.users.read",
				"authcore.users.update",
				"authcore.users.delete",
				"authcore.users.lock",
			},
		},
		PermissionAssignment{
			Role: "authcore.editor",
			Permissions: []Permission{
				"authcore.users.read",
				"authcore.users.update",
			},
		},
	}
	roleResolver := roleResolver{}

	// 1. authcore.users.read is in the permissions of authcore.admin (1/1)
	service := NewService(roleResolver, permissionAssignmentList)
	err := service.Authorize(
		context.TODO(),
		"authcore.users.read",
	)
	assert.Nil(t, err)

	// 2. authcore.users.delete is not in the permissions of authcore.admin (0/1)
	err = service.Authorize(
		context.TODO(),
		"authcore.users.delete",
	)
	assert.Error(t, err)

	// 3. authcore.users.read and authcore.users.update
	//    are in the permissions of authcore.admin (2/2)
	err = service.Authorize(
		context.TODO(),
		"authcore.users.read", "authcore.users.update",
	)
	assert.Nil(t, err)

	// 4. authcore.users.read is in but authcore.users.delete
	//    is not in the permissions of authcore.admin (1/2)
	err = service.Authorize(
		context.TODO(),
		"authcore.users.read", "authcore.users.delete",
	)
	assert.Error(t, err)

	// 5. neither authcore.users.lock nor authcore.users.delete
	//    is in the permissions of authcore.admin (0/2)
	err = service.Authorize(
		context.TODO(),
		"authcore.users.delete", "authcore.users.lock",
	)
	assert.Error(t, err)
}
