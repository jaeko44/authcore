package server

import (
	"context"
	"testing"

	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/nulls"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestNewAuthorizationUnaryInterceptor(t *testing.T) {
	accessTokenResolver := func(ctx context.Context, token string) (string, string, error) {
		return "1", "test", nil
	}

	userResolver := func(ctx context.Context, userID string) (*user.User, error) {
		return &user.User{
			ID:       1,
			Username: nulls.NewString("testing"),
		}, nil
	}

	sessionResolver := func(ctx context.Context, sessionID string) (*session.Session, error) {
		return &session.Session{
			ID:     1,
			UserID: 1,
		}, nil
	}

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (res interface{}, err error) {
		handlerCalled = true
		return "response", nil
	}

	interceptor := NewAuthorizationUnaryInterceptor(accessTokenResolver, userResolver, sessionResolver)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer testtest"))
	res, err := interceptor(ctx, "request", nil, handler)
	if assert.NoError(t, err) {
		assert.Equal(t, "response", res)
		assert.True(t, handlerCalled)
	}
}
