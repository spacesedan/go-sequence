package lobby

import (
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
	Clients  map[WsConnection]bool
}

type Settings struct {
	NumOfPlayers string `json:"num_of_players"`
	MaxHandSize  string `json:"max_hand_size"`
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
		Lobbies: make(map[string]GameLobby),
		WsChan:  make(chan WsPayload),
		Clients: make(map[WsConnection]string),
		logger:  l,
	}
}

func (lm *LobbyManager) ListenToWsChannel() {
	var response WsJsonResponse
	for {
		e := <-lm.WsChan
		fmt.Printf("%v", e)
		switch e.Action {
		case "create_lobby":
			response.Action = "new_lobby"
			lm.logger.Info("Creating new lobby")
			lobbyId := lm.CreateLobby(e.Settings, e.Conn)
			lm.logger.Info("Lobby created", slog.String("lobby id", lobbyId))

			for client := range lm.Clients {
				if e.Conn == client {
					fmt.Println("true")
					response.Message = fmt.Sprintf(`<a href="/lobby/%s" id="lobby_link">Go to lobby</a>`, lobbyId)
					err := client.WriteMessage(websocket.TextMessage, []byte(response.Message))
					if err != nil {
						fmt.Println(err)
					}
				}

			}

			fmt.Println("Number of active lobbies", len(lm.Lobbies))
		default:
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
			fmt.Println(err)
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

func (lm *LobbyManager) CreateLobby(s Settings, conn WsConnection) string {
	lobbyId := generateUniqueLobbyId()

	newLobby := GameLobby{
		ID:       lobbyId,
		Game:     game.NewGameService(game.BoardCellsJSONPath),
		Clients:  make(map[WsConnection]bool),
		Settings: s,
	}

	newLobby.Clients[conn] = true
	lm.Lobbies[lobbyId] = newLobby

	return lobbyId
}
