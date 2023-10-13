package lobby

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/partials"
)

func (lm *LobbyManager) ListenToWsChannel() {
	for {
		e := <-lm.WsChan
		switch e.Action {
		case "create_lobby":
			lm.logger.Info("Creating new lobby")

		case "join_lobby":
			for _, c := range lm.Clients {
				c.WriteMessage(websocket.TextMessage, []byte("Poop join"))
			}

		case "leave_lobby":
			lm.logger.Info("Leaving lobby")

		default:
		}
	}
}

func (lm *LobbyManager) ListenForWs(conn *WsConnection, lobbyId, username string) {
	lm.logger.Info(lobbyId)
	defer func() {
		conn.WriteMessage(websocket.TextMessage, []byte("Closing connection"))
		delete(lm.Clients, username)
		if r := recover(); r != nil {
			lm.logger.Error("Error: Attempting to recover", slog.Any("reason", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// ... just ignore it
		} else {
			payload.Conn = *conn
			switch lobbyId {
			case "":
				lm.WsChan <- payload
			default:
				lm.Lobbies[lobbyId].GameChan <- payload
			}

		}
	}
}

func (lm *LobbyManager) ListenToLobbyWsChan(lobbyId string) {
	lm.logger.Info("Listening to lobbyChan")
	lm.logger.Info("Lobby info", slog.String("lobbyId", lobbyId))
	for {
		lobby := lm.Lobbies[lobbyId]
		e := <-lobby.GameChan
        switch(e.Action){

        }
        fmt.Printf("%#v", e)
		w, err := e.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
        err = partials.ChatMessage(e.Message).Render(context.Background(), w)
        if err != nil {
            lm.logger.Error(err.Error())
        }

	}
}
