package lobby

import (
	"github.com/spacesedan/go-sequence/internal/game"
)

func (lm *LobbyManager) CreateLobby(s Settings) string {
	lobbyId := generateUniqueLobbyId()

	newLobby := &GameLobby{
		ID:       lobbyId,
		Game:     game.NewGameService(game.BoardCellsJSONPath),
		GameChan: make(chan WsPayload),
		Clients:  make(map[string]WsConnection),
		Settings: s,
	}

	lm.Lobbies[lobbyId] = newLobby

	return lobbyId
}

func (lm *LobbyManager) JoinLobby(lobbyId, username string) bool {
	_, ok := lm.Lobbies[lobbyId]
	if !ok {
		return false
	}

	return true
}
