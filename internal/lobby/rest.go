package lobby

import (
	"errors"

	"github.com/spacesedan/go-sequence/internal/game"
)

func (lm *LobbyManager) CreateLobby(s Settings) string {
	lobbyId := generateUniqueLobbyId()

	newLobby := &GameLobby{
		ID:       lobbyId,
		Game:     game.NewGameService(game.BoardCellsJSONPath),
		Clients:  make(map[string]WsConnection),
		GameChan: make(chan WsPayload),
		Settings: s,
	}

	lm.Lobbies[lobbyId] = newLobby

	return lobbyId
}

func (lm *LobbyManager) JoinLobby(lobbyId, username string) (bool, error) {
	if _, ok := lm.Lobbies[lobbyId]; !ok {
		return false, errors.New("could not join; lobby not found")
	}
	return true, nil
}
