package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/lobby"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		fmt.Println("URL", r.URL.RawQuery)

		return true
	},
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
	m.HandleFunc("/lobby/ws", lh.Serve)
}

type Task struct {
	Action  string `json:"action"`
	Subject string `json:"subject"`
}

func (lm *LobbyHandler) Serve(w http.ResponseWriter, r *http.Request) {
	lm.logger.Info("Connected to socket")

	action := r.URL.Query().Get("action")
	subject := r.URL.Query().Get("subject")
    lobbyId := r.URL.Query().Get("lobbyID")

	switch subject {
	case "lobby":
		fmt.Println("lobby", action, lobbyId)
	}

	// if i add any infomration to the url i can parse it out here and
	// uses it however i want

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		lm.logger.Error("Something went wrong",
			slog.String("err", err.Error()))
		return
	}

	var response lobby.WsJsonResponse
	response.Action = "connected"
	response.Message = `<h1 id="wsStatus">Welcome to Go-Sequence</h1>`

	err = ws.WriteMessage(websocket.TextMessage, []byte(response.Message))
	if err != nil {
		lm.logger.Error("Something when trying to send a message to the client",
			slog.String("err", err.Error()))
	}

	conn := lobby.WsConnection{Conn: ws}
	lm.LobbyManager.Clients[conn] = true

	go lm.LobbyManager.ListenForWs(&conn)
}
