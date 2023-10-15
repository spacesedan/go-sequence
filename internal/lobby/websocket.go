package lobby

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/partials"
)

func (s *WsConnection) ReadPump() {
	defer func() {
		s.LobbyManager.UnregisterChan <- s
		s.Conn.Close()
	}()

	var payload WsPayload

	for {
		err := s.Conn.ReadJSON(&payload)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ERROR]: %v", err)
			}
			break
		}

		log.Printf("[INFO] sending: %v", payload)
		s.LobbyManager.WsChan <- payload

	}

}

func (s *WsConnection) WritePump() {
	defer func() {
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
				s.Conn.WriteMessage(websocket.TextMessage, []byte("Joined"))
			case "chat-message":
				b := bytes.NewBuffer([]byte{})
                err := partials.ChatMessage(payload.Message).Render(context.Background(), b)
                if err != nil {
                    log.Println(err)
                }
                s.Conn.WriteMessage(websocket.TextMessage, []byte(b.String()))
			default:
				fmt.Printf("%#v", payload.Action)
				s.Conn.WriteMessage(websocket.TextMessage, []byte("POOP"))
			}
		}
	}
}
