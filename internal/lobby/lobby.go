package lobby

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/spacesedan/go-sequence/internal/components"
	"github.com/spacesedan/go-sequence/internal/game"
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

	if len(id) != 0 {
		lobbyId = id[0]
	} else {
		lobbyId = generateUniqueLobbyId()
	}
	m.logger.Info("Creating a new game lobby", slog.String("lobbyId", lobbyId))

	colors := make(map[string]bool, 3)
	colors["red"] = true
	colors["blue"] = true
	colors["green"] = true

	l := &GameLobby{
		ID:              lobbyId,
		Game:            game.NewGameService(),
		Players:         make(map[string]*WsConnection),
		Settings:        settings,
		AvailableColors: colors,

		Sessions:       make(map[*WsConnection]struct{}),
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
	for {
		select {
		case session := <-l.RegisterChan:
			l.logger.Info("[REGISTERING]", slog.String("user", session.Username))
			l.Sessions[session] = struct{}{}

		case session := <-l.UnregisterChan:
			l.logger.Info("[UNREGISTERING]", slog.String("user", session.Username))
			if _, ok := l.Sessions[session]; ok {
				delete(l.Sessions, session)
				close(session.Send)
			}

		// handle the incomign payload
		case payload := <-l.PayloadChan:
			switch payload.Action {
			case "join_lobby":
				response.Action = "join_lobby"
				response.Message = fmt.Sprintf("%v joined", payload.SenderSession.Username)
				response.SkipSender = true
				response.PayloadSession = payload.SenderSession
				l.broadcastResponse(response)

			case "left_lobby":
				response.Action = "left"
				response.Message = fmt.Sprintf("%v left", payload.SenderSession.Username)
				response.SkipSender = true
				response.PayloadSession = payload.SenderSession
				l.broadcastResponse(response)

			case "chat_message":
				response.Action = "new_chat_message"
				response.Message = payload.Message
				response.SkipSender = false
				response.PayloadSession = payload.SenderSession
				l.broadcastResponse(response)

			case "choose_color":
				response.PayloadSession = payload.SenderSession
				response.Message = payload.Message
				response.SkipSender = false
				response.Action = "choose_color"
				fmt.Println("Choosing color", l.AvailableColors)
				l.broadcastResponse(response)

			case "sync_colors":
				l.syncColors()
			}

			// add some logic to close the lobby if there are no players
			// case <-time.After(5 * time.Second):
		}
	}
}

func (l *GameLobby) syncColors() {
	var b bytes.Buffer
	for color, available := range l.AvailableColors {
		if available {
			components.PlayerColorComponent(color).Render(context.Background(), &b)
		} else {
			components.PlayerColorUnavailableComponent(color).Render(context.Background(), &b)
		}
		l.broadcastResponse(WsResponse{Action: "sync_colors", Message: b.String()})
		b.Reset()
	}

}

func (l *GameLobby) broadcastResponse(response WsResponse) {
	for session := range l.Sessions {
		session.Send <- response
	}
}
