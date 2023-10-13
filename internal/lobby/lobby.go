package lobby

import (
	"log/slog"
	"math/rand"

	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/game"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type WsConnection struct {
	*websocket.Conn
}

type WsJsonResponse struct {
	Headers        map[string]interface{} `json:"HEADERS"`
	Action         string                 `json:"action"`
	Message        string                 `json:"message"`
	LobbyID        string
	MessageType    string       `json:"message_type"`
	SkipSender     bool         `json:"-"`
	CurrentConn    WsConnection `json:"-"`
	ConnectedUsers []string     `json:"-"`
}

type WsPayload struct {
	Headers  map[string]string `json:"HEADERS"`
	Action   string            `json:"action"`
	Settings Settings          `json:"settings"`
	ID       string            `json:"id"`
	LobbyID  string            `json:"lobby_id"`
	Username string            `json:"username"`
	Message  string            `json:"message"`
	Conn     WsConnection
}

type GameLobby struct {
	ID       string
	Game     game.GameService
	Settings Settings
	Clients  map[string]WsConnection
	GameChan chan WsPayload
}

type Settings struct {
	NumOfPlayers string `json:"num_of_players"`
	MaxHandSize  string `json:"max_hand_size"`
	Teams        bool
}

type LobbyManager struct {
	logger    *slog.Logger
	Lobbies   map[string]*GameLobby
	Clients   map[string]WsConnection
	WsChan    chan WsPayload
	LobbyChan chan WsPayload
}

func NewLobbyManager(l *slog.Logger) *LobbyManager {
	return &LobbyManager{
		Lobbies: make(map[string]*GameLobby),
		WsChan:  make(chan WsPayload),
		Clients: make(map[string]WsConnection),
		logger:  l,
	}
}

func generateUniqueLobbyId() string {
	result := make([]byte, 4)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
