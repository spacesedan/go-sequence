package lobby

import (
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/game"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type WsConnection struct {
	LobbyManager *LobbyManager
	Conn         *websocket.Conn
	Username     string
	LobbyID      string
	Send         chan WsPayload
}

type WsJsonResponse struct {
	Headers        map[string]interface{} `json:"HEADERS"`
	Action         string                 `json:"action"`
	Message        string                 `json:"message"`
	LobbyID        string
	MessageType    string       `json:"message_type"`
	SkipSender     bool         `json:"-"`
	CurrentSession WsConnection `json:"-"`
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
	Session  WsConnection
}

type GameLobby struct {
	ID       string
	Game     game.GameService
	Settings Settings
	Sessions map[WsConnection]struct{}
	GameChan chan WsPayload
}

type Settings struct {
	NumOfPlayers string `json:"num_of_players"`
	MaxHandSize  string `json:"max_hand_size"`
	Teams        bool
}

type LobbyManager struct {
	logger         *slog.Logger
	Lobbies        map[string]*GameLobby
	Sessions       map[*WsConnection]struct{}
	WsChan         chan WsPayload
	Broadcast      chan []byte
	RegisterChan   chan *WsConnection
	UnregisterChan chan *WsConnection
}

func NewLobbyManager(l *slog.Logger) *LobbyManager {
	lobbies := make(map[string]*GameLobby)

	// for debugging
	lobbies["ASDA"] = &GameLobby{
		ID:       "ASDA",
		GameChan: make(chan WsPayload),
		Game:     game.NewGameService(game.BoardCellsJSONPath),
		Sessions: make(map[WsConnection]struct{}),
		Settings: Settings{
			NumOfPlayers: "2",
			MaxHandSize:  "7",
		},
	}
	lobbies["JKLK"] = &GameLobby{
		ID:       "JKLK",
		GameChan: make(chan WsPayload),
		Game:     game.NewGameService(game.BoardCellsJSONPath),
		Sessions: make(map[WsConnection]struct{}),
		Settings: Settings{
			NumOfPlayers: "2",
			MaxHandSize:  "7",
		},
	}

	return &LobbyManager{
		Lobbies:        lobbies,
		WsChan:         make(chan WsPayload),
		RegisterChan:   make(chan *WsConnection),
		UnregisterChan: make(chan *WsConnection),
		Broadcast:      make(chan []byte),
		Sessions:       make(map[*WsConnection]struct{}),
		logger:         l,
	}
}

func generateUniqueLobbyId() string {
	result := make([]byte, 4)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func (l *LobbyManager) Run() {
	for {
		select {
		case session := <-l.RegisterChan:
            l.logger.Info("[REGISTERING]", slog.String("user", session.Username))
			l.Sessions[session] = struct{}{}
		case session := <-l.UnregisterChan:
            l.logger.Info("[UNREGISTERING]", slog.String("user", session.Username))
			if _, ok := l.Sessions[session]; ok {
				delete(l.Sessions, session)
			}

			for sess := range l.Sessions {
				sess.Send <- WsPayload{
					Action:   "left",
					Username: session.Username,
					LobbyID:  sess.LobbyID,
				}
			}
		case message := <-l.WsChan:
			for session := range l.Sessions {
				select {
				case session.Send <- message:
				default:
					close(session.Send)
					delete(l.Sessions, session)
				}
			}

		}
	}
}
