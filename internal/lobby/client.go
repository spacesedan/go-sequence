package lobby

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
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
				response.Action = "joined"
				response.CurrentSession = s
                response.SkipSender = true
				response.Message = fmt.Sprintf("%v joined", payload.Username)
				s.LobbyManager.Broadcast <- response
			case "chat_message":
				response.Action = "new_message"
				response.CurrentSession = s
				response.SkipSender = false
				response.Message = payload.Message

				s.LobbyManager.Broadcast <- response
			case "choose_color":
                if s.Color != "" {
                    continue
                }
				s.Color = Color(payload.Message)

				response.Action = "update_colors"
				response.CurrentSession = s
				response.SkipSender = false
				response.Message = payload.Message

				s.LobbyManager.Broadcast <- response

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
