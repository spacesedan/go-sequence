package db

import (
	"encoding/json"
	"log/slog"

	goredis "github.com/go-redis/redis/v8"
	"github.com/gomodule/redigo/redis"
	"github.com/spacesedan/go-sequence/internal/game"
)

type LobbyRepo interface {
	SetLobby(lobby *LobbyState) error
	GetLobby(lobbyID string) (*LobbyState, error)
	DeleteLobby(lobbyID string) error

    SetPlayer(lobbyID string, player *PlayerState) error
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

type LobbyState struct {
	ID              string
	CurrentState    CurrentState
	Players         map[string]struct{}
	ColorsAvailable map[string]bool
	Settings        game.Settings
}

// SetLobby sets and updates the lobby state stored in the db using the lobby
func (l *lobbyRepo) SetLobby(lobby *LobbyState) error {
	l.logger.Info("lobbyRepo.SetLobby",
		slog.Group("writing lobby to db"))

	conn := l.redisClient
	rh := NewReJSONHandler(conn)

	_, err := rh.JSONSet(lobbyKey(lobby.ID), ".", &LobbyState{
		CurrentState:    InLobby,
		Settings:        lobby.Settings,
		ColorsAvailable: lobby.ColorsAvailable,
		Players:         lobby.Players,
	})

	if err != nil {
		return err
	}

	return nil
}

// SetPlayer sets and updates player data in the db
func (l *lobbyRepo) SetPlayer(lobby_id string, p *PlayerState) error {
	l.logger.Info("lobbyRepo.SetPlayer",
		slog.Group("writing player to db",
			slog.String("player", p.Username)))

	rh := NewReJSONHandler(l.redisClient)

	if _, err := rh.JSONSet(playerKey(lobby_id, p.Username), ".", &PlayerState{
		Username: p.Username,
		LobbyId:  lobby_id,
		Color:    p.Color,
		Ready:    p.Ready,
	}); err != nil {
		return err
	}

	return nil
}

// GetLobby gets a lobby from teh db using the lobby id
func (l *lobbyRepo) GetLobby(lobby_id string) (*LobbyState, error) {
	l.logger.Info("lobbyRepo.GetLobby",
		slog.Group("reading lobby from goredis",
			slog.String("lobby_id", lobby_id)))

	var lobbyState *LobbyState

	rh := NewReJSONHandler(l.redisClient)
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
func (l *lobbyRepo) DeleteLobby(lobby_id string) error {
	l.logger.Info("lobbyRepo.DeleteLobby",
		slog.Group("deleting lobby from goredis",
			slog.String("lobby_id", lobby_id)))

	rh := NewReJSONHandler(l.redisClient)

	if _, err := rh.JSONDel(lobbyKey(lobby_id), "."); err != nil {
		return err
	}
	return nil
}

// DeletePlayer deletes a player from the db
func (l *lobbyRepo) DeletePlayer(lobby_id string, username string) error {
	l.logger.Info("lobbyRepo.DeletePlayer",
		slog.Group("deleting player from db",
			slog.String("lobby_id", lobby_id),
			slog.String("username", username)))

	rh := NewReJSONHandler(l.redisClient)

	_, err := rh.JSONDel(playerKey(lobby_id, username), ".")
	if err != nil {
		return err
	}

	return nil

}

// Current state ... are the players in the lobby still choosing thier colors,
// or are they in the game
type CurrentState uint

const (
	Unknown CurrentState = iota
	InLobby
	InGame
)

// String get a stringified version of the current game state
func (c CurrentState) String() string {
	switch c {
	case InLobby:
		return "lobby"
	case InGame:
		return "game"
	default:
		return "unknown"

	}
}
