package lobby

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/partials"
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
	b := bytes.Buffer{}
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
				if s.LobbyID == payload.LobbyID {
					if s.Username != payload.Username {
						err := partials.PlayerStatus(fmt.Sprintf("%v joined", payload.Username)).Render(context.Background(), &b)
						if err != nil {
							log.Println(err)
						}
						s.Conn.WriteMessage(websocket.TextMessage, []byte(b.String()))
					}
				}
				b.Reset()
			case "chat_message":
				if s.LobbyID == payload.LobbyID {
					payload.Message = strings.TrimSpace(payload.Message)
					if payload.Message == "" {
						continue
					}
					if s.Username == payload.Username {
						err := partials.ChatMessageSender(payload.Message, fmt.Sprintf("avatar for user %v", payload.Username), generateUserAvatar(payload.Username)).Render(context.Background(), &b)
						if err != nil {
							log.Println(err)
						}
					} else {
						err := partials.ChatMessageReciever(payload.Message, fmt.Sprintf("avatar for user %v", s.Username), generateUserAvatar(payload.Username)).Render(context.Background(), &b)
						if err != nil {
							log.Println(err)
						}
					}
					s.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Sending to lobby: %v", b.String())))
				}
				b.Reset()
            case "choose_color":
                s.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("color selected %v", payload.Message)))
			case "left":
				s.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%v left", payload.Username)))
				err := partials.PlayerStatus(fmt.Sprintf("%v left", payload.Username)).Render(context.Background(), &b)
				if err != nil {
					log.Println(err)
				}
				s.Conn.WriteMessage(websocket.TextMessage, []byte(b.String()))
				b.Reset()
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
