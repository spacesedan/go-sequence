package internal

import "github.com/gomodule/redigo/redis"

func NewRedis() (*redis.Pool, error) {
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "0.0.0.0:6379")
		},
	}

	return pool, nil
}
