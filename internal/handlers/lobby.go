package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/lobby"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type LobbyHandler struct {
	LobbyManager *lobby.LobbyManager
	logger       *slog.Logger
}

func NewLobbyHandler(lm *lobby.LobbyManager, l *slog.Logger) *LobbyHandler {
	go lm.ListenToWsChannel()
	return &LobbyHandler{
		LobbyManager: lm,
		logger:       l,
	}
}

func (lh *LobbyHandler) Register(m *chi.Mux) {
	m.HandleFunc("/ws", lh.Serve)
}

func (lm *LobbyHandler) Serve(w http.ResponseWriter, r *http.Request) {
	lm.logger.Info("Connected to socker")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		lm.logger.Error("Something went wrong",
			slog.String("err", err.Error()))
		return
	}

	var response lobby.WsJsonResponse
	response.Action = "connected"
	response.Message = `<p id="wsStatus">Welcome to the lobby</p>`

	err = ws.WriteMessage(websocket.TextMessage, []byte(response.Message))
	if err != nil {
		lm.logger.Error("Something when trying to send a message to the client",
		slog.String("err", err.Error()))
	}

	conn := lobby.WsConnection{Conn: ws}
	lm.LobbyManager.Clients[conn] = ""

	go lm.LobbyManager.ListenForWs(&conn)
}
