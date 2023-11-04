package lobby

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	"github.com/gomodule/redigo/redis"
	"github.com/nitishm/go-rejson/v4"
)

// reJSONHandler redis JSON handler connects to the redis pool and handles a
// single query
type reJSONHandler struct {
	*rejson.Handler
}

// NewReJSONHandler
func NewReJSONHandler(conn redis.Conn) reJSONHandler {
	rh := rejson.NewReJSONHandler()
	rh.SetRedigoClient(conn)

	return reJSONHandler{rh}

}

// Current state ... are the players in the lobby still choosing thier colors,
// or are they in the game
type CurrentState uint

const (
	Unknown CurrentState = iota
	Lobby
	Game
)

// String get a stringified version of the current game state
func (c CurrentState) String() string {
	switch c {
	case Lobby:
		return "lobby"
	case Game:
		return "game"
	default:
		return "unknown"

	}
}

// LobbyRepo responsible for interfacing with the data stored in the cache
type LobbyRepo struct {
	redis  *redis.Pool
	logger *slog.Logger
}

func NewLobbyRepo(r *redis.Pool, l *slog.Logger) *LobbyRepo {
	l.Info("lobby.NewLobbyRepo: created new lobby repo")
	return &LobbyRepo{
		redis:  r,
		logger: l,
	}
}

type LobbyState struct {
	CurrentState    CurrentState
	Players         map[string]*PlayerState
	ColorsAvailable map[string]bool
	Settings        Settings
}

func (l *LobbyRepo) SetLobby(lobby *GameLobby) error {
	l.logger.Info("lobbyRepo.SetLobby",
		slog.Group("writing to cache",
			slog.String("lobby_id", lobby.ID)))

	conn := l.redis.Get()
	rh := NewReJSONHandler(conn)
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to communicate to redis-server @ %v", err)
		}
	}()

	res, err := rh.JSONSet(lobbyKey(lobby.ID), ".", &LobbyState{
		CurrentState:    Lobby,
		Settings:        lobby.Settings,
		ColorsAvailable: lobby.AvailableColors,
		Players:         lobby.Players,
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

func (l *LobbyRepo) GetLobby(lobby_id string) (*LobbyState, error) {
	l.logger.Info("lobbyRepo.GetLobby",
		slog.Group("reading from cache",
			slog.String("lobby_id", lobby_id)))

	return nil, nil
}

type PlayerState struct {
	LobbyId  string `json:"lobby_id"`
	Username string `json:"username"`
	Color    string `json:"color"`
	Ready    bool   `json:"ready"`
}

func (l *LobbyRepo) SetPlayer(lobby_id string, p *WsConnection) error {
	l.logger.Info("lobbyRepo.SetPlayer",
		slog.Group("writing player to cache",
			slog.String("player", p.Username)))

	conn := l.redis.Get()
	rh := NewReJSONHandler(conn)
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to communicate to redis-server @ %v", err)
		}
	}()

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

func (l *LobbyRepo) GetPlayer(lobby_id string, username string) (*PlayerState, error) {
	l.logger.Info("Getting player from cache",
		slog.String("lobby_id", lobby_id),
		slog.String("username", username))

	var ps *PlayerState
	conn := l.redis.Get()
	rh := NewReJSONHandler(conn)
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to communicate to redis-server @ %v", err)
		}
	}()

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

func (l *LobbyRepo) GetMPlayers(gl *GameLobby) ([]*PlayerState, error) {
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

func (l *LobbyRepo) RemovePlayer(lobby_id string, username string) error {
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

func lobbyKey(lobby_id string) string {
	return fmt.Sprintf("lobby_id-%v.gamestate", lobby_id)
}

func playerKey(lobby_id string, u string) string {
	return fmt.Sprintf("lobby_id-%v|username-%v.gamestate", lobby_id, u)
}
