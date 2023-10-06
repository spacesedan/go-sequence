package lobby

import (
	"math/rand"

	"github.com/google/uuid"
	"github.com/spacesedan/go-sequence/internal/game"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

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
    result := make([]byte, 4)
    for i := range result {
        result[i] = charset[rand.Intn(len(charset))]
    }
    return string(result)
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
