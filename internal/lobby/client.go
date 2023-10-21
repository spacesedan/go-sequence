package lobby

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/components"
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
	Color        string
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
					log.Printf("[ERROR]: %v on session %v", err, s.Username)
				}
				log.Printf("[ERROR]: %v on session %v", err, s.Username)
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
					components.PlayerStatus(response.Message).Render(context.Background(), &b)
					if err := s.broadcastMessage(b.String()); err != nil {
						break
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
					components.ChatMessageSender(response.Message, alt, generateUserAvatar(payload.Username)).Render(context.Background(), &b)
				} else {
					components.ChatMessageReciever(response.Message, alt, generateUserAvatar(payload.Username)).Render(context.Background(), &b)
				}
				if err := s.broadcastMessage(b.String()); err != nil {
					break
				}

				b.Reset()
			case "choose_color":
				response.Message = payload.Message
				if s.Color == "" {
					s.setColor(payload)
				} else {
					s.updateColor(payload)
				}
				if !payload.Enabled {
					fmt.Printf("\n[COLOR TAKEN] %v your a bum\n", s.Username)
					return
				}

				s.syncColors(b)
				b.Reset()
			case "update_colors":
				fmt.Printf("\n[ACTION] %v\n", payload.Action)
			case "left":
				response.Action = "left"
				response.CurrentSession = s
				response.SkipSender = true
				response.Message = fmt.Sprintf("%v left", payload.Username)

				s.LobbyManager.Broadcast <- response
			default:
				break
			}
		case <-ticker.C:
			s.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.Conn.WriteMessage(websocket.PingMessage, []byte("PING")); err != nil {
				return
			}
		}

	}
}

func (s *WsConnection) syncColors(b bytes.Buffer) {
	for color, available := range s.Lobby.AvailableColors {
		if available {
			components.PlayerColorComponent(color).Render(context.Background(), &b)
		} else {
			components.PlayerColorUnavailableComponent(color).Render(context.Background(), &b)
		}
		s.broadcastMessage(b.String())
		b.Reset()
	}


}

// setColor sets the Player color
func (s *WsConnection) setColor(payload WsPayload) {
	if payload.Username == s.Username {
		s.Color = payload.Message
		s.Lobby.AvailableColors[s.Color] = false
	}

}

// updateColor updates the player color and resets the previous color
func (s *WsConnection) updateColor(payload WsPayload) {
	if payload.Username == s.Username {
		if payload.Message != s.Color {
			color := s.Color
			s.Color = payload.Message
			s.Lobby.AvailableColors[s.Color] = false
			s.Lobby.AvailableColors[color] = true
		}
	}
}

func (s *WsConnection) broadcastMessage(msg string) error {
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
