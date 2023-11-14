package lobby

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spacesedan/go-sequence/internal"
	"github.com/spacesedan/go-sequence/internal/db"
	"github.com/spacesedan/go-sequence/internal/game"
)

type LobbyChannel uint

const (
	UnknownChannel LobbyChannel = iota
	RegisterChannel
	DeregisterChannel
	PayloadChannel
	StateChannel
)

func (c LobbyChannel) String(lobby_id string) string {
	switch c {
	case UnknownChannel:
		return "unknown"
	case RegisterChannel:
		return fmt.Sprintf("lobby.%v.registerChannel", lobby_id)
	case DeregisterChannel:
		return fmt.Sprintf("lobby.%v.unregisterChannel", lobby_id)
	case PayloadChannel:
		return fmt.Sprintf("lobby.%v.payloadChannel", lobby_id)
	case StateChannel:
		return fmt.Sprintf("lobby.%v.stateChannel", lobby_id)
	default:
		return ""
	}
}

type Lobby struct {
	// game data
	ID              string
	Game            game.GameService
	ColorsAvailable map[string]bool
	Settings        internal.Settings
	Players         map[string]*internal.Player
	CurrentState    internal.CurrentState

	handler      LobbyHandler
	lobbyRepo    db.LobbyRepo
	lobbyManager *LobbyManager
	logger       *slog.Logger
	redisClient  *redis.Client

	errorChan chan error
}

// Create a new lobby
func (m *LobbyManager) NewLobby(settings internal.Settings, id ...string) string {
	var lobbyId string
	m.lobbiesMu.Lock()
	defer m.lobbiesMu.Unlock()

	colors := make(map[string]bool, 3)

	if len(id) != 0 {
		lobbyId = id[0]
	} else {
		lobbyId = generateUniqueLobbyId()
	}

	m.logger.Info("lobbyManager.NewLobby",
		slog.Group("Creating new lobby",
			slog.String("lobbyId", lobbyId)))

	colors["red"] = true
	colors["blue"] = true
	colors["green"] = true

	l := &Lobby{
		ID:              lobbyId,
		Game:            game.NewGameService(),
		Settings:        settings,
		CurrentState:    internal.InLobby,
		ColorsAvailable: colors,
		Players:         make(map[string]*internal.Player),
		lobbyManager:    m,
		logger:          m.logger,
		redisClient:     m.redisClient,
		lobbyRepo:       db.NewLobbyRepo(m.redisClient, m.logger),
		errorChan:       make(chan error, 1),
	}

	l.lobbyRepo.SetLobby(&internal.Lobby{
		ID:              l.ID,
		Players:         l.Players,
		Settings:        l.Settings,
		ColorsAvailable: l.ColorsAvailable,
		CurrentState:    l.CurrentState,
	})

	l.handler = NewLobbyHandler(m.redisClient, l, l.logger)

	m.Lobbies[lobbyId] = l

	l.redisClient.Publish(context.Background(), "lobby_manager.create", fmt.Sprintf("created a new lobby id: %v ", l.ID))

	go l.Subscribe()

	return lobbyId
}

func (m *LobbyManager) CloseLobby(id string) {
	m.lobbiesMu.Lock()
	defer m.lobbiesMu.Unlock()

	lobby := m.Lobbies[id]

	lobby.lobbyRepo.DeleteLobby(lobby.ID)

	m.logger.Info("lobbyManager.CloseLobby",
		slog.Group("Closing Lobby",
			slog.String("lobby_id", id)))

	delete(m.Lobbies, id)
}

// Subscribe listens to the lobby payload channel and once it recieves a payload it
// sends a response to the appropriate channel
func (l *Lobby) Subscribe() {
	var payload WsPayload
	chanKey := fmt.Sprintf("lobby.%v.*", l.ID)
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(time.Minute)
	sub := l.redisClient.PSubscribe(ctx, chanKey)

	defer func() {
		sub.Close()

		ticker.Stop()
		cancel()
	}()

	ch := sub.Channel()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if err := payload.Unmarshal(msg.Payload); err != nil {
				l.logger.Error("lobby.Subscribe",
					slog.Group("failed to unmarshal payload",
						slog.Any("reason", err)))
				return
			}
			switch msg.Channel {
			case RegisterChannel.String(l.ID):
				l.handler.RegisterPlayer(payload)
			case StateChannel.String(l.ID):
				l.handler.ChangeState()
			case DeregisterChannel.String(l.ID):
				l.handler.DeregisterPlayer(payload)
			case PayloadChannel.String(l.ID):
				l.handler.DispatchAction(payload)
			}

		case err := <-l.errorChan:
			l.logger.Error("lobby.Subscribe",
				slog.Group("something went wrong",
					slog.Any("reason", err)))

			return
		case <-ticker.C:
			err := sub.Ping(ctx)
			if err != nil {
				return
			}
			if l.handler.EmptyLobby() {
				return
			}
		}

	}
}

func (l *Lobby) HasPlayer(username string) bool {
	if _, ok := l.Players[username]; ok {
		return true
	}
	return false
}

func toLobbyState(l *Lobby) *internal.Lobby {
	return &internal.Lobby{
		ID:              l.ID,
		Players:         l.Players,
		ColorsAvailable: l.ColorsAvailable,
		Settings:        l.Settings,
		CurrentState:    l.CurrentState,
	}
}
