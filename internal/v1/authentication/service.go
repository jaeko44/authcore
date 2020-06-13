package authentication

import (
	"authcore.io/authcore/internal/user"

	"github.com/go-redis/redis"
)

// Service provides services related to authentication.
type Service struct {
	Redis     *redis.Client
	UserStore *user.Store
}

// NewService initialize a new Service.
func NewService(redis *redis.Client, userStore *user.Store) *Service {
	return &Service{
		Redis:     redis,
		UserStore: userStore,
	}
}
