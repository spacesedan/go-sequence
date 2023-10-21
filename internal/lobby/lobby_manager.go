package lobby

import (
	"log/slog"
	"math/rand"
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
	Enabled  bool              `json:"enabled"`
	Username string            `json:"username"`
	Message  string            `json:"message"`
	Session  *WsConnection
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

	devSettings := Settings{
		NumOfPlayers: "2",
		MaxHandSize:  "7",
	}

	lm := &LobbyManager{
		Lobbies:        make(map[string]*GameLobby),
		WsChan:         make(chan WsPayload),
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
			close(lobby.GameChan)
		}
		close(m.RegisterChan)
		close(m.Broadcast)
		close(m.UnregisterChan)
		close(m.WsChan)
	}()
	for {
		select {
		case session := <-m.RegisterChan:
			m.logger.Info("[REGISTERING]", slog.String("user", session.Username))
			// add session to all sessions
			m.Sessions[session] = struct{}{}
			// if the player
			m.Lobbies[session.LobbyID].Players[session.Username] = session
		case session := <-m.UnregisterChan:
			// delete user from all sessions
			m.logger.Info("[UNREGISTERING]", slog.String("user", session.Username))
			if _, ok := m.Sessions[session]; ok {
				delete(m.Sessions, session)
				close(session.Send)
			}
			// delete user from lobby session
			if _, ok := m.Lobbies[session.LobbyID].Sessions[session]; ok {
				delete(m.Lobbies[session.LobbyID].Players, session.Username)
			}
		case payload := <-m.WsChan:
			// only send to lobby chan, the lobby chan will send only to sessions
			// associated with that lobby
			if ok := m.LobbyExists(payload.LobbyID); ok {
				m.Lobbies[payload.LobbyID].GameChan <- payload
			}
		}
	}
}

func (m *LobbyManager) LobbyExists(lobbyId string) bool {
	_, ok := m.Lobbies[lobbyId]
	return ok
}
