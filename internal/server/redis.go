package server

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

// NewRedisClientFromConfig creates a redis client according to config values.
func NewRedisClientFromConfig() *redis.Client {
	if viper.GetBool("redis_sentinel_enabled") {
		return redis.NewFailoverClient(NewRedisFailoverOptionsFromConfig())
	}
	return redis.NewClient(NewRedisOptionsFromConfig())
}

// NewRedisOptionsFromConfig creates a redis.Options according to config values.
func NewRedisOptionsFromConfig() *redis.Options {
	addr := viper.GetString("redis_address")
	pass := viper.GetString("redis_password")
	db := viper.GetInt("redis_db")
	return &redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	}
}

// NewRedisFailoverOptionsFromConfig creates a Redis Sentinel options according to config values.
func NewRedisFailoverOptionsFromConfig() *redis.FailoverOptions {
	senteinelAddrs := viper.GetStringSlice("redis_sentinel_addresses")
	masterName := viper.GetString("redis_sentinel_master_name")
	pass := viper.GetString("redis_password")
	db := viper.GetInt("redis_db")
	return &redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: senteinelAddrs,
		Password:      pass,
		DB:            db,
	}
}
