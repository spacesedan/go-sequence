package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal"
	"github.com/spacesedan/go-sequence/internal/db"
	"github.com/spacesedan/go-sequence/internal/lobby"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type WsClient struct {
	Conn     *websocket.Conn
	Username string
	LobbyID  string

	Color   string
	IsReady bool

	playerState *internal.Player
	clientRepo  db.ClientRepo
	redisClient *redis.Client
	logger      *slog.Logger
	errorChan   chan error
}

type PublishChannel uint

const (
	UnknownChannel PublishChannel = iota
	StateChannel
	PayloadChannel
	PingChannel
	RegisterChannel
	UnregisterChannel
)

func (p PublishChannel) String() string {
	switch p {
	case PayloadChannel:
		return "payloadChannel"
	case RegisterChannel:
		return "registerChannel"
	case UnregisterChannel:
		return "unregisterChannel"
	case UnknownChannel:
		return "unknown"
	case PingChannel:
		return "pingChannel"
	case StateChannel:
		return "stateChannel"
	default:
		return "unknown"
	}
}

func NewWsClient(ws *websocket.Conn, r *redis.Client, logger *slog.Logger, username, lobbyId string) *WsClient {

	// how should i get the redis client
	// passed it to the redis client to the lobbyHandler.
	return &WsClient{
		Conn:     ws,
		Username: username,
		LobbyID:  lobbyId,

		playerState: &internal.Player{},
		clientRepo:  db.NewClientRepo(r, logger),
		redisClient: r,
		logger:      logger,
		errorChan:   make(chan error, 1),
	}
}

func (s *WsClient) ReadPump() {
	s.logger.Info("wsClient.ReadPump",
		slog.Group("starting readpump",
			slog.String("lobby_id", s.LobbyID),
			slog.String("usename", s.Username),
		))

	var payload lobby.WsPayload

	defer func() {
		s.logger.Info("wsClient.ReadPump",
			slog.Group("Read Pump closing",
				slog.String("username", s.Username)))

		// unregister the connection when the ws connection closes
		s.publishToLobby(UnregisterChannel, lobby.WsPayload{
			Action:   "unregister",
			Username: s.Username,
		})
		s.Conn.Close()

	}()

	s.Conn.SetReadDeadline(time.Now().Add(pongWait))
	s.Conn.SetPongHandler(func(string) error { s.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// register the session to the lobby
	s.publishToLobby(RegisterChannel, lobby.WsPayload{
		Action:   "register",
		Username: s.Username,
	})

	for {
		err := s.Conn.ReadJSON(&payload)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			}
			s.logger.Error("wsClient.ReadPump",
				slog.Group("Error occrured, terminating readpump",
					slog.String("reason", err.Error())))
			return
		}

		payload.Username = s.Username

		if err := s.publishToLobby(PayloadChannel, payload); err != nil {
			s.logger.Error("wsClient.ReadPump",
				slog.Group("error publishing to lobby",
					slog.String("lobby_id", s.LobbyID),
					slog.String("username", s.Username)))
		}

	}

}

// SubscribeToLobby
func (s *WsClient) SubscribeToLobby() {
	s.logger.Info("wsClient.SubscribeToLobby",
		slog.Group("subscribing to lobby",
			slog.String("lobby_id", s.LobbyID),
			slog.String("username", s.Username)))

	responseChannel := fmt.Sprintf("lobby.%v.responseChannel", s.LobbyID)
	ctx, cancel := context.WithCancel(context.Background())
	sub := s.redisClient.Subscribe(ctx, responseChannel)
	ch := sub.Channel()
	ticker := time.NewTicker(time.Minute)

	defer func() {
		sub.Close()
		cancel()

		ticker.Stop()
	}()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				s.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			var response lobby.WsResponse
			err := response.Unmarshal(msg.Payload)
			if err != nil {
				fmt.Printf("ERR: %v\n", err)
				return
			}

			switch msg.Channel {
			case responseChannel:
				switch response.Action {
				case lobby.JoinLobbyPayloadEvent:
					s.handleJoinLobby(response)
				case lobby.JoinGamePayloadEvent:
					s.handleJoinGame(response)
				case lobby.NewMessageResponseEvent:
					s.handleChatMessage(response)
				case lobby.ChooseColorResponseEvent:
					s.handleChooseColor(response)
				case lobby.SetReadyStatusResponseEvent:
					s.handlePlayerReady(response)
				}
			}

		case <-s.errorChan:
			return

		case <-ticker.C:
			err := sub.Ping(ctx)
			if err != nil {
				return
			}
		}
	}
}

// PublishPayloadToLobby sends a payload to the lobby
func (s *WsClient) publishToLobby(channel PublishChannel, payload lobby.WsPayload) error {
	s.logger.Info("wsClient.PublishPayloadToLobby",
		slog.Group("sending payload"))

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	chanKey := fmt.Sprintf("lobby.%v.%v", s.LobbyID, channel.String())

	pb, err := payload.MarshalBinary()
	if err != nil {
		s.logger.Error("wsClient.PublishPayloadToLobby",
			slog.Group("failed to marshal payload",
				slog.String("reason", err.Error())))

		return err
	}

	err = s.redisClient.Publish(ctx, chanKey, pb).Err()
	if err != nil {
		s.logger.Error("wsClient.PublishToLobby",
			slog.Group("error trying to publish",
				slog.String("username", s.Username)))

		return err
	}
	return nil
}

// setColor sets the Player color
func (s *WsClient) setColor(response lobby.WsResponse) {
	if response.Sender == s.Username {
		s.Color = response.Message
	}

}

// updateColor updates the player color and resets the previous color
func (s *WsClient) updateColor(response lobby.WsResponse) {
	if response.Sender == s.Username {
		if response.Message != s.Color {
			s.Color = response.Message
		}
	}
}

// sendResonse sends the response to the client
func (s *WsClient) sendResponse(msg string) error {
	w, err := s.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil

}

// generateUserAvatar creates a link that will be used by the clinet to fetch a
// avatar image for the the current user
func generateUserAvatar(username string, size int) string {
	u := url.URL{
		Scheme: "https",
		Host:   "api.dicebear.com",
		Path:   "7.x/pixel-art/svg",
	}

	q := u.Query()
	q.Set("seed", username)
	q.Set("size", fmt.Sprintf("%v", size))
	q.Set("radius", "50")
	q.Set("beard", "variant01,variant02,variant03,variant04")

	u.RawQuery = q.Encode()

	return u.String()
}

func createWebsocketConnectionString(lobbyId string) string {
	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:42069",
		Path:   "/lobby/ws",
	}
	q := u.Query()
	q.Set("lobby-id", lobbyId)
	u.RawQuery = q.Encode()

	return u.String()
}
