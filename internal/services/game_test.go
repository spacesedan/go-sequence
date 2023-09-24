package services

import (
	"fmt"
	"testing"
)

func TestNewGameCreated(t *testing.T) {
	gameOptions := GameOptions{
		NumberOfPlayers: 2,
		MaxHandSize:     7,
	}
	newGame := NewGame(gameOptions)

	if newGame == nil {
		t.Errorf("Failed to create a new game")
	}
}

func TestAddingPlayer(t *testing.T) {
	testCases := []struct {
		GameOptions GameOptions
		Players     Players
	}{
		{
			GameOptions{
				MaxHandSize:     7,
				NumberOfPlayers: 2,
			},
			make(Players),
		},
		{
			GameOptions{
				MaxHandSize:     7,
				NumberOfPlayers: 3,
			},
			make(Players),
		},
	}

	for _, tc := range testCases {
		newGame := NewGame(tc.GameOptions)
		for i := 0; i < tc.GameOptions.NumberOfPlayers; i++ {
			newGame.AddPlayer(&Player{
				Name:  fmt.Sprintf("Player %v", i),
				Color: fmt.Sprintf("%v", i),
			})
		}

		numOfPlayers := len(newGame.GetPlayers())

		if numOfPlayers != tc.GameOptions.NumberOfPlayers {
			t.Errorf("Expected number of players: %v, but got %v",
				tc.GameOptions.NumberOfPlayers,
				numOfPlayers)
		}
	}
}

func TestAddingTooManyPlayers(t *testing.T) {
	gO := GameOptions{
		NumberOfPlayers: 1,
	}

	ng := NewGame(gO)

	for i := 0; i < gO.NumberOfPlayers+1; i++ {
		ng.AddPlayer(&Player{
			Name:  fmt.Sprintf("Player %v", i),
			Color: fmt.Sprintf("%v", i),
		})
	}

	numOfPlayers := len(ng.GetPlayers())

	if numOfPlayers != gO.NumberOfPlayers {
		t.Errorf("Expected number of players to equal %v, but got %v",
			gO.NumberOfPlayers,
			numOfPlayers)
	}

}

func TestRemovingPlayers(t *testing.T) {
	gameOptions := GameOptions{
		NumberOfPlayers: 1,
	}

	ng := NewGame(gameOptions)

	playerID := ng.AddPlayer(&Player{
		Name:  "Player 1",
		Color: "green",
	})

	ng.RemovePlayer(playerID)

	players := ng.GetPlayers()

	if len(players) != 0 {
		t.Errorf("Expected number of players to be 0; but got %v", len(players))
	}
}

func TesetDealingCardsTwice(t *testing.T) {
	gameOptions := GameOptions{
		NumberOfPlayers: 1,
		MaxHandSize:     7,
	}

	ng := NewGame(gameOptions)

	playerId := ng.AddPlayer(&Player{
		Name:  "Player 1",
		Color: "Green",
	})

	ng.DealCards()
	ng.DealCards()

	playerHandLength := len(ng.GetPlayers()[playerId].Hand)

	if playerHandLength > gameOptions.MaxHandSize {
		t.Errorf("Expected player hand size to equal %v; but got %v",
			gameOptions.MaxHandSize,
			playerHandLength)
	}

}



