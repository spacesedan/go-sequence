package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Pallinder/go-randomdata"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/lobby"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {

		return true
	},
}

type LobbyHandler struct {
	LobbyManager *lobby.LobbyManager
	sm           *scs.SessionManager
	logger       *slog.Logger
}

func NewLobbyHandler(lm *lobby.LobbyManager, l *slog.Logger, sm *scs.SessionManager) *LobbyHandler {
	go lm.ListenToWsChannel()
	return &LobbyHandler{
		LobbyManager: lm,
		logger:       l,
		sm:           sm,
	}
}

func (lh *LobbyHandler) Register(m *chi.Mux) {
	m.Route("/lobby", func(r chi.Router) {
		r.HandleFunc("/ws", lh.Serve)
		r.Get("/generate_username", lh.GenerateUsername)
		r.Post("/create_lobby", lh.CreateGameLobby)
	})
}

type Task struct {
	Action  string `json:"action"`
	Subject string `json:"subject"`
}

func (lm *LobbyHandler) Serve(w http.ResponseWriter, r *http.Request) {
	lm.logger.Info("Connected to socket")

	lobbyId := r.URL.Query().Get("lobbyID")

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
	response.LobbyID = lobbyId
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

func (lm *LobbyHandler) CreateGameLobby(w http.ResponseWriter, r *http.Request) {

	// get the settings
	numberOfPlayers := r.FormValue("num_of_players")
	maxHandSize := r.FormValue("max_hand_size")

	// create the lobby
	lobbyId :=lm.LobbyManager.CreateLobby(lobby.Settings{
		NumOfPlayers: numberOfPlayers,
		MaxHandSize: maxHandSize,
	})


	// Redirect to the lobby page after it has been created
	w.Header().Set("HX-Redirect", fmt.Sprintf("/lobby/%s", lobbyId))

}

// GenerateUsername generates a username and stores the value in the session.
func (lm *LobbyHandler) GenerateUsername(w http.ResponseWriter, r *http.Request) {
	randomName := randomdata.SillyName()
	randomNumber := randomdata.Number(42069)

	// construct the username
	userName := fmt.Sprintf("%d%s", randomNumber, randomName)

	cookie := http.Cookie{
		Name:     "username",
		Value:    userName,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, &cookie)
	// add the username to the session
	lm.sm.Put(r.Context(), fmt.Sprintf("username:%s", userName), userName)

	// send the response back to the client
	render.Text(w, http.StatusOK, userName)
}
