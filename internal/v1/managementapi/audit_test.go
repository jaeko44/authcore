package managementapi

import (
	"context"
	"testing"

	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/managementapi"

	"github.com/stretchr/testify/assert"
)

func TestListAuditLogs(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &managementapi.ListAuditLogsRequest{
		PageSize: 2,
	}
	res, err := srv.ListAuditLogs(ctx, req)
	auditLogs := res.AuditLogs
	nextPageToken := res.NextPageToken

	if assert.NoError(t, err) {
		// assert.Equal(t, int32(7), totalSize)
		assert.NotEmpty(t, nextPageToken)
		assert.Len(t, auditLogs, 2)
		assert.Equal(t, "7", auditLogs[0].Id)
	}
}

func TestListAuditLogsFilterByUser(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &managementapi.ListAuditLogsRequest{
		UserId:   "2",
		PageSize: 2,
	}
	res, err := srv.ListAuditLogs(ctx, req)
	auditLogs := res.AuditLogs
	nextPageToken := res.NextPageToken

	if assert.NoError(t, err) {
		// assert.Equal(t, int32(3), totalSize)
		assert.NotEmpty(t, nextPageToken)
		assert.Len(t, auditLogs, 2)
		assert.Equal(t, "6", auditLogs[0].Id)
	}
}

func TestListAuditLogsFilterByUserNoLogs(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	u, err := srv.UserStore.UserByID(context.Background(), 1)
	if !assert.NoError(t, err) {
		return
	}
	ctx := user.NewContextWithCurrentUser(context.Background(), u)

	req := &managementapi.ListAuditLogsRequest{
		UserId:   "4",
		PageSize: 2,
	}
	res, err := srv.ListAuditLogs(ctx, req)
	auditLogs := res.AuditLogs

	if assert.NoError(t, err) {
		assert.Len(t, auditLogs, 0)
	}
}
