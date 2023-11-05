package pubsub

import (
	"time"

	"github.com/go-redis/redis/v8"
)

const healthCheckPerion = time.Minute

type Publisher struct {
	*redis.Client
}

func NewPublisher(r *redis.Client) *Publisher {
	return &Publisher{r}
}

type Subscriber struct {
	*redis.Client
}


func NewSubscriber(r *redis.Client) *Subscriber {
	return &Subscriber{r}
}
