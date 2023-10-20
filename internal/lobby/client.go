package lobby

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/partials"
	"github.com/spacesedan/go-sequence/internal/views"
)

type Color string

const (
	ColorRed  Color = "red"
	ColorGeen Color = "green"
	ColorBlue Color = "blue"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type WsConnection struct {
	LobbyManager *LobbyManager
	Lobby        *GameLobby
	Conn         *websocket.Conn
	Color        Color
	Username     string
	LobbyID      string
	Send         chan WsPayload
}

func (s *WsConnection) ReadPump() {
	defer func() {
		s.LobbyManager.UnregisterChan <- s
		s.Conn.Close()
	}()
	s.Conn.SetReadLimit(maxMessageSize)
	s.Conn.SetReadDeadline(time.Now().Add(pongWait))
	s.Conn.SetPongHandler(func(string) error { s.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	var payload WsPayload

	for {
		err := s.Conn.ReadJSON(&payload)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[ERROR]: %v", err)
				}
				log.Printf("[ERROR]: %v", err)
			}
			break
		}

		s.LobbyManager.WsChan <- payload
	}

}

func (s *WsConnection) WritePump() {
	var response WsResponse
	var b bytes.Buffer
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.Conn.Close()
	}()

	for {
		select {
		case payload, ok := <-s.Send:
			if !ok {
				s.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			switch payload.Action {
			case "join_lobby":
				response.Message = fmt.Sprintf("%v joined", payload.Username)

				if payload.Username != s.Username {
					partials.PlayerStatus(response.Message).Render(context.Background(), &b)
					if err := s.broadcastMessage(b.String()); err != nil {
						s.Conn.Close()
					}
				}
				b.Reset()
			case "chat_message":
				response.Action = "new_message"
				response.CurrentSession = s
				response.SkipSender = false
				response.Message = payload.Message

				alt := fmt.Sprintf("Avatar for %v", payload.Username)

				if payload.Username == s.Username {
					partials.ChatMessageSender(response.Message, alt, generateUserAvatar(payload.Username)).Render(context.Background(), &b)
				} else {
					partials.ChatMessageReciever(response.Message, alt, generateUserAvatar(payload.Username)).Render(context.Background(), &b)
				}
				if err := s.broadcastMessage(b.String()); err != nil {
					s.Conn.Close()
				}

				b.Reset()
			case "choose_color":
				response.Action = "update_colors"
				response.CurrentSession = s
				response.SkipSender = false
				response.Message = payload.Message

				// if send and no color set, set the player color
				if payload.Enabled {
					views.PlayerColorUnavailableComponent(response.Message).Render(context.Background(), &b)
				} else {
					views.PlayerColorComponent(response.Message).Render(context.Background(), &b)
				}
				fmt.Printf("\n%v\n%v\n%v\n", payload.Enabled, payload.Message, b.String())
				if err := s.broadcastMessage(b.String()); err != nil {
					s.Conn.Close()
				}

				b.Reset()
			case "left":
				response.Action = "left"
				response.CurrentSession = s
				response.SkipSender = true
				response.Message = fmt.Sprintf("%v left", payload.Username)

				s.LobbyManager.Broadcast <- response
			default:
				fmt.Printf("%#v", payload.Action)
				s.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("ACTION: %v", payload.Action)))
			}
		case <-ticker.C:
			s.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.Conn.WriteMessage(websocket.PingMessage, []byte("PING")); err != nil {
				return
			}
		}

	}
}

func (s *WsConnection) setColor(payload WsPayload) error {
    if s.Color == "" {
        s.Color = Color(payload.Message)
    }

    return nil
}
func (s *WsConnection) updateColor() error{
    return nil
}

func (s *WsConnection) broadcastMessage(msg string) error {
	return s.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func generateUserAvatar(username string) string {
	// https://ui-avatars.com/api/?name=poop&amp;size=32&amp;rounded=true
	u := url.URL{
		Scheme: "https",
		Host:   "ui-avatars.com",
		Path:   "api",
	}

	q := u.Query()
	q.Set("name", username)
	q.Set("size", "32")
	q.Set("rounded", "true")

	u.RawQuery = q.Encode()

	return u.String()
}
