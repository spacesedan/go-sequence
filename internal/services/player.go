package services

import (
	"errors"

	"github.com/google/uuid"
)

// Player contains information for a single player
type Player struct {
	Hand  []Card
	Color string
	ID    uuid.UUID
	Name  string
}

type Players map[uuid.UUID]*Player

type PlayerService interface {
    AddPlayer(*Player)
    RemovePlayer(uuid.UUID) error
}

type playerService struct{
    Players
}

// NewPlayerService create a player service
func NewPlayerService() PlayerService {
    return &playerService{
        Players: make(Players),
    }
}

// AddPlayer add a player to the player list
func (p *playerService) AddPlayer(player *Player) {
    p.Players[player.ID] = player
}

// RemovePlayer removes a player from the player list
func (p *playerService) RemovePlayer(playerId uuid.UUID) error {
    // check to see if the player exists
    if _, ok := p.Players[playerId]; !ok {
        // If not found return an error
        return WrapErrorf(errors.New("No player found"), ErrorCodeNotFound, "playerService.RemovePlayer")
    }

    // if the player does exist remove them from the player list
    delete(p.Players, playerId)

    return nil
}

