package lobby

import (
	"fmt"
	"log/slog"
	"math/rand"

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
	logger    *slog.Logger
	Lobbies   map[string]GameLobby
	Clients   map[WsConnection]string
	WsChan    chan WsPayload
	LobbyChan chan WsPayload
}

func NewLobbyManager(l *slog.Logger) *LobbyManager {
	return &LobbyManager{
		Lobbies: map[string]GameLobby{},
		WsChan:  make(chan WsPayload),
        Clients: map[WsConnection]string{},
		logger:  l,
	}
}

func (lm *LobbyManager) ListenToWsChannel() {
	var response WsJsonResponse
	for {
		e := <-lm.WsChan
		switch e.Action {
		default:
			fmt.Println(response)
		}
	}
}

func (lm *LobbyManager) ListenForWs(conn *WsConnection) {
	defer func() {
		if r := recover(); r != nil {
			lm.logger.Error("Error: Attempting to recover", slog.Any("reason", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// ... just ignore it
		} else {
			payload.Conn = *conn
			lm.WsChan <- payload
		}
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
