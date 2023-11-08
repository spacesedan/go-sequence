package lobby

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spacesedan/go-sequence/internal/db"
	"github.com/spacesedan/go-sequence/internal/game"
)

type Lobby struct {
	// game data
	ID              string
	Game            game.GameService
	AvailableColors map[string]bool
	Settings        game.Settings
	Players         map[string]struct{}
	Sessions        map[*WsClient]struct{}

	// connection stuff
	// gets the incoming messages from players
	PayloadChan chan WsPayload
	ReadyChan   chan *WsClient
	// registers players to the lobby
	RegisterChan chan *WsClient
	// reconnectes players to the lobby
	ReconnectChan chan *WsClient
	// unregisters players from the lobby
	UnregisterChan chan *WsClient

	abort chan struct{}

	lobbyState *db.LobbyState
	lobbyRepo  db.LobbyRepo

	//
	lobbyManager *LobbyManager
	logger       *slog.Logger
	redisClient  *redis.Client
}

// Create a new lobby
func (m *LobbyManager) NewLobby(settings game.Settings, id ...string) string {
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
		Players:         map[string]struct{}{},
		Settings:        settings,
		AvailableColors: colors,

		Sessions:       make(map[*WsClient]struct{}),
		PayloadChan:    make(chan WsPayload),
		ReadyChan:      make(chan *WsClient),
		RegisterChan:   make(chan *WsClient),
		ReconnectChan:  make(chan *WsClient),
		UnregisterChan: make(chan *WsClient),

		lobbyManager: m,
		logger:       m.logger,
		redisClient:  m.redisClient,
		lobbyRepo:    db.NewLobbyRepo(m.redisClient, m.logger),
	}

	l.lobbyRepo.SetLobby(&db.LobbyState{
		ID:              l.ID,
		Players:         l.Players,
		Settings:        l.Settings,
		ColorsAvailable: l.AvailableColors,
	})

	m.Lobbies[lobbyId] = l

	go l.Listen()
	go l.Subscribe()

	return lobbyId
}

func (m *LobbyManager) CloseLobby(id string) {
	m.lobbiesMu.Lock()
	defer m.lobbiesMu.Unlock()

	lobby := m.Lobbies[id]
	close(lobby.PayloadChan)
	close(lobby.ReadyChan)
	close(lobby.RegisterChan)
	close(lobby.UnregisterChan)

	lobby.lobbyRepo.DeleteLobby(lobby.ID)

	m.logger.Info("lobbyManager.CloseLobby",
		slog.Group("Closing Lobby",
			slog.String("lobby_id", id)))

	delete(m.Lobbies, id)
}

// Subscribe listens to the lobby payload channel and once it recieves a payload it
// sends a response to the appropriate channel
func (l *Lobby) Subscribe() {
	payloadChanKey := fmt.Sprintf("lobby.%v.payloadChannel", l.ID)
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(time.Minute)
	sub := l.redisClient.Subscribe(ctx, payloadChanKey)

	defer func() {
		sub.Unsubscribe(ctx, payloadChanKey)
		sub.Close()

		ticker.Stop()
		cancel()
	}()

	ch := sub.Channel()

	for {
		select {
		case msg := <-ch:
			var payload WsPayload
			if err := payload.Unmarshal(msg.Payload); err != nil {
				l.logger.Error("lobby.Subscribe",
					slog.Group("failed to unmarshal payload",
						slog.Any("reason", err)))
				return
			}

			l.handlePayload(payload)

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

// this function will handle
func (l *Lobby) Listen() {
	t := time.NewTicker(time.Second * 3)
	_, cancel := context.WithCancel(context.Background())

	defer func() {
		l.lobbyManager.UnregisterChan <- l
		t.Stop()
		cancel()
	}()

	l.logger.Info("lobby.Listen",
		slog.Group("Listening for incoming payloads",
			slog.String("lobby_id", l.ID)))

	for {
		select {
		// case when a session connects to the ws server
		case session := <-l.RegisterChan:
			l.handleRegisterSession(session)

		case session := <-l.ReconnectChan:
			l.handlerReconnectingSession(session)

		// case when a session connection is closed
		case session := <-l.UnregisterChan:
			l.handleUnregisterSession(session)

		// gets sessions that are ready to start the game
		case session := <-l.ReadyChan:
			l.handleReadyState(session)

		// case when sessions send payloads to the lobby
		// case payload := <-l.PayloadChan:
		// 	l.handlePayload(payload)
		// for prod: if there are no players in the lobby the lobby will close
		// after a certain time.
		case <-t.C:
			// ok := l.handleNoPlayers()
			// if ok {
			// 	return
			// }
		}
	}
}

func (l *Lobby) handleRegisterSession(session *WsClient) {
	l.logger.Info("lobby.handleRegisterSession",
		slog.Group("Registering player connection",
			slog.String("lobby_id", l.ID),
			slog.String("user", session.Username)))

	l.Sessions[session] = struct{}{}
	l.Players[session.Username] = struct{}{}

	l.lobbyRepo.SetPlayer(l.ID, &db.PlayerState{
		Color:    session.Color,
		Username: session.Username,
		Ready:    session.IsReady,
		LobbyId:  l.ID,
	})

}

// handlerReconnectingSession attemps to reconnect a player to the lobby
// if not reconnection is attempted within a certain time the player that
// disconnected will be unregistered
func (l *Lobby) handlerReconnectingSession(s *WsClient) {
	l.logger.Info("lobby.handlerReconnectingSession",
		slog.Group("attempting to reconnect player",
			slog.String("lobby_id", l.ID),
			slog.String("username", s.Username)))

	for {
		select {
		case <-time.After(10 * time.Second):
			l.logger.Info("lobby.handlerReconnectingSession",
				slog.Group("reconn window closed, unregistering player",
					slog.String("lobby_id", l.ID),
					slog.String("username", s.Username)))
			l.UnregisterChan <- s
			return
		}
	}
}

func (l *Lobby) handleUnregisterSession(session *WsClient) {
	l.logger.Info("lobby.handleUnregisterSession",
		slog.Group("Unregistering player connection",
			slog.String("lobby_id", l.ID),
			slog.String("user", session.Username)))

	if _, ok := l.Sessions[session]; ok {
		delete(l.Sessions, session)
		delete(l.Players, session.Username)
		l.lobbyRepo.SetLobby(toLobbyState(l))
		l.lobbyRepo.DeletePlayer(l.ID, session.Username)
		close(session.Send)
	}

}

func (l *Lobby) handleReadyState(session *WsClient) {
	l.logger.Info("lobby.handleReadyState",
		slog.Group("player is ready",
			slog.String("lobby_id", l.ID),
			slog.String("username", session.Username)))

	var response WsResponse

	// allReady used to start the game
	allReady := true
	for s := range l.Sessions {
		// if any player is not ready allReady is false
		if !s.IsReady {
			allReady = false
		}
	}
	// once all players are ready start the game
	if allReady {
		l.logger.Info("lobby.Listen",
			slog.Group("All players are ready game starting"))

		response.Action = StartGameResponseEvent
		l.broadcastResponse(response)
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

func (l *Lobby) broadcastResponse(response WsResponse) {
	for session := range l.Sessions {
		session.Send <- response
	}
}

func (l *Lobby) HasPlayer(username string) bool {
	if _, ok := l.Players[username]; ok {
		return true
	}
	return false
}

func toPlayerState(c *WsClient) *db.PlayerState {
	return &db.PlayerState{
		LobbyId:  c.Lobby.ID,
		Username: c.Username,
		Color:    c.Color,
		Ready:    c.IsReady,
	}
}

func toLobbyState(l *Lobby) *db.LobbyState {
	return &db.LobbyState{
		ID:              l.ID,
		Players:         l.Players,
		ColorsAvailable: l.AvailableColors,
		Settings:        l.Settings,
	}
}
