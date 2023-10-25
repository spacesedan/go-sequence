package lobby

import (
	"bytes"
	"context"
	"errors"
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
	Send         chan WsResponse
	ErrorChan    chan error
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
		}

		lobbyExists := s.LobbyManager.LobbyExists(s.LobbyID)
		if lobbyExists {
			payload.SenderSession = s
			s.Lobby.PayloadChan <- payload
		} else {
			s.ErrorChan <- errors.New("Lobby does not exist")
		}
	}

}

func (s *WsConnection) WritePump() {
	var b bytes.Buffer
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.Conn.Close()
	}()

writeLoop:
	for {
		select {
		case response, ok := <-s.Send:
			if !ok {
				s.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			switch response.Action {

			case "join_lobby":
				if response.PayloadSession != s {
					components.PlayerStatus(response.Message).Render(context.Background(), &b)
					if err := s.broadcastMessage(b.String()); err != nil {
						s.ErrorChan <- err
					}
				}
				b.Reset()

			case "new_chat_message":
				alt := fmt.Sprintf("avatar image for %v", s.Username)

				if response.PayloadSession == s {
					components.ChatMessageSender(
						response.Message,
						alt,
						generateUserAvatar(response.PayloadSession.Username)).
						Render(context.Background(), &b)
				} else {
					components.ChatMessageReciever(
						response.Message,
						alt,
						generateUserAvatar(response.PayloadSession.Username)).
						Render(context.Background(), &b)
				}
				if err := s.broadcastMessage(b.String()); err != nil {
					s.ErrorChan <- err
				}

				b.Reset()

			case "choose_color":
				if s.Color == "" {
					s.setColor(response)
				} else {
					s.updateColor(response)
				}

				s.Lobby.PayloadChan <- WsPayload{
					Action:        "sync_colors",
					Message:       "",
					SenderSession: s,
				}

			case "sync_colors":
				s.broadcastMessage(response.Message)
			case "left":
				components.PlayerStatus(response.Message).Render(context.Background(), &b)
				if err := s.broadcastMessage(b.String()); err != nil {
					s.ErrorChan <- err
				}

				b.Reset()

			default:
				s.ErrorChan <- errors.New(fmt.Sprintf("Something unexpected: %v", response))
			}
		case err := <-s.ErrorChan:
			fmt.Println("[ERROR] ", err)
			break writeLoop

		case <-ticker.C:
			s.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.Conn.WriteMessage(websocket.PingMessage, []byte("")); err != nil {
				s.ErrorChan <- err
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
