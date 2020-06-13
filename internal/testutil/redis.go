package testutil

import (
	"log"
	"sync"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
)

var redisOnce sync.Once
var miniRedis *miniredis.Miniredis
var redisClient *redis.Client

// RedisForTest creates a *redis.Client that for tests.
func RedisForTest() *redis.Client {
	redisOnce.Do(func() {
		miniRedis, err := miniredis.Run()
		if err != nil {
			log.Fatalf("failed to create miniredis")
		}
		redisClient = redis.NewClient(&redis.Options{
			Addr: miniRedis.Addr(),
		})
	})
	return redisClient
}
