package user

import (
	"context"
	"testing"

	"authcore.io/authcore/pkg/nulls"

	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	user := &User{
		ID:       1,
		Username: nulls.NewString("test"),
	}
	ctx := NewContextWithCurrentUser(context.Background(), user)

	userOut, ok := CurrentUserFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, user.ID, userOut.ID)
	assert.Equal(t, user.Username, userOut.Username)
}
