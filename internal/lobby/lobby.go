package lobby

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spacesedan/go-sequence/internal"
	"github.com/spacesedan/go-sequence/internal/db"
	"github.com/spacesedan/go-sequence/internal/game"
)

type Lobby struct {
	// game data
	ID              string
	Game            game.GameService
	AvailableColors map[string]bool
	Settings        internal.Settings
	Players         map[string]struct{}

	lobbyState   *internal.Lobby
	lobbyRepo    db.LobbyRepo
	lobbyManager *LobbyManager
	logger       *slog.Logger
	redisClient  *redis.Client
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
		AvailableColors: colors,
		Players:         make(map[string]struct{}),

		lobbyManager: m,
		logger:       m.logger,
		redisClient:  m.redisClient,
		lobbyRepo:    db.NewLobbyRepo(m.redisClient, m.logger),
	}

	l.lobbyRepo.SetLobby(&internal.Lobby{
		ID:              l.ID,
		Players:         l.Players,
		Settings:        l.Settings,
		ColorsAvailable: l.AvailableColors,
	})

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
		sub.Unsubscribe(ctx, chanKey)
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
			case fmt.Sprintf("lobby.%v.registerChannel", l.ID):
				l.handleRegisterSession(payload)
			case fmt.Sprintf("lobby.%v.unregisterChannel", l.ID):
				l.handleUnregisterSession(payload)
			case fmt.Sprintf("lobby.%v.payloadChannel", l.ID):
				l.handlePayload(payload)
			}

		case <-ticker.C:
			err := sub.Ping(ctx)
			if err != nil {
				return
			}
		}

	}
}

func (l *Lobby) publishResponse(response WsResponse) error {
	l.logger.Info("lobby.publishResponse",
		slog.Group("sending response to players",
			slog.String("lobby_id", l.ID)))

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()
	responseChanKey := fmt.Sprintf("lobby.%v.responseChannel", l.ID)

	rb, err := response.MarshalBinary()
	if err != nil {
		l.logger.Error("wsClient.PublishPayloadToLobby",
			slog.Group("failed to marshal payload",
				slog.String("reason", err.Error())))
		return err
	}

	err = l.redisClient.Publish(ctx, responseChanKey, rb).Err()
	if err != nil {
		l.logger.Error("lobby.publishResponse", slog.Group("error trying to publish", slog.String("lobby_id", l.ID)))
		return err
	}
	return nil
}

func (l *Lobby) handleRegisterSession(payload WsPayload) {
	l.logger.Info("lobby.handleRegisterSession",
		slog.Group("Registering player connection",
			slog.String("lobby_id", l.ID),
			slog.String("user", payload.Username)))

	l.Players[payload.Username] = struct{}{}

	l.lobbyRepo.SetPlayer(l.ID, &internal.Player{
		Username: payload.Username,
		LobbyId:  l.ID,
		Color:    "",
		Ready:    false,
	})

}

func (l *Lobby) handleUnregisterSession(payload WsPayload) {
	l.logger.Info("lobby.handleUnregisterSession",
		slog.Group("Unregistering player connection",
			slog.String("lobby_id", l.ID),
			slog.String("user", payload.Username)))

	if _, ok := l.Players[payload.Username]; ok {
		delete(l.Players, payload.Username)
		l.lobbyRepo.SetLobby(toLobbyState(l))
		l.lobbyRepo.DeletePlayer(l.ID, payload.Username)
	}

}

func (l *Lobby) handleReadyState() {
	l.logger.Info("lobby.handleReadyState",
		slog.Group("player is ready",
			slog.String("lobby_id", l.ID)))

	var response WsResponse

	// allReady used to start the game
	allReady := true
	// for s := range l.Sessions {
	// 	// if any player is not ready allReady is false
	// 	if !s.IsReady {
	// 		allReady = false
	// 	}
	// }
	// once all players are ready start the game
	if allReady {
		l.logger.Info("lobby.Listen",
			slog.Group("All players are ready game starting"))

		response.Action = StartGameResponseEvent
	}

}

func (l *Lobby) handlePayload(payload WsPayload) {
	var response WsResponse

	switch payload.Action {
	case JoinPayloadEvent:
		response.Action = JoinResponseEvent
		response.Message = fmt.Sprintf("%v joined", payload.Username)
		response.SkipSender = true
		response.Sender = payload.Username
		response.ConnectedUsers = l.getPlayerUsernames()
		if err := l.publishResponse(response); err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}

	case LeavePayloadEvent:
		response.Action = LeftResponseEvent
		response.Message = fmt.Sprintf("%v left", payload.Username)
		response.SkipSender = true
		response.Sender = payload.Username
		response.ConnectedUsers = l.getPlayerUsernames()
		if err := l.publishResponse(response); err != nil {
		}

	case ChatPayloadEvent:
		response.Action = NewMessageResponseEvent
		response.Message = payload.Message
		response.SkipSender = false
		response.Sender = payload.Username
		response.ConnectedUsers = l.getPlayerUsernames()
		if err := l.publishResponse(response); err != nil {
		}

	case ChooseColorPayloadEvent:
		response.Action = ChooseColorResponseEvent
		response.Sender = payload.Username
		response.Message = payload.Message
		response.SkipSender = false
		if err := l.publishResponse(response); err != nil {
		}

	case SetReadyStatusPayloadEvent:
		response.Action = SetReadyStatusResponseEvent
		response.Message = payload.Message
		response.Sender = payload.Username
		response.SkipSender = false
		if err := l.publishResponse(response); err != nil {
		}

	}

}

func (l *Lobby) handleNoPlayers() bool {
	if len(l.Players) == 0 {
		l.logger.Info("lobby.handleNoPlayers",
			slog.Group("triggering closing lobby",
				slog.String("reason", "no players in lobby"),
				slog.String("lobby_id", l.ID),
			))
		return true
	}

	return false

}

func (l *Lobby) getPlayerUsernames() []string {
	var usernames []string
	for u := range l.Players {
		usernames = append(usernames, u)
	}

	sort.Strings(usernames)

	return usernames
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
		ColorsAvailable: l.AvailableColors,
		Settings:        l.Settings,
	}
}
