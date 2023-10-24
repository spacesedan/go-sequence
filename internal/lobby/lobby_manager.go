package lobby

import (
	"log/slog"
	"math/rand"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type WsResponse struct {
	Action         string        `json:"action"`
	Message        string        `json:"message"`
	SkipSender     bool          `json:"-"`
	PayloadSession *WsConnection `json:"-"`
	ConnectedUsers []string      `json:"-"`
}

type WsPayload struct {
	Headers       map[string]string `json:"HEADERS"`
	Action        string            `json:"action"`
	Enabled       bool              `json:"enabled"`
	Message       string            `json:"message"`
	SenderSession *WsConnection
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
	WsPayloadChan  chan WsPayload
	Broadcast      chan WsResponse
	RegisterChan   chan *WsConnection
	UnregisterChan chan *WsConnection
}

func NewLobbyManager(l *slog.Logger) *LobbyManager {

	devSettings := Settings{
		NumOfPlayers: "2",
		MaxHandSize:  "7",
	}

	lm := &LobbyManager{
		Lobbies:        make(map[string]*GameLobby),
		WsPayloadChan:  make(chan WsPayload),
		RegisterChan:   make(chan *WsConnection),
		UnregisterChan: make(chan *WsConnection),
		Broadcast:      make(chan WsResponse),
		Sessions:       make(map[*WsConnection]struct{}),
		logger:         l,
	}

	lm.NewGameLobby(devSettings, "ASDA")
	lm.NewGameLobby(devSettings, "JKLK")

	return lm
}

func generateUniqueLobbyId() string {
	result := make([]byte, 4)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func (m *LobbyManager) Run() {
	defer func() {
		for _, lobby := range m.Lobbies {
			close(lobby.PayloadChan)
			close(lobby.RegisterChan)
			close(lobby.UnregisterChan)
		}
		close(m.RegisterChan)
		close(m.Broadcast)
		close(m.UnregisterChan)
		close(m.WsPayloadChan)
	}()
	for {
		select {
		case _ = <-m.RegisterChan:
		case _ = <-m.UnregisterChan:
		case _ = <-m.WsPayloadChan:
		}
	}
}

func (m *LobbyManager) LobbyExists(lobbyId string) bool {
	_, ok := m.Lobbies[lobbyId]
	return ok
}
