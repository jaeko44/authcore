package ratelimiter

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
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

func TestRateLimiter(t *testing.T) {
	redis := RedisForTest()
	redisKeyPrefix := "rate_limiter/"
	requests := int64(2)
	duration := 1 * time.Second

	rateLimiter := NewRateLimiter(redis, redisKeyPrefix, requests, duration)

	err := rateLimiter.Check("a_random_key")
	assert.NoError(t, err)
	err = rateLimiter.Increment("a_random_key")
	assert.NoError(t, err)

	err = rateLimiter.Check("a_random_key")
	assert.NoError(t, err)
	err = rateLimiter.Increment("a_random_key")
	assert.NoError(t, err)

	err = rateLimiter.Check("a_random_key")
	assert.Error(t, err) // Returns error as it surpasses the 2 requests/s condition
	err = rateLimiter.Increment("a_random_key")
	assert.Error(t, err)

	time.Sleep(1 * time.Second)

	err = rateLimiter.Check("a_random_key")
	assert.NoError(t, err)
	err = rateLimiter.Increment("a_random_key")
	assert.NoError(t, err)
}
