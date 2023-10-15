package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/lobby"
	"github.com/spacesedan/go-sequence/internal/partials"
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
		r.Post("/create", lh.CreateGameLobby)
		r.Post("/join", lh.JoinLobby)

		lobbyHTMXGroup := r.Group(nil)
		lobbyHTMXGroup.Route("/view", func(r chi.Router) {
			r.Get("/toast/prompt-username", lh.PromptUserToGenerateUsername)
			r.Get("/modal/join-lobby", lh.SendJoinLobbyModal)
		})
	})
}

func (lm *LobbyHandler) Serve(w http.ResponseWriter, r *http.Request) {
	lm.logger.Info("Connected to socket")

	username, err := getUsernameFromCookie(r)
	if err != nil {
		w.Header().Set("HX-Redirect", "/")
		render.Text(w, http.StatusSeeOther, "")
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		lm.logger.Error("Something went wrong",
			slog.String("err", err.Error()))
		return
	}

	session := &lobby.WsConnection{Conn: ws, Username: username, LobbyManager: lm.LobbyManager, Send: make(chan lobby.WsPayload)}
	session.LobbyManager.RegisterChan <- session

	go session.WritePump()
	go session.ReadPump()

}

func (lm *LobbyHandler) CreateGameLobby(w http.ResponseWriter, r *http.Request) {

	// get the settings
	numberOfPlayers := r.FormValue("num_of_players")
	maxHandSize := r.FormValue("max_hand_size")

	// create the lobby
	lobbyId := lm.LobbyManager.CreateLobby(lobby.Settings{
		NumOfPlayers: numberOfPlayers,
		MaxHandSize:  maxHandSize,
	})

	// Redirect to the lobby page after it has been created
	w.Header().Set("HX-Redirect", fmt.Sprintf("/lobby/%s", lobbyId))

}

func (lm *LobbyHandler) JoinLobby(w http.ResponseWriter, r *http.Request) {
	username, _ := getUsernameFromCookie(r)
	lobbyID := r.FormValue("lobby-id")

	exists := lm.LobbyManager.LobbyExists(lobbyID, username)
	if !exists {
		content := "make sure you entered a valid lobby id"
		topic := "Lobby not found"
		partials.ToastComponent(topic, content).Render(r.Context(), w)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/lobby/%v", lobbyID))
	render.Text(w, http.StatusSeeOther, "")

}

// GenerateUsername generates a username and stores the value in the session.
func (lm *LobbyHandler) GenerateUsername(w http.ResponseWriter, r *http.Request) {
	userName, userCookie := generateUserCookie()

	http.SetCookie(w, userCookie)
	// add the username to the session
	lm.sm.Put(r.Context(), fmt.Sprintf("username:%s", userName), userName)

	w.Header().Set("HX-Redirect", "/")

	// send the response back to the client
	render.Text(w, http.StatusSeeOther, "")
}

func (lm *LobbyHandler) PromptUserToGenerateUsername(w http.ResponseWriter, r *http.Request) {
	topic := "Generate a username first."
	content := `this site work better when you have a username click on "generate username" to get yours`
	partials.ToastComponent(topic, content).Render(r.Context(), w)

}

func (lm *LobbyHandler) SendJoinLobbyModal(w http.ResponseWriter, r *http.Request) {
	partials.JoinLobbyModal().Render(r.Context(), w)

}
