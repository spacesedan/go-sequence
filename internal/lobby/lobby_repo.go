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

// NewLobbyRepo creates a LobbyRepo instance to interact with the redis server
func NewLobbyRepo(r *redis.Pool, l *slog.Logger) *LobbyRepo {
	l.Info("lobby.NewLobbyRepo: created new lobby repo")
	return &LobbyRepo{
		redis:  r,
		logger: l,
	}
}

type LobbyState struct {
	CurrentState    CurrentState
	Players         map[string]struct{}
	ColorsAvailable map[string]bool
	Settings        Settings
}

// SetLobby sets and updates the lobby state stored in the db using the lobby
func (l *LobbyRepo) SetLobby(lobby *GameLobby) error {
	l.logger.Info("lobbyRepo.SetLobby",
		slog.Group("writing lobby to db",
			slog.String("lobby_id", lobby.ID)))

	conn := l.redis.Get()
	rh := NewReJSONHandler(conn)
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to communicate to redis-server @ %v", err)
		}
	}()

	_, err := rh.JSONSet(lobbyKey(lobby.ID), ".", &LobbyState{
		CurrentState:    Lobby,
		Settings:        lobby.Settings,
		ColorsAvailable: lobby.AvailableColors,
		Players:         lobby.Players,
	})

	if err != nil {
		return err
	}

	return nil
}

// GetLobby gets a lobby from teh db using the lobby id
func (l *LobbyRepo) GetLobby(lobby_id string) (*LobbyState, error) {
	l.logger.Info("lobbyRepo.GetLobby",
		slog.Group("reading lobby from redis",
			slog.String("lobby_id", lobby_id)))

	var lobbyState *LobbyState

	conn := l.redis.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	rh := NewReJSONHandler(conn)
	lobbyJSON, err := redis.Bytes(rh.JSONGet(lobbyKey(lobby_id), "."))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(lobbyJSON, &lobbyState)
	if err != nil {
		return nil, err
	}

	return lobbyState, nil
}

// DeleteLobby deletes a lobby from the db using the lobby id
func (l *LobbyRepo) DeleteLobby(lobby_id string) error {
	l.logger.Info("lobbyRepo.DeleteLobby",
		slog.Group("deleting lobby from redis",
			slog.String("lobby_id", lobby_id)))

	conn := l.redis.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	rh := NewReJSONHandler(conn)

	if _, err := rh.JSONDel(lobbyKey(lobby_id), "."); err != nil {
		return err
	}
	return nil
}

type PlayerState struct {
	LobbyId  string `json:"lobby_id"`
	Username string `json:"username"`
	Color    string `json:"color"`
	Ready    bool   `json:"ready"`
}

// SetPlayer sets and updates player data in the db
func (l *LobbyRepo) SetPlayer(lobby_id string, p *WsClient) error {
	l.logger.Info("lobbyRepo.SetPlayer",
		slog.Group("writing player to db",
			slog.String("player", p.Username)))

	conn := l.redis.Get()
	rh := NewReJSONHandler(conn)
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to communicate to redis-server @ %v", err)
		}
	}()

	if _, err := rh.JSONSet(playerKey(lobby_id, p.Username), ".", &PlayerState{
		Username: p.Username,
		LobbyId:  lobby_id,
		Color:    p.Color,
		Ready:    p.IsReady,
	}); err != nil {
		return err
	}

	return nil
}

// GetPlayer gets a player from the db using the lobby id and player username
func (l *LobbyRepo) GetPlayer(lobby_id string, username string) (*PlayerState, error) {
	l.logger.Info("lobbyRepo.GetPlayer",
		slog.Group("reading player from db",
			slog.String("lobby_id", lobby_id),
			slog.String("username", username)))

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

// GetMPlayers gets multiple players using lobby details
func (l *LobbyRepo) GetMPlayers(gl *GameLobby) ([]*PlayerState, error) {
	l.logger.Info("lobbyRepo.GetMPlayers",
		slog.Group("reading all players from db for the lobby",
			slog.String("lobby_id", gl.ID)))

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

// DeletePlayer deletes a player from the db
func (l *LobbyRepo) DeletePlayer(lobby_id string, username string) error {
	l.logger.Info("lobbyRepo.DeletePlayer",
		slog.Group("deleting player from db",
			slog.String("lobby_id", lobby_id),
			slog.String("username", username)))

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

// lobbyKey helper that returns a string used to associate the lobby in redis
func lobbyKey(lobby_id string) string {
	return fmt.Sprintf("lobby_id-%v.gamestate", lobby_id)
}

// playerKey helper that returns a string used to associate the player in redis
func playerKey(lobby_id string, u string) string {
	return fmt.Sprintf("lobby_id-%v|username-%v.playerstate", lobby_id, u)
}
