package db

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)

// reJSONHandler goredis JSON handler connects to the redis pool and handles a
// single query
type reJSONHandler struct {
	rj  *rejson.Handler
	rdb *redis.Client
}

// NewReJSONHandler
func NewReJSONHandler(conn *redis.Client) reJSONHandler {
	rh := rejson.NewReJSONHandler()
	rh.SetGoRedisClient(conn)

	return reJSONHandler{rj: rh, rdb: conn}
}

func (r *reJSONHandler) JSONSet(key string, obj interface{}) (interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res, err := r.rj.JSONSet(key, ".", obj)
	if err != nil {
		return res, err
	}
	err = r.rdb.Expire(ctx, key, time.Minute*30).Err()
	if err != nil {
		return res, err
	}

	return res, nil

}

// instead of deleting player connect i want to set a certain amount of time
// allowed for them to reconnect. If they do not come back in an alloted amount of
// time their session information would be deleted from the db
func (r *reJSONHandler) Expire(k string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return r.rdb.Expire(ctx, k, time.Second*30).Val()
}

// lobbyKey helper that returns a string used to associate the lobby in goredis
func lobbyKey(lobby_id string) string {
	return fmt.Sprintf("lobby_id-%v.gamestate", lobby_id)
}

// playerKey helper that returns a string used to associate the player in goredis
func playerKey(lobby_id string, u string) string {
	return fmt.Sprintf("lobby_id-%v|username-%v.playerstate", lobby_id, u)
}
