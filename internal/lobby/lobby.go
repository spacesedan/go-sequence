package lobby

import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/spacesedan/go-sequence/internal/game"
)

const (
	lobbyEventJoin        = "join_lobby"
	lobbyEventLeft        = "left_lobby"
	lobbyEventChatMessage = "chat_message"
	lobbyEventChooseColor = "choose_color"
	lobbyEventSyncColors  = "sync_colors"
)

type GameLobby struct {
	// game data
	ID              string
	Game            game.GameService
	AvailableColors map[string]bool
	Settings        Settings
	Players         map[string]*WsConnection
	Sessions        map[*WsConnection]struct{}

	// connection stuff
	// gets the incoming messages from players
	PayloadChan chan WsPayload
	// registers players to the lobby
	RegisterChan chan *WsConnection
	// unregisters players from the lobby
	UnregisterChan chan *WsConnection

	//
	lobbyManager *LobbyManager
	logger       *slog.Logger
}

// Create a new lobby
func (m *LobbyManager) NewGameLobby(settings Settings, id ...string) string {
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
		Players:         make(map[string]*WsConnection, settings.NumOfPlayers),
		Settings:        settings,
		AvailableColors: colors,

		Sessions:       make(map[*WsConnection]struct{}, settings.NumOfPlayers),
		PayloadChan:    make(chan WsPayload),
		RegisterChan:   make(chan *WsConnection),
		UnregisterChan: make(chan *WsConnection),

		lobbyManager: m,
		logger:       m.logger,
	}

	m.Lobbies[lobbyId] = l

	go l.Listen()

	return lobbyId
}

func (l *GameLobby) Listen() {
	var response WsResponse

	l.logger.Info("Listening for incoming payloads", slog.String("lobby_id", l.ID))

	for {
		select {
		// case when a session connects to the ws server
		case session := <-l.RegisterChan:
			l.logger.Info("[REGISTERING]", slog.String("user", session.Username))
			l.Players[session.Username] = session

			// case when a session connection is closed
		case session := <-l.UnregisterChan:
			l.logger.Info("[UNREGISTERING]", slog.String("user", session.Username))
			if _, ok := l.Players[session.Username]; ok {
				delete(l.Players, session.Username)
				close(session.Send)
			}

		// case when sessions send payloads to the lobby
		case payload := <-l.PayloadChan:
			switch payload.Action {
			case lobbyEventJoin:
				response.Action = "join_lobby"
				response.Message = fmt.Sprintf("%v joined", payload.SenderSession.Username)
				response.SkipSender = true
				response.PayloadSession = payload.SenderSession
				response.ConnectedUsers = l.getPlayerUsernames()
				l.broadcastResponse(response)

			case lobbyEventLeft:
				response.Action = "left"
				response.Message = fmt.Sprintf("%v left", payload.SenderSession.Username)
				response.SkipSender = true
				response.PayloadSession = payload.SenderSession
				l.broadcastResponse(response)

			case lobbyEventChatMessage:
				response.Action = "new_chat_message"
				response.Message = payload.Message
				response.SkipSender = false
				response.PayloadSession = payload.SenderSession
				l.broadcastResponse(response)

			case lobbyEventChooseColor:
				response.PayloadSession = payload.SenderSession
				response.Message = payload.Message
				response.SkipSender = false
				response.Action = "choose_color"
				l.broadcastResponse(response)

			}

			// add some logic to close the lobby if there are no players
			// case <-time.After(5 * time.Second):
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
	for _, session := range l.Players {
		session.Send <- response
	}
}
