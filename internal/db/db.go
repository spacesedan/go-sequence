package db

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)

// reJSONHandler goredis JSON handler connects to the redis pool and handles a
// single query
type reJSONHandler struct {
	*rejson.Handler
}

// NewReJSONHandler
func NewReJSONHandler(conn *redis.Client) reJSONHandler {
	rh := rejson.NewReJSONHandler()
	rh.SetGoRedisClient(conn)

	return reJSONHandler{rh}
}

// lobbyKey helper that returns a string used to associate the lobby in goredis
func lobbyKey(lobby_id string) string {
	return fmt.Sprintf("lobby_id-%v.gamestate", lobby_id)
}

// playerKey helper that returns a string used to associate the player in goredis
func playerKey(lobby_id string, u string) string {
	return fmt.Sprintf("lobby_id-%v|username-%v.playerstate", lobby_id, u)
}
