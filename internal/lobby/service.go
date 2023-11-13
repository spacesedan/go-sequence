package lobby

import (
	"fmt"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/spacesedan/go-sequence/internal"
	"github.com/spacesedan/go-sequence/internal/db"
)

// ill handle changes to the lobby state here

type LobbyService interface {
	NewPlayer(*internal.Player) (*internal.Player, error)
	ShortenPlayerExpiration(string)
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

func (s *lobbyService) NewPlayer(p *internal.Player) (*internal.Player, error) {
	s.logger.Info("lobbyService.NewPlayer",
		fmt.Sprintf("player: %s joined: %s", p.Username, p.LobbyId), "OK")

	err := s.repo.SetPlayer(s.lobby.ID, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (s *lobbyService) ShortenPlayerExpiration(username string) {
	// in order to prevent having tons of unused data store in the db
	// i could shorten the expiration of unregistered users to 30 secs
	// if a user reconnects in that time span it the expiration time goes back
	// to the regular expiration time
	s.repo.Expire(s.lobby.ID, username)

}
