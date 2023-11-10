package client

import (
	"bytes"
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
	"github.com/spacesedan/go-sequence/internal/views/components"
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

	playerState internal.Player
	clientRepo  db.ClientRepo
	redisClient *redis.Client
	logger      *slog.Logger
}

type PublishChannel uint

const (
	UnknownChannel PublishChannel = iota
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
	default:
		return "unknown"
	}
}

func NewWsClient(ws *websocket.Conn, r *redis.Client, logger *slog.Logger, username, lobbyId string) *WsClient {
	// how should i get the redis client

	return &WsClient{
		Conn:     ws,
		Username: username,
		LobbyID:  lobbyId,

		clientRepo:  db.NewClientRepo(r, logger),
		redisClient: r,
		logger:      logger,
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

		// when closing the read pump publish a message to the lobby to remove
		// the player
		s.publishToLobby(UnregisterChannel, lobby.WsPayload{
			Action:   "unregister",
			Username: s.Username,
		})
		s.Conn.Close()
	}()

	s.Conn.SetReadDeadline(time.Now().Add(pongWait))
	s.Conn.SetPongHandler(func(string) error { s.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// send a message to the lobby to register the player
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

	p := fmt.Sprintf("lobby.%v.responseChannel", s.LobbyID)
	ctx, cancel := context.WithCancel(context.Background())
	sub := s.redisClient.Subscribe(ctx, p)
	ticker := time.NewTicker(time.Minute)

	defer func() {
		sub.Unsubscribe(ctx, p)
		sub.Close()
		cancel()

		ticker.Stop()
	}()

	ch := sub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				s.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			var b bytes.Buffer
			var response lobby.WsResponse
			err := response.Unmarshal(msg.Payload)
			if err != nil {
				fmt.Printf("ERR: %v\n", err)
			}

			fmt.Printf("%#v\n", response)

			switch response.Action {
			case lobby.JoinResponseEvent:
				if response.Sender != s.Username {
					components.PlayerStatus(response.Message).Render(ctx, &b)
					if err := s.sendResponse(b.String()); err != nil {
						fmt.Println("[ACTION] join_lobby", err.Error())
						return
					}
				}
				b.Reset()

				components.PlayerDetails(response.ConnectedUsers).Render(ctx, &b)
				if err := s.sendResponse(b.String()); err != nil {
					return
				}
				b.Reset()

			case lobby.NewMessageResponseEvent:
				alt := fmt.Sprintf("avatar image for %v", s.Username)

				if response.Sender == s.Username {
					components.ChatMessageSender(
						response.Message,
						alt,
						generateUserAvatar(response.Sender, 32)).
						Render(ctx, &b)
				} else {
					components.ChatMessageReciever(
						response.Message,
						alt,
						generateUserAvatar(response.Sender, 32)).
						Render(ctx, &b)
				}
				if err := s.sendResponse(b.String()); err != nil {
					fmt.Println("[ACTION] new_chat_message", err.Error())
					return
				}

				b.Reset()
			case lobby.ChooseColorResponseEvent:
				if s.Color == "" {
					s.setColor(response)
				} else {
					s.updateColor(response)
				}

				sender, err := s.clientRepo.GetPlayer(s.LobbyID, response.Sender)
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
					return
				}

				err = components.PlayerDetailsColored(
					sender.Username,
					response.Message,
					sender.Ready,
				).Render(ctx, &b)
				if err != nil {
					fmt.Printf("ERR: %v\n", err)
				}

				if err := s.sendResponse(b.String()); err != nil {
					return
				}
				b.Reset()

			case lobby.SetReadyStatusResponseEvent:
				if response.Sender == s.Username {
					if s.Color == "" {
						title := "Missing player color"
						content := "can't ready up without selecting a color"
						components.ToastWSComponent(title, content).Render(context.Background(), &b)
						s.sendResponse(b.String())
						b.Reset()
						continue
					}
				}

				sender, err := s.clientRepo.GetPlayer(s.LobbyID, response.Sender)
				if err != nil {
					return
				}

				components.PlayerDetailsColored(
					sender.Username,
					sender.Color,
					sender.Ready,
				).
					Render(context.Background(), &b)

				if err := s.sendResponse(b.String()); err != nil {
					return
				}
				b.Reset()
			}

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

// func (s *WsClient) WritePump() {
// 	s.logger.Info("wsClient.WritePump",
// 		slog.Group("starting write pump",
// 			slog.String("lobby_id", s.LobbyID),
// 			slog.String("usename", s.Username),
// 		))
//
// 	var b bytes.Buffer
// 	ticker := time.NewTicker(pingPeriod)
// 	defer func() {
// 		s.logger.Info("wsClient.WritePump",
// 			slog.Group("closing write pump",
// 				slog.String("username", s.Username)))
//
// 		ticker.Stop()
// 		s.Conn.Close()
// 	}()
//
// 	for {
// 		select {
// 		case response, ok := <-s.Send:
// 			if !ok {
// 				s.Conn.WriteMessage(websocket.CloseMessage, []byte{})
// 				return
// 			}
//
// 			// Subscribes to incoming messages from the client send channel based on
// 			// the incoming response action
// 			switch response.Action {
//
// 			case lobby.JoinResponseEvent:
// 				if response.Sender != s.Username {
// 					components.PlayerStatus(response.Message).Render(context.Background(), &b)
// 					if err := s.sendResponse(b.String()); err != nil {
// 						fmt.Println("[ACTION] join_lobby", err.Error())
// 						return
// 					}
// 				}
// 				b.Reset()
//
// 				components.PlayerDetails(response.ConnectedUsers).Render(context.Background(), &b)
// 				if err := s.sendResponse(b.String()); err != nil {
// 					return
// 				}
// 				b.Reset()
//
// 			case lobby.NewMessageResponseEvent:
// 				alt := fmt.Sprintf("avatar image for %v", s.Username)
//
// 				if response.Sender == s.Username {
// 					components.ChatMessageSender(
// 						response.Message,
// 						alt,
// 						generateUserAvatar(response.Sender, 32)).
// 						Render(context.Background(), &b)
// 				} else {
// 					components.ChatMessageReciever(
// 						response.Message,
// 						alt,
// 						generateUserAvatar(response.Sender, 32)).
// 						Render(context.Background(), &b)
// 				}
// 				if err := s.sendResponse(b.String()); err != nil {
// 					fmt.Println("[ACTION] new_chat_message", err.Error())
// 					return
// 				}
//
// 				b.Reset()
//
// 			case lobby.ChooseColorResponseEvent:
// 				if s.Color == "" {
// 					s.setColor(response)
// 				} else {
// 					s.updateColor(response)
// 				}
//
// 				sender, err := s.clientRepo.GetPlayer(s.LobbyID, response.Sender)
// 				if err != nil {
// 					return
// 				}
//
// 				components.PlayerDetailsColored(
// 					sender.Username,
// 					sender.Color,
// 					sender.Ready,
// 				).
// 					Render(context.Background(), &b)
//
// 				if err := s.sendResponse(b.String()); err != nil {
// 					return
// 				}
// 				b.Reset()
//
// 			case lobby.LeftResponseEvent:
// 				components.PlayerStatus(response.Message).Render(context.Background(), &b)
// 				if err := s.sendResponse(b.String()); err != nil {
// 					fmt.Println("[ACTION] left", err.Error())
// 					return
// 				}
//
// 				b.Reset()
//
// 			case lobby.SetReadyStatusResponseEvent:
// 				if response.Sender == s.Username {
// 					if s.Color == "" {
// 						title := "Missing player color"
// 						content := "can't ready up without selecting a color"
// 						components.ToastWSComponent(title, content).Render(context.Background(), &b)
// 						s.sendResponse(b.String())
// 						b.Reset()
// 						continue
// 					}
// 					s.Lobby.ReadyChan <- s
// 				}
//
// 				sender, err := s.clientRepo.GetPlayer(s.LobbyID, response.Sender)
// 				if err != nil {
// 					return
// 				}
//
// 				components.PlayerDetailsColored(
// 					sender.Username,
// 					sender.Color,
// 					sender.Ready,
// 				).
// 					Render(context.Background(), &b)
//
// 				if err := s.sendResponse(b.String()); err != nil {
// 					return
// 				}
// 				b.Reset()
//
// 			case lobby.StartGameResponseEvent:
// 				s.logger.Info("Starting game")
// 				views.Game(createWebsocketConnectionString(s.LobbyID)).Render(context.Background(), &b)
// 				fmt.Printf("%v\n", b.String())
// 				s.sendResponse(b.String())
// 				b.Reset()
//
// 			default:
// 				fmt.Printf("\n%#v\n", response)
// 				return
// 			}
// 		case <-ticker.C:
// 			s.Conn.SetWriteDeadline(time.Now().Add(writeWait))
// 			if err := s.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
// 				fmt.Println("[ACTION] ticker", err.Error())
// 				return
// 			}
// 		}
//
// 	}
// }

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
