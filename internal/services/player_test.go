package services

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewPlayerService(t *testing.T) {
	ts := NewPlayerService()
	if ts == nil {
		t.Error("Expected player service to not be nil")
	}
}

func TestAddPlayer(t *testing.T) {
	ps := NewPlayerService()

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}

	err := ps.AddPlayer(player)
	if err != nil {
		t.Error("Expected player to be added without any issues")
	}

	players := len(ps.GetPlayers())

	if players != 1 {
		t.Error("Expected players length to be 1 after adding a single player")
	}

}

func TestAddPlayerTwice(t *testing.T) {
	ps := NewPlayerService()

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}
	ps.AddPlayer(player)
	err := ps.AddPlayer(player)
	if err == nil {
		t.Error("Expected an error after adding the same player twice")
	}
}

func TestRemovePlayer(t *testing.T) {
	ps := NewPlayerService()

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}

	ps.AddPlayer(player)

	err := ps.RemovePlayer(player.ID)
	if err != nil {
		t.Error("Expected player to be removed without any errors")
	}

	players := len(ps.GetPlayers())

	if players != 0 {
		t.Error("Expected the number of players to be 0")
	}

}

func TestRemovePlayerTwice(t *testing.T) {
	ps := NewPlayerService()

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}

	ps.AddPlayer(player)

	ps.RemovePlayer(player.ID)

	err := ps.RemovePlayer(player.ID)
	if err == nil {
		t.Error("Expected error after removing the same player twice")
	}

}

func TestGetPlayer(t *testing.T) {
	ps := NewPlayerService()

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}

	ps.AddPlayer(player)

	p, err := ps.GetPlayer(player.ID)
	if err != nil {
		t.Error("Expected no error when tring to get a player that exists")
	}

	if p != player {
		t.Error("Expected the player recived to be that same player that we created")
	}

}

func TestGetPlayerThatDoesNotExist(t *testing.T) {
	ps := NewPlayerService()

	badId := uuid.New()

	_, err := ps.GetPlayer(badId)
	if err == nil {
		t.Error("Expected an error when trying to get a player that does not exist")
	}
}

func TestPlayerAddCardTooHand(t *testing.T) {
	ps := NewPlayerService()

	card := Card{
		Suit: "Spade",
		Type: "Ace",
	}

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}

	ps.AddPlayer(player)

	ps.PlayerAddCardTooHand(player, card)

	player, _ = ps.GetPlayer(player.ID)

    if player.Hand[0] != card {
        t.Error("Expected this card to equal the one we just added")
    }

}
