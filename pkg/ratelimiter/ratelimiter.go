package ratelimiter

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// RateLimiter is a rate limiter limiting user-based requests
type RateLimiter struct {
	redis          *redis.Client
	redisKeyPrefix string
	requestLimit   int64
	duration       time.Duration
}

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter(redis *redis.Client, redisKeyPrefix string, requestLimit int64, duration time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:          redis,
		redisKeyPrefix: redisKeyPrefix,
		requestLimit:   requestLimit,
		duration:       duration,
	}
}

// Check checks if there are too much requests for the rate limiter for the key
func (r *RateLimiter) Check(key string) error {
	redisKey := r.redisKeyPrefix + key
	requestsLen, err := r.redis.LLen(redisKey).Result()
	if err != nil {
		return err
	}
	if requestsLen < r.requestLimit {
		return nil
	}
	requestTimestampString, err := r.redis.LIndex(redisKey, int64(0)).Result()
	if err != nil && err != redis.Nil {
		return err
	} else if err == nil {
		requestTimestamp, err := strconv.ParseInt(requestTimestampString, 10, 64)
		if err != nil {
			return err
		} else if time.Unix(requestTimestamp, 0).Add(r.duration).Before(time.Now()) {
			_, err = r.redis.LPop(redisKey).Result()
			if err != nil {
				return err
			}
		} else {
			return errors.New("too many requests")
		}
	}
	return nil
}

// Increment appends a timestamp to the rate limiter of the current key
func (r *RateLimiter) Increment(key string) error {
	err := r.Check(key)
	if err != nil {
		return err
	}
	redisKey := r.redisKeyPrefix + key
	err = r.redis.RPush(redisKey, time.Now().Unix()).Err()
	if err != nil {
		return err
	}
	err = r.redis.Expire(redisKey, r.duration).Err()
	if err != nil {
		return err
	}
	return nil
}
