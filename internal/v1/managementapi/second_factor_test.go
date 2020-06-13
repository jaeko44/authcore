package managementapi

import (
	"context"
	"encoding/json"
	"testing"

	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/api/managementapi"

	"github.com/stretchr/testify/assert"
)

// Lists
func TestListSecondFactors(t *testing.T) {
	srv, teardown := ServiceForTest()
	defer teardown()

	currentUser, err := srv.UserStore.UserByID(context.Background(), 1)
	assert.NoError(t, err)
	ctx := user.NewContextWithCurrentUser(context.Background(), currentUser)

	// 1. List the second factors
	req := &managementapi.ListSecondFactorsRequest{
		UserId: "3",
	}
	res, err := srv.ListSecondFactors(ctx, req)
	assert.NoError(t, err)
	assert.Len(t, res.SecondFactors, 2)
	jRes, err := json.Marshal(res)
	assert.NoError(t, err)
	assert.Contains(t, string(jRes), "Factor's jPhone")                     // contains the identifier
	assert.NotContains(t, string(jRes), "THISISAWEAKTOTPSECRETFORTESTSXX2") // does not contain the secret
}
