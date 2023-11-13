package lobby

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spacesedan/go-sequence/internal"
	"github.com/spacesedan/go-sequence/internal/db"
)


type LobbyService interface {
	NewPlayer(string) (*internal.Player, error)

	SetLobby(*internal.Lobby) error

	SetPlayer(*internal.Player) error
	GetPlayer(string) (*internal.Player, error)
	SetExpiration(string, time.Duration)

	GetPlayerNames() []string
}

type lobbyService struct {
	lobby  *Lobby
	repo   db.LobbyRepo
	logger *slog.Logger
}

func NewLobbyService(r *redis.Client, l *Lobby, logger *slog.Logger) LobbyService {
	return &lobbyService{
		lobby:  l,
		repo:   db.NewLobbyRepo(r, logger),
		logger: logger,
	}
}

func (s *lobbyService) SetLobby(l *internal.Lobby) error {
    return s.repo.SetLobby(l)
}

func (s *lobbyService) NewPlayer(username string) (*internal.Player, error) {
	s.logger.Info("lobbyService.NewPlayer",
		fmt.Sprintf("player: %s joined: %s", username, s.lobby.ID), "OK")

	err := s.repo.SetPlayer(s.lobby.ID, &internal.Player{
		Username: username,
		LobbyId:  s.lobby.ID,
	})
	if err != nil {
		return nil, err
	}

	ps, err := s.repo.GetPlayer(s.lobby.ID, username)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (s *lobbyService) SetPlayer(state *internal.Player) error {
	return s.repo.SetPlayer(s.lobby.ID, state)
}

func (s *lobbyService) GetPlayer(username string) (*internal.Player, error) {
	s.logger.Info("lobbyService.GetPlayer")
	ps, err := s.repo.GetPlayer(s.lobby.ID, username)
	if err != nil {
		return nil, err
	}

	s.SetExpiration(username, time.Duration(30*time.Minute))
	return ps, nil
}

func (s *lobbyService) GetPlayerNames() []string {
	var playerNames []string
	for p := range s.lobby.Players {
		playerNames = append(playerNames, p)
	}
	return playerNames
}

func (s *lobbyService) SetExpiration(username string, dur time.Duration) {
	// in order to prevent having tons of unused data store in the db
	// i could shorten the expiration of unregistered users to 30 secs
	// if a user reconnects in that time span it the expiration time goes back
	// to the regular expiration time
	s.repo.Expire(s.lobby.ID, username, dur)

}
