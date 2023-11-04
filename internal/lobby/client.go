package lobby

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/components"
	"github.com/spacesedan/go-sequence/internal/views"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type WsConnection struct {
	LobbyManager *LobbyManager
	Lobby        *GameLobby
	Send         chan WsResponse
	Conn         *websocket.Conn
	Username     string
	LobbyID      string

	Color   string
	IsReady bool
}

func (s *WsConnection) ReadPump() {
	defer func() {
		s.LobbyManager.logger.Info("Read Pump closing", slog.String("username", s.Username))
		s.Lobby.UnregisterChan <- s
		s.Conn.Close()
	}()

	s.Conn.SetReadDeadline(time.Now().Add(pongWait))
	s.Conn.SetPongHandler(func(string) error { s.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	var payload WsPayload

	for {
		err := s.Conn.ReadJSON(&payload)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ERROR] %v\n", err)
			}
			s.LobbyManager.logger.Error("[ERROR]", slog.Any("err", err.Error()))
			break
		}

		_, lobbyExists := s.LobbyManager.LobbyExists(s.LobbyID)
		if lobbyExists {
			payload.SenderSession = s
			s.LobbyManager.logger.Info("[PAYLOAD]", slog.Any("payload", payload))
			s.Lobby.PayloadChan <- payload
		} else {
			break
		}
	}

}

func (s *WsConnection) WritePump() {
	var b bytes.Buffer
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		s.LobbyManager.logger.Info("Write Pump closing", slog.String("username", s.Username))
		ticker.Stop()
		s.Conn.Close()
	}()

	for {
		select {
		case response, ok := <-s.Send:
			// access to the player state associated with the session
			playerState, err := s.Lobby.lobbyRepo.GetPlayer(s.Lobby.ID, s.Username)
			if err != nil {
				return
			}

			s.Lobby.logger.Info("Player state", slog.Any("state", playerState))
			if !ok {
				s.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			s.LobbyManager.logger.Info("[INCOMING]", slog.Any("response", response))
			switch response.Action {

			case JoinResponseEvent:
				if response.PayloadSession != s {
					components.PlayerStatus(response.Message).Render(context.Background(), &b)
					if err := s.sendResponse(b.String()); err != nil {
						fmt.Println("[ACTION] join_lobby", err.Error())
						return
					}
				}
				b.Reset()
				components.PlayerDetails(response.ConnectedUsers).Render(context.Background(), &b)
				if err := s.sendResponse(b.String()); err != nil {
					return
				}
				b.Reset()

			case NewMessageResponseEvent:
				alt := fmt.Sprintf("avatar image for %v", s.Username)

				if response.PayloadSession == s {
					components.ChatMessageSender(
						response.Message,
						alt,
						generateUserAvatar(response.PayloadSession.Username, 32)).
						Render(context.Background(), &b)
				} else {
					components.ChatMessageReciever(
						response.Message,
						alt,
						generateUserAvatar(response.PayloadSession.Username, 32)).
						Render(context.Background(), &b)
				}
				if err := s.sendResponse(b.String()); err != nil {
					fmt.Println("[ACTION] new_chat_message", err.Error())
					return
				}

				b.Reset()

			case ChooseColorResponseEvent:
				if s.Color == "" {
					s.setColor(response)
				} else {
					s.updateColor(response)
				}

				components.PlayerDetailsColored(
					response.PayloadSession.Username,
					response.PayloadSession.Color,
					response.PayloadSession.IsReady,
				).
					Render(context.Background(), &b)

				if err := s.sendResponse(b.String()); err != nil {
					return
				}
				b.Reset()

			case LeftResponseEvent:
				components.PlayerStatus(response.Message).Render(context.Background(), &b)
				if err := s.sendResponse(b.String()); err != nil {
					fmt.Println("[ACTION] left", err.Error())
					return
				}

				b.Reset()

			case SetReadyStatusResponseEvent:
				if response.PayloadSession == s {
					if s.Color == "" {
						title := "Missing player color"
						content := "can't ready up without selecting a color"
						components.ToastWSComponent(title, content).Render(context.Background(), &b)
						s.sendResponse(b.String())
						b.Reset()
						continue
					}
					s.Lobby.ReadyChan <- s
				}

				components.PlayerDetailsColored(
					response.PayloadSession.Username,
					response.PayloadSession.Color,
					response.PayloadSession.IsReady,
				).
					Render(context.Background(), &b)

				if err := s.sendResponse(b.String()); err != nil {
					return
				}
				b.Reset()

			case StartGameResponseEvent:
				s.Lobby.logger.Info("Starting game")
				views.Game(createWebsocketConnectionString(s.Lobby.ID)).Render(context.Background(), &b)
				fmt.Printf("%v\n", b.String())
				s.sendResponse(b.String())
				b.Reset()

			default:
				fmt.Printf("\n%#v\n", response)
				return
			}
		case <-ticker.C:
			s.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Println("[ACTION] ticker", err.Error())
				return
			}
		}

	}
}

// setColor sets the Player color
func (s *WsConnection) setColor(response WsResponse) {
	if response.PayloadSession == s {
		s.Color = response.Message
		s.Lobby.AvailableColors[s.Color] = false
	}

}

// updateColor updates the player color and resets the previous color
func (s *WsConnection) updateColor(response WsResponse) {
	if response.PayloadSession == s {
		if response.Message != s.Color {
			color := s.Color
			s.Color = response.Message
			s.Lobby.AvailableColors[s.Color] = false
			s.Lobby.AvailableColors[color] = true
		}
	}
}

// sendResonse sends the response to the client
func (s *WsConnection) sendResponse(msg string) error {
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
