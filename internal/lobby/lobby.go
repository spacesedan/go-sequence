package lobby

import (
	"fmt"
	"log/slog"

	"github.com/spacesedan/go-sequence/internal/game"
)

type GameLobby struct {
	// game data
	ID              string
	Game            game.GameService
	AvailableColors map[string]bool
	Settings        Settings
	Players         map[string]*WsConnection

	// connection stuff
	// gets the incoming messages from players
	PayloadChan chan WsPayload
	// returns the response to send to the players
	ResponseChan chan WsResponse
	// registers players to the lobby
	RegisterChan chan *WsConnection
	// unregisters players from the lobby
	UnregisterChan chan *WsConnection

	//
	lobbyManager *LobbyManager
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

		PayloadChan:    make(chan WsPayload),
		ResponseChan:   make(chan WsResponse),
		RegisterChan:   make(chan *WsConnection),
		UnregisterChan: make(chan *WsConnection),

		lobbyManager: m,
	}

	m.Lobbies[lobbyId] = l

	go l.Listen()

	return lobbyId
}

func (l *GameLobby) Listen() {
	for {
		select {
		case _ = <-l.RegisterChan:
		case _ = <-l.UnregisterChan:
		case _ = <-l.PayloadChan:
			// for _, session := range l.Players {
			// 	select {
			// 	case session.Send <- payload:
			// 	default:
			// 		session.Conn.Close()
			// 		close(session.Send)
			// 	}
			// }
		case _ = <-l.ResponseChan:

		}
	}
}
