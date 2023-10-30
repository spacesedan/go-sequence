package redis

import "github.com/go-redis/redis/v8"

type Player struct {
	client *redis.Client
}

func NewPlayer(c *redis.Client) *Player {
	return &Player{
		client: c,
	}
}


