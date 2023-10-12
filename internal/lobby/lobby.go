package lobby

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/game"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type GameLobby struct {
	ID       string
	Game     game.GameService
	Settings Settings
	Clients  map[string]WsConnection
}

type Settings struct {
	NumOfPlayers string `json:"num_of_players"`
	MaxHandSize  string `json:"max_hand_size"`
	Teams        bool
}

type LobbyManager struct {
	logger    *slog.Logger
	Lobbies   map[string]*GameLobby
	Clients   map[WsConnection]bool
	WsChan    chan WsPayload
	LobbyChan chan WsPayload
}

func NewLobbyManager(l *slog.Logger) *LobbyManager {
	return &LobbyManager{
		Lobbies: make(map[string]*GameLobby),
		WsChan:  make(chan WsPayload),
		Clients: make(map[WsConnection]bool),
		logger:  l,
	}
}

func (lm *LobbyManager) ListenToWsChannel() {
	// var response WsJsonResponse
	for {
		e := <-lm.WsChan
		switch e.Action {
		case "create_lobby":
			lm.logger.Info("Creating new lobby")

		case "join_lobby":

		case "leave_lobby":
			lm.logger.Info("Leaving lobby")

		default:
		}
	}
}

func (lm *LobbyManager) ListenForWs(conn *WsConnection) {
	fmt.Printf("%#v", len(lm.Clients))
	defer func() {
		// if anything happens to the connection remove the connection and recover
		conn.WriteMessage(websocket.TextMessage, []byte("Closing connection"))
		delete(lm.Clients, *conn)
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

	newLobby := &GameLobby{
		ID:       lobbyId,
		Game:     game.NewGameService(game.BoardCellsJSONPath),
		Clients:  make(map[string]WsConnection),
		Settings: s,
	}

	lm.Lobbies[lobbyId] = newLobby

	fmt.Println("number of lobbies", len(lm.Lobbies))

	return lobbyId
}

func (lm *LobbyManager) JoinLobby(lobbyId, username string) error {
    fmt.Println(username)
	if _, ok := lm.Lobbies[lobbyId]; !ok {
		return errors.New("could not join; lobby not found")
	}
	return nil
}
