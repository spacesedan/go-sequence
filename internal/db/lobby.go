package db

import (
	"encoding/json"
	"log/slog"

	goredis "github.com/go-redis/redis/v8"
	"github.com/gomodule/redigo/redis"
	"github.com/spacesedan/go-sequence/internal"
)

type LobbyRepo interface {
	SetLobby(lobby *internal.Lobby) error
	GetLobby(lobbyID string) (*internal.Lobby, error)
	DeleteLobby(lobbyID string) error

	GetPlayer(lobbyID string, username string) (*internal.Player, error)
	SetPlayer(lobbyID string, player *internal.Player) error
	DeletePlayer(lobby_id string, username string) error
}

// LobbyRepo responsible for interfacing with the data stored in the cache
type lobbyRepo struct {
	redisClient *goredis.Client
	logger      *slog.Logger
}

// NewLobbyRepo creates a LobbyRepo instance to interact with the goredis server
func NewLobbyRepo(r *goredis.Client, l *slog.Logger) LobbyRepo {
	l.Info("lobby.NewLobbyRepo: created new lobby repo")
	return &lobbyRepo{
		redisClient: r,
		logger:      l,
	}
}

// SetLobby sets and updates the lobby state stored in the db using the lobby
func (l *lobbyRepo) SetLobby(lobby *internal.Lobby) error {
	l.logger.Info("lobbyRepo.SetLobby",
		slog.Group("writing lobby to db"))

	conn := l.redisClient
	rh := NewReJSONHandler(conn)

	_, err := rh.JSONSet(lobbyKey(lobby.ID), &internal.Lobby{
		ID:              lobby.ID,
		CurrentState:    lobby.CurrentState,
		Settings:        lobby.Settings,
		ColorsAvailable: lobby.ColorsAvailable,
		Players:         lobby.Players,
	})

	if err != nil {
		return err
	}

	return nil
}

// GetLobby gets a lobby from teh db using the lobby id
func (l *lobbyRepo) GetLobby(lobby_id string) (*internal.Lobby, error) {
	l.logger.Info("lobbyRepo.GetLobby",
		slog.Group("reading lobby from goredis",
			slog.String("lobby_id", lobby_id)))

	var lobbyState *internal.Lobby

	rh := NewReJSONHandler(l.redisClient)
	lobbyJSON, err := redis.Bytes(rh.rj.JSONGet(lobbyKey(lobby_id), "."))
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
func (l *lobbyRepo) DeleteLobby(lobby_id string) error {
	l.logger.Info("lobbyRepo.DeleteLobby",
		slog.Group("deleting lobby from goredis",
			slog.String("lobby_id", lobby_id)))

	rh := NewReJSONHandler(l.redisClient)

	if _, err := rh.rj.JSONDel(lobbyKey(lobby_id), "."); err != nil {
		return err
	}
	return nil
}

// SetPlayer sets and updates player data in the db
func (l *lobbyRepo) SetPlayer(lobby_id string, p *internal.Player) error {
	l.logger.Info("lobbyRepo.SetPlayer",
		slog.Group("writing player to db",
			slog.String("player", p.Username)))

	rh := NewReJSONHandler(l.redisClient)

	if _, err := rh.JSONSet(playerKey(lobby_id, p.Username), &internal.Player{
		Username: p.Username,
		LobbyId:  lobby_id,
		Color:    p.Color,
		Ready:    p.Ready,
	}); err != nil {
		return err
	}

	return nil
}

// GetPlayer gets a player from the db using the lobby id and player username
func (l *lobbyRepo) GetPlayer(lobby_id string, username string) (*internal.Player, error) {
	l.logger.Info("lobbyRepo.GetPlayer",
		slog.Group("reading player from db",
			slog.String("lobby_id", lobby_id),
			slog.String("username", username)))

	var ps *internal.Player

	rh := NewReJSONHandler(l.redisClient)

	pj, err := redis.Bytes(rh.rj.JSONGet(playerKey(lobby_id, username), "."))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(pj, &ps)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

// DeletePlayer deletes a player from the db
func (l *lobbyRepo) DeletePlayer(lobby_id string, username string) error {
	l.logger.Info("lobbyRepo.DeletePlayer",
		slog.Group("deleting player from db",
			slog.String("lobby_id", lobby_id),
			slog.String("username", username)))

	rh := NewReJSONHandler(l.redisClient)

	_, err := rh.rj.JSONDel(playerKey(lobby_id, username), ".")
	if err != nil {
		return err
	}

	return nil

}
