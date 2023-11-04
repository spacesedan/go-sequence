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
	Players         map[string]*PlayerState
	Sessions        map[*WsConnection]struct{}

	// connection stuff
	// gets the incoming messages from players
	PayloadChan chan WsPayload
	ReadyChan   chan *WsConnection
	// registers players to the lobby
	RegisterChan chan *WsConnection
	// unregisters players from the lobby
	UnregisterChan chan *WsConnection

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
	m.logger.Info("Creating a new game lobby", slog.String("lobbyId", lobbyId))

	colors["red"] = true
	colors["blue"] = true
	colors["green"] = true

	l := &GameLobby{
		ID:              lobbyId,
		Game:            game.NewGameService(),
		Players:         make(map[string]*PlayerState, settings.NumOfPlayers),
		Settings:        settings,
		AvailableColors: colors,

		Sessions:       make(map[*WsConnection]struct{}, settings.NumOfPlayers),
		PayloadChan:    make(chan WsPayload),
		ReadyChan:      make(chan *WsConnection, settings.NumOfPlayers),
		RegisterChan:   make(chan *WsConnection, settings.NumOfPlayers),
		UnregisterChan: make(chan *WsConnection, settings.NumOfPlayers),

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

	m.logger.Info("Closing Lobby", slog.String("lobby_id", id))

	delete(m.Lobbies, id)
}

func (l *GameLobby) Listen() {
	var response WsResponse
	t := time.NewTicker(time.Second * 10)

	defer func() {
		l.lobbyManager.UnregisterChan <- l
		t.Stop()
	}()

	l.logger.Info("Listening for incoming payloads", slog.String("lobby_id", l.ID))

	for {
		select {
		// case when a session connects to the ws server
		case session := <-l.RegisterChan:
			l.logger.Info("lobby.Listen",
				slog.Group("Registering player connection",
					slog.String("lobby_id", l.ID),
					slog.String("user", session.Username)))

			ps, _ := l.lobbyRepo.GetPlayer(l.ID, session.Username)
			if ps == nil {
				l.logger.Info("player state not found creating a new record", slog.String("username", session.Username))
				err := l.lobbyRepo.SetPlayer(
					l.ID,
					session,
				)
				if err != nil {
					l.logger.Error("Something went wrong", slog.String("err", err.Error()))
				}

			}

			l.Sessions[session] = struct{}{}
			l.Players[session.Username] = ps

			// Set the player in the lobby state

			// case when a session connection is closed
		case session := <-l.UnregisterChan:
			l.logger.Info("lobby.Listen",
				slog.Group("Unregistering player connection",
					slog.String("lobby_id", l.ID),
					slog.String("user", session.Username)))

			if _, ok := l.Sessions[session]; ok {
				delete(l.Sessions, session)
				delete(l.Players, session.Username)
				close(session.Send)
			}

			// gets sessions that are ready to start the game
		case session := <-l.ReadyChan:
			l.logger.Info("Player ready",
				slog.String("lobby_id", l.ID),
				slog.String("username", session.Username))

			// allReady used to start the game
			allReady := true
			for s := range l.Sessions {
				// if any player is not ready allReady is false
				if !s.IsReady {
					l.logger.Info("Player not ready", slog.String("username", s.Username))
					allReady = false
				}
			}
			// once all players are ready start the game
			if allReady {
				response.Action = StartGameResponseEvent
				l.broadcastResponse(response)
			}

		// case when sessions send payloads to the lobby
		case payload := <-l.PayloadChan:
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

		// for prod: if there are no players in the lobby the lobby will close
		// after a certain time.
		// case <-t.C:
		// 	if len(l.Players) == 0 {
		// 		l.logger.Info("lobby.Listen",
		// 			slog.Group("triggering closing lobby",
		// 				slog.String("reason", "no players in lobby"),
		// 				slog.String("lobby_id", l.ID),
		// 			))
		// 		return
		// 	}
		//
		}
	}
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
