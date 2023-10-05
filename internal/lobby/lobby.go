package lobby

import (
	"github.com/google/uuid"
	"github.com/spacesedan/go-sequence/internal/game"
)

type GameLobby struct {
	ID       string
	Game     game.GameService
	Settings Settings
}

type Settings struct {
	NumOfPlayers int
	MaxHandSize  int
	Teams        bool
}

type LobbyManager struct {
	Lobbies map[string]GameLobby
}

func NewLobbyManager() *LobbyManager {
	return &LobbyManager{
		Lobbies: map[string]GameLobby{},
	}
}

func generateUniqueLobbyId() string {
    newId := uuid.New()
    return newId.String()
}

func (lm *LobbyManager) CreateLobby(s Settings) string {
	lobbyId := generateUniqueLobbyId()

	newLobby := GameLobby{
		ID:       lobbyId,
		Game:     game.NewGameService(game.BoardCellsJSONPath),
		Settings: s,
	}

    lm.Lobbies[lobbyId] = newLobby

    return lobbyId

}
