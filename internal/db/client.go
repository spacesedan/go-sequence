package db

import (
	"encoding/json"
	"fmt"
	"log/slog"

	goredis "github.com/go-redis/redis/v8"
	"github.com/gomodule/redigo/redis"
	"github.com/spacesedan/go-sequence/internal"
)

type ClientRepo interface {
	GetPlayer(lobbyID string, username string) (*internal.Player, error)
	GetMPlayers(lobbyID string, players []string) ([]*internal.Player, error)
}

type clientRepo struct {
	redisClient *goredis.Client
	logger      *slog.Logger
}

func NewClientRepo(r *goredis.Client, l *slog.Logger) ClientRepo {
	return &clientRepo{
		redisClient: r,
		logger:      l,
	}
}



// GetPlayer gets a player from the db using the lobby id and player username
func (c *clientRepo) GetPlayer(lobby_id string, username string) (*internal.Player, error) {
	c.logger.Info("lobbyRepo.GetPlayer",
		slog.Group("reading player from db",
			slog.String("lobby_id", lobby_id),
			slog.String("username", username)))

	var ps *internal.Player

	rh := NewReJSONHandler(c.redisClient)

	pj, err := redis.Bytes(rh.JSONGet(playerKey(lobby_id, username), "."))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(pj, &ps)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

// GetMPlayers gets multiple players using lobby details
func (c *clientRepo) GetMPlayers(lobbyID string, players []string) ([]*internal.Player, error) {
	c.logger.Info("lobbyRepo.GetMPlayers",
		slog.Group("reading all players from db for the lobby"))

	var ps []*internal.Player
	var playerKeys []string

	for _, username := range players {
		playerKeys = append(playerKeys, playerKey(lobbyID, username))
	}

	rh := NewReJSONHandler(c.redisClient)

	pb, err := redis.ByteSlices(rh.JSONMGet(".", playerKeys...))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for _, b := range pb {
		var p *internal.Player
		err = json.Unmarshal(b, &p)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)

	}

	return ps, nil

}

