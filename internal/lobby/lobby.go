package lobby

import (
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/spacesedan/go-sequence/internal/game"
)

type GameLobby struct {
	// game data
	ID              string
	Game            game.GameService
	AvailableColors map[string]bool
	Settings        Settings
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

	lobbyState *LobbyState
	lobbyRepo  *LobbyRepo

	//
	lobbyManager *LobbyManager
	logger       *slog.Logger
}

// Create a new lobby
func (m *LobbyManager) NewGameLobby(settings Settings, id ...string) string {
	m.lobbiesMu.Lock()
	defer m.lobbiesMu.Unlock()

	var lobbyId string
	colors := make(map[string]bool, 3)

	if len(id) != 0 {
		lobbyId = id[0]
	} else {
		lobbyId = generateUniqueLobbyId()
	}

	m.logger.Info("lobbyManager.NewGameLobby",
		slog.Group("Creating new lobby",
			slog.String("lobbyId", lobbyId)))

	colors["red"] = true
	colors["blue"] = true
	colors["green"] = true

	l := &GameLobby{
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

		abort:        make(chan struct{}),
		lobbyManager: m,
		lobbyRepo:    NewLobbyRepo(m.redisPool, m.logger),
		logger:       m.logger,
	}

	m.Lobbies[lobbyId] = l

	l.lobbyRepo.SetLobby(l)

	go l.Listen()

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

func (l *GameLobby) Listen() {
	t := time.NewTicker(time.Second * 10)

	defer func() {
		l.lobbyManager.UnregisterChan <- l
		t.Stop()
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
		case payload := <-l.PayloadChan:
			l.handlePayload(payload)
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

func (l *GameLobby) handleRegisterSession(session *WsClient) {
	l.logger.Info("lobby.handleRegisterSession",
		slog.Group("Registering player connection",
			slog.String("lobby_id", l.ID),
			slog.String("user", session.Username)))

	l.Sessions[session] = struct{}{}
	l.Players[session.Username] = struct{}{}

	l.lobbyRepo.SetPlayer(l.ID, session)

}

// handlerReconnectingSession attemps to reconnect a player to the lobby
// if not reconnection is attempted within a certain time the player that
// disconnected will be unregistered
func (l *GameLobby) handlerReconnectingSession(s *WsClient) {
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

func (l *GameLobby) handleUnregisterSession(session *WsClient) {
	l.logger.Info("lobby.handleUnregisterSession",
		slog.Group("Unregistering player connection",
			slog.String("lobby_id", l.ID),
			slog.String("user", session.Username)))

	if _, ok := l.Sessions[session]; ok {
		delete(l.Sessions, session)
		delete(l.Players, session.Username)
		l.lobbyRepo.SetLobby(l)

		l.lobbyRepo.DeletePlayer(l.ID, session.Username)
		close(session.Send)
	}

}

func (l *GameLobby) handleReadyState(session *WsClient) {
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

func (l *GameLobby) handlePayload(payload WsPayload) {
	var response WsResponse

	switch payload.Action {
	case JoinPayloadEvent:
		response.Action = JoinResponseEvent
		response.Message = fmt.Sprintf("%v joined", payload.SenderSession.Username)
		response.SkipSender = true
		response.PayloadSession = payload.SenderSession
		response.ConnectedUsers = l.getPlayerUsernames()
		l.broadcastResponse(response)

	case LeavePayloadEvent:
		response.Action = LeftResponseEvent
		response.Message = fmt.Sprintf("%v left", payload.SenderSession.Username)
		response.SkipSender = true
		response.PayloadSession = payload.SenderSession
		l.broadcastResponse(response)

	case ChatPayloadEvent:
		response.Action = NewMessageResponseEvent
		response.Message = payload.Message
		response.SkipSender = false
		response.PayloadSession = payload.SenderSession
		l.broadcastResponse(response)

	case ChooseColorPayloadEvent:
		response.Action = ChooseColorResponseEvent
		response.PayloadSession = payload.SenderSession
		response.Message = payload.Message
		response.SkipSender = false
		l.broadcastResponse(response)

	case SetReadyStatusPayloadEvent:
		response.Action = SetReadyStatusResponseEvent
		response.Message = payload.Message
		response.PayloadSession = payload.SenderSession
		response.SkipSender = false
		l.broadcastResponse(response)

	}

}

func (l *GameLobby) handleNoPlayers() bool {
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

func (l *GameLobby) getPlayerUsernames() []string {
	var usernames []string
	for u := range l.Players {
		usernames = append(usernames, u)
	}

	sort.Strings(usernames)

	return usernames
}

func (l *GameLobby) broadcastResponse(response WsResponse) {
	for session := range l.Sessions {
		session.Send <- response
	}
}

func (l *GameLobby) HasPlayer(username string) bool {
	if _, ok := l.Players[username]; ok {
		return true
	}
	return false
}
