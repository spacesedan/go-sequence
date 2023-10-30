package lobby

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	"github.com/gomodule/redigo/redis"
	"github.com/nitishm/go-rejson/v4"
)

type LobbyState struct {
	redis  *redis.Pool
	logger *slog.Logger
}

func NewLobbyState(r *redis.Pool, l *slog.Logger) *LobbyState {
	return &LobbyState{
		redis:  r,
		logger: l,
	}
}

func (l *LobbyState) SetPlayer(lobby_id string, p *WsConnection) error {
	conn := l.redis.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to communicate to redis-server @ %v", err)
		}
	}()

	rh := rejson.NewReJSONHandler()
	rh.SetRedigoClient(conn)

	res, err := rh.JSONSet(playerKey(lobby_id, p.Username), ".", &PlayerState{
		Username: p.Username,
		LobbyId:  lobby_id,
		Color:    p.Color,
		Ready:    p.IsReady,
	})
	if err != nil {
		return err
	}

	if res.(string) == "OK" {
		l.logger.Info("Successfully set to cache")
	} else {
		l.logger.Info("Failed to cache")
	}
	return nil
}

func (l *LobbyState) GetPlayer(lobby_id string, username string) (*PlayerState, error) {
	l.logger.Info("Getting player from cache",
		slog.String("lobby_id", lobby_id),
		slog.String("username", username))

	var ps *PlayerState
	conn := l.redis.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to communicate to redis-server @ %v", err)
		}
	}()

	rh := rejson.NewReJSONHandler()
	rh.SetRedigoClient(conn)

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

func (l *LobbyState) GetMPlayers(gl *GameLobby) ([]*PlayerState, error) {
	l.logger.Info("Getting all players from cache")
	var ps []*PlayerState
	var playerKeys []string

	conn := l.redis.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to communicate to redis-server @ %v", err)
		}
	}()

	for username := range gl.Players {
		playerKeys = append(playerKeys, playerKey(gl.ID, username))
	}

	rh := rejson.NewReJSONHandler()
	rh.SetRedigoClient(conn)

	pb, err := redis.ByteSlices(rh.JSONMGet(".", playerKeys...))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for _, b := range pb {
		var p *PlayerState
		err = json.Unmarshal(b, &p)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)

	}

	return ps, nil

}

func (l *LobbyState) RemovePlayer(lobby_id string, username string) error {
	conn := l.redis.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to communicate to redis-server @ %v", err)
		}
	}()

	rh := rejson.NewReJSONHandler()
	rh.SetRedigoClient(conn)

	_, err := rh.JSONDel(playerKey(lobby_id, username), ".")
	if err != nil {
		return err
	}

	return nil

}

func playerKey(lobby_id string, u string) string {
	return fmt.Sprintf("lobby_id-%v|username-%v.gamestate", lobby_id, u)
}
