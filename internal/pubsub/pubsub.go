package pubsub

import "github.com/gomodule/redigo/redis"

type PubSub struct {
	*redis.PubSubConn
}

func NewPubSub(r redis.Conn) *PubSub {
    return &PubSub{
        &redis.PubSubConn{Conn: r},
    }
}
