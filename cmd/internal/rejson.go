package internal

import (
	"github.com/gomodule/redigo/redis"
	"github.com/nitishm/go-rejson/v4"
)

// NewReJSONHandler
func NewReJSONHandler(r *redis.Pool) (*rejson.Handler, error) {
	rh := rejson.NewReJSONHandler()
	rh.SetRedigoClient(r.Get())

	return rh, nil
}
