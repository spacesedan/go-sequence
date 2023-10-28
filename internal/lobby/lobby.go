package lobby

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/nitishm/go-rejson/v4"
	"github.com/spacesedan/go-sequence/internal/game"
)

const (
	lobbyEventJoin           = "join_lobby"
	lobbyEventLeft           = "left_lobby"
	lobbyEventChatMessage    = "chat_message"
	lobbyEventChooseColor    = "choose_color"
	lobbyEventSyncColors     = "sync_colors"
	lobbyEventSetReadyStatus = "set_ready_status"
)

type PlayerState struct {
	LobbyId  string `json:"lobby_id"`
	Username string `json:"username"`
	Color    string `json:"color"`
	Ready    bool   `json:"ready"`
}

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

	//
	lobbyManager     *LobbyManager
	logger           *slog.Logger
	redisJSONHandler *rejson.Handler
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

		lobbyManager:     m,
		logger:           m.logger,
		redisJSONHandler: m.redisJSONHandler,
	}

	m.Lobbies[lobbyId] = l

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
		l.lobbyManager.CloseLobby(l.ID)
		t.Stop()
	}()

	l.logger.Info("Listening for incoming payloads", slog.String("lobby_id", l.ID))

	for {
		select {
		// case when a session connects to the ws server
		case session := <-l.RegisterChan:
			l.logger.Info("[REGISTERING]", slog.String("user", session.Username))
			l.Sessions[session] = struct{}{}
			l.Players[session.Username] =
				&PlayerState{
					Username: session.Username,
					LobbyId:  l.ID,
				}

				// redis things
			err := l.setPlayerJSON(
				session,
				&PlayerState{
					Username: session.Username,
					LobbyId:  l.ID,
				})
			if err != nil {
				l.logger.Error("Something went wrong", slog.String("err", err.Error()))
			}
			ps, _ := l.getPlayerJSON(session)
			fmt.Printf("[PLAYER] %#v\n", ps)

			// case when a session connection is closed
		case session := <-l.UnregisterChan:
			l.logger.Info("[UNREGISTERING]", slog.String("user", session.Username))
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
				response.Action = "start_game"
				l.broadcastResponse(response)
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

			case lobbyEventSetReadyStatus:
				response.Action = "set_ready_status"
				response.Message = payload.Message
				response.PayloadSession = payload.SenderSession
				response.SkipSender = false
				l.broadcastResponse(response)

			}

		case <-t.C:
			l.logger.Info("Checking for players",
				slog.String("lobby_id", l.ID),
				slog.Int("number of connected players", len(l.Players)))
			if len(l.Players) == 0 {
				return
			}

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

func (l *GameLobby) redisKey(s *WsConnection) string {
	return fmt.Sprintf("lobby_id-%v|username-%v", l.ID, s.Username)
}

func (l *GameLobby) setPlayerJSON(s *WsConnection, ps *PlayerState) error {
	l.logger.Info("Setting player state to cache")
	res, err := l.redisJSONHandler.JSONSet(l.redisKey(s), ".", ps)
	if err != nil {
		return err
	}

	if res.(string) == "OK" {
		l.logger.Info("Successfully set to cache")
	} else {
		l.logger.Info("Failed to cache")
	}
	return nil
}

func (l *GameLobby) getPlayerJSON(s *WsConnection) (*PlayerState, error) {
	var ps *PlayerState

	pj, err := redis.Bytes(l.redisJSONHandler.JSONGet(l.redisKey(s), "."))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(pj, &ps)
	if err != nil {
		return nil, err
	}

	return ps, nil
}
