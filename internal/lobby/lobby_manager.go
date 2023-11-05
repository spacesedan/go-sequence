package lobby

import (
	"log/slog"
	"math/rand"
	"sync"

	"github.com/go-redis/redis/v8"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type WsResponse struct {
	Action         ResponseEvent `json:"action"`
	Message        string        `json:"message"`
	SkipSender     bool          `json:"-"`
	PayloadSession *WsClient     `json:"-"`
	ConnectedUsers []string      `json:"-"`
}

type WsPayload struct {
	Action        PayloadEvent `json:"action"`
	Message       string       `json:"message"`
	SenderSession *WsClient    `json:"-"`
}

type Settings struct {
	NumOfPlayers int `json:"num_of_players"`
	MaxHandSize  int `json:"max_hand_size"`
	Teams        bool
}

type LobbyManager struct {
	logger      *slog.Logger
	redisClient *redis.Client

	lobbiesMu      sync.Mutex
	Lobbies        map[string]*GameLobby
	Sessions       map[*WsClient]struct{}
	RegisterChan   chan *GameLobby
	UnregisterChan chan *GameLobby
}

func NewLobbyManager(r *redis.Client, l *slog.Logger) *LobbyManager {
	l.Info("NewLobbyManager", slog.String("reason", "starting up lobby manager"))
	devSettings := Settings{
		NumOfPlayers: 2,
		MaxHandSize:  7,
	}

	lm := &LobbyManager{
		logger:      l,
		redisClient: r,

		Lobbies:        make(map[string]*GameLobby),
		RegisterChan:   make(chan *GameLobby),
		UnregisterChan: make(chan *GameLobby),
		Sessions:       make(map[*WsClient]struct{}),
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
			delete(m.Lobbies, lobby.ID)
		}
	}
}

func (m *LobbyManager) LobbyExists(lobbyId string) (*GameLobby, bool) {
	l, ok := m.Lobbies[lobbyId]
	return l, ok
}
