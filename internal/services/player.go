package services

import (
	"errors"

	"github.com/google/uuid"
)

// Player contains information for a single player
type Player struct {
	Hand       []Card
	CellsTaken map[string]string
	Color      string
	ID         uuid.UUID
	Name       string
}

type Players map[uuid.UUID]*Player

type PlayerService interface {
	AddPlayer(*Player) error
	RemovePlayer(uuid.UUID) error
	GetPlayer(uuid.UUID) (*Player, error)
	GetPlayers() Players

	PlayerPlayCardFromHand(*Player, int) (Card, error)
	PlayerAddCardTooHand(*Player, Card)
}

type playerService struct {
	Players
}

// NewPlayerService create a player service
func NewPlayerService() PlayerService {
	return &playerService{
		Players: make(Players),
	}
}

// AddPlayer add a player to the player list
func (p *playerService) AddPlayer(player *Player) error {
	if _, ok := p.Players[player.ID]; ok {
		return WrapErrorf(errors.New("Error: Player already exists"),
			ErrorCodeInvalidArgument,
			"playerService.AddPlayer",
		)
	}
	player.CellsTaken = make(map[string]string)
	p.Players[player.ID] = player

	return nil
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

func (p *playerService) GetPlayers() Players {
	return p.Players
}

// Get a player form the player list using their id
func (p *playerService) GetPlayer(playerId uuid.UUID) (*Player, error) {
	// check to see if the player exists
	player, ok := p.Players[playerId]
	if !ok {
		// If not found return an error
		return nil, WrapErrorf(errors.New("No player found"), ErrorCodeNotFound, "playerService.RemovePlayer")
	}

	return player, nil
}

// PlayerPlayCardFromHand player plays a card from their hand using a card index,
// it returns the card a the players played or an error
func (p *playerService) PlayerPlayCardFromHand(player *Player, cardIndex int) (Card, error) {
	// check to see if the card played is in the players hand
	if cardIndex > len(player.Hand) {
		return Card{}, WrapErrorf(
			errors.New("Illegal move; cannot play card that is not in your hand"),
			ErrorCodeIllegalMove,
			"playerService.PlayerPlayCardFromHand")

	}

	// newHand holds the value of the player hand minus the card that was played
	var newHand []Card

	// cardPlayed is the card a player want to play
	cardPlayed := player.Hand[cardIndex]

	// Update the player hand
	for i := 0; i < len(player.Hand); i++ {
		// we don't want to add this card back to the player hand so we ignore it
		// in the loop
		if cardPlayed == player.Hand[i] {
			continue
		}

		newHand = append(newHand, player.Hand...)
		// update the player hand with the new hand
		player.Hand = newHand
	}

	return cardPlayed, nil
}

// PlayerAddCardTooHand adds a card to the players hand
func (p *playerService) PlayerAddCardTooHand(player *Player, card Card) {
	player.Hand = append(player.Hand, card)
}
