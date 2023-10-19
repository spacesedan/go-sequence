package lobby

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/game"
	"github.com/spacesedan/go-sequence/internal/partials"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type WsResponse struct {
	Action         string `json:"action"`
	Message        string `json:"message"`
	LobbyID        string `json:"lobby_id"`
	Username       string
	SkipSender     bool          `json:"-"`
	CurrentSession *WsConnection `json:"-"`
	ConnectedUsers []string      `json:"-"`
}

type WsPayload struct {
	Headers  map[string]string `json:"HEADERS"`
	Action   string            `json:"action"`
	Settings Settings          `json:"settings"`
	ID       string            `json:"id"`
	LobbyID  string            `json:"lobby_id"`
	Username string            `json:"username"`
	Message  string            `json:"message"`
	Session  *WsConnection
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
	Broadcast      chan WsResponse
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

	return &LobbyManager{
		Lobbies:        lobbies,
		WsChan:         make(chan WsPayload),
		RegisterChan:   make(chan *WsConnection),
		UnregisterChan: make(chan *WsConnection),
		Broadcast:      make(chan WsResponse),
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
		case payload := <-l.WsChan:
			for session := range l.Sessions {
				select {
				case session.Send <- payload:
				default:
					close(session.Send)
					delete(l.Sessions, session)
				}
			}
		case response := <-l.Broadcast:
			switch response.Action {
			case "new_message":
				l.broadcastChatMessage(response)
			case "joined", "left":
				fmt.Println()
				fmt.Println("number of connections", len(l.Sessions))
				fmt.Println()
				l.broadcastUserStatus(response)
			}

		}
	}
}

func (l *LobbyManager) broadcastUserStatus(response WsResponse) {
	var b bytes.Buffer
	defer b.Reset()

	for session := range l.Sessions {
		fmt.Println(response.Username)

		err := partials.PlayerStatus(response.Message).Render(context.Background(), &b)
		if err != nil {
		}

		err = session.Conn.WriteMessage(websocket.TextMessage, []byte(b.String()))
		if err != nil {
			_ = session.Conn.Close()
			delete(l.Sessions, session)
		}
	}

}

func (l *LobbyManager) broadcastChatMessage(response WsResponse) {
	_ = bytes.Buffer{}

	for session := range l.Sessions {
		if session == response.CurrentSession {
			fmt.Printf("%v %v %v\n", true, session.Username, time.Now())

		} else {
			fmt.Printf("%v %v %v\n", true, session.Username, time.Now())
		}

	}

}
