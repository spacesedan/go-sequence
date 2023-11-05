package pubsub

import (
	"time"

	"github.com/go-redis/redis/v8"
)

const healthCheckPerion = time.Minute

type PubSub struct {
	*redis.PubSub
}

func NewPubSub(r redis.Conn) *PubSub {
    return &PubSub{
        &redis.PubSubConn{Conn: r},
    }
}

func (p *PubSub) Listen(conn redis.Conn) error {
    defer conn.Close()



    return nil
}

