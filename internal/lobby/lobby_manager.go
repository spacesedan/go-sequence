package lobby

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/spacesedan/go-sequence/internal/game"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type WsResponse struct {
	Action         ResponseEvent `json:"action"`
	Message        string        `json:"message"`
	Sender         string        `json:"sender"`
	SkipSender     bool          `json:"skip_sender"`
	ConnectedUsers []string      `json:"connected_users"`
}

func (r WsResponse) MarshalBinary() ([]byte, error) {
	return json.Marshal(r)
}

func (r *WsResponse) Unmarshal(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type WsPayload struct {
	Action        PayloadEvent `json:"action"`
	Message       string       `json:"message"`
	Username      string       `json:"username"`
	SenderSession *WsClient    `json:"-"`
}

func (p WsPayload) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

func (p *WsPayload) Unmarshal(s string) error {
	return json.Unmarshal([]byte(s), &p)
}


type LobbyManager struct {
	logger      *slog.Logger
	redisClient *redis.Client

	lobbiesMu      sync.Mutex
	Lobbies        map[string]*Lobby
	Sessions       map[*WsClient]struct{}
	RegisterChan   chan *Lobby
	UnregisterChan chan *Lobby
}

func NewLobbyManager(r *redis.Client, l *slog.Logger) *LobbyManager {
	l.Info("NewLobbyManager", slog.String("reason", "starting up lobby manager"))
	devSettings := game.Settings{
		NumOfPlayers: 2,
		MaxHandSize:  7,
	}

	lm := &LobbyManager{
		logger:      l,
		redisClient: r,

		Lobbies:        make(map[string]*Lobby),
		RegisterChan:   make(chan *Lobby),
		UnregisterChan: make(chan *Lobby),
		Sessions:       make(map[*WsClient]struct{}),
	}

	lm.NewLobby(devSettings, "ASDA")
	lm.NewLobby(devSettings, "JKLK")

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
			m.CloseLobby(lobby.ID)
		}
		close(m.RegisterChan)
		close(m.UnregisterChan)
	}()
	for {
		select {
		case lobby := <-m.RegisterChan:
			m.Lobbies[lobby.ID] = lobby
		case lobby := <-m.UnregisterChan:
			m.CloseLobby(lobby.ID)
		}
	}
}

func (m *LobbyManager) LobbyExists(lobbyId string) (*Lobby, bool) {
	l, ok := m.Lobbies[lobbyId]
	return l, ok
}
