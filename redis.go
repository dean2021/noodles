package noodles

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

func NewRedisPool(options Options) *redis.Pool {
	redisClient := &redis.Pool{
		MaxIdle:     options.RedisMaxIdle,
		MaxActive:   options.RedisMaxActive,
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", options.RedisAddr)
			if err != nil {
				return nil, err
			}
			if options.RedisPass != "" {
				// 认证
				if _, err := c.Do("AUTH", options.RedisPass); err != nil {
					c.Close()
					return nil, err
				}
			}
			// 选择db
			c.Do("SELECT", options.RedisDB)
			return c, nil
		},
	}
	return redisClient
}
