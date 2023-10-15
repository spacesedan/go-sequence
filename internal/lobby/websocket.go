package lobby

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/partials"
)

func (lm *LobbyManager) ListenForWs(conn *WsConnection) {
	defer func() {
		if r := recover(); r != nil {
			lm.logger.Error("[Error] Attempting to recover", slog.Any("reason", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.Conn.ReadJSON(&payload)
		if err != nil {
			// ... just ignore it
		} else {
			payload.Conn = *conn
			lm.WsChan <- payload
		}

	}
}

func (lm *LobbyManager) ListenToWsChannel() {
	for {
		e := <-lm.WsChan
		switch e.Action {
		case "join_lobby":

			fmt.Println("JOINING")
			fmt.Println("\nNUMBER OF LOBBIES\n", len(lm.Lobbies))
            for l, lobby := range lm.Lobbies {
                fmt.Printf("Lobby ID: %v\nLobby: %#v\nNumbe of Clients: %v\nConnection: %v\n", l, lobby.Clients, len(lm.Clients), e.Conn)
            }
            lm.Lobbies[e.LobbyID].Clients[e.Username] = e.Conn

            fmt.Println("Sent")
			lm.Lobbies[e.LobbyID].GameChan <- e
		case "chat-message":
			lm.Lobbies[e.LobbyID].GameChan <- e
		default:
			for _, c := range lm.Clients {
				c.Conn.WriteMessage(websocket.TextMessage, []byte("Poop join"))
			}
		}
	}
}

func (lm *LobbyManager) ListenToLobbyWsChan() {
	for {
		for lobbyId, lobby := range lm.Lobbies {
			e := <-lobby.GameChan
            if _, ok := lobby.Clients[e.Username]; !ok {
                fmt.Println("Not ok")
            } else {
                fmt.Println("Ok")
            }
			switch e.Action {
			case "join_lobby":
				lm.logger.Info("Joined", slog.String("lobby", lobbyId))
				for username, conn := range lobby.Clients {
					fmt.Println(username)

                    err :=conn.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("joined lobby:%s", lobbyId)))
                    if err != nil {
                        fmt.Printf("[ERROR] %v", err.Error())
                    }
				}
			case "chat-message":
				fmt.Println("Message")
				for _, client := range lobby.Clients {
					client.Conn.WriteMessage(websocket.TextMessage, []byte("POOPING"))
				}
			}
		}
	}
}

func (lm *LobbyManager) handleChatMessage(payload WsPayload) {

	if payload.Message == "" {
		return
	}

	for _, client := range lm.Clients {
		w, _ := client.Conn.NextWriter(websocket.TextMessage)
		err := partials.ChatMessage(payload.Message).Render(context.Background(), w)
		if err != nil {
			fmt.Println(err.Error())
		}

		w.Close()
	}

}
