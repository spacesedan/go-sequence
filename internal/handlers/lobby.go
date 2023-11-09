package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/spacesedan/go-sequence/internal/views/components"
	"github.com/spacesedan/go-sequence/internal/game"
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
	logger       *slog.Logger
}

func NewLobbyHandler(lm *lobby.LobbyManager, l *slog.Logger) *LobbyHandler {
	return &LobbyHandler{
		LobbyManager: lm,
		logger:       l,
	}
}

func (lh *LobbyHandler) Register(m *chi.Mux) {
	m.Route("/lobby", func(r chi.Router) {
		r.HandleFunc("/ws", lh.Serve)
		r.Get("/generate_username", lh.handleGenerateUsername)
		r.Post("/create", lh.handleCreateGameLobby)
		r.Post("/join", lh.handleJoinLobby)

		lobbyHTMXGroup := r.Group(nil)
		lobbyHTMXGroup.Route("/view", func(r chi.Router) {
			r.Get("/toast/prompt-username", lh.handlePromptUserToGenerateUsername)
			r.Get("/modal/join-lobby", lh.handleSendJoinLobbyModal)
		})
	})
}

func (lm *LobbyHandler) Serve(w http.ResponseWriter, r *http.Request) {
	lm.logger.Info("lobbyHandler.Serve", slog.Group("connected to socker"))

	lobbyId := r.URL.Query().Get("lobby-id")

	username, err := getUsernameFromCookie(r)
	if err != nil {
		w.Header().Set("HX-Redirect", "/")
		render.Text(w, http.StatusSeeOther, "")
		return
	}

	l, ok := lm.LobbyManager.LobbyExists(lobbyId)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		render.Text(w, http.StatusSeeOther, "")
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	fmt.Printf("%#v\n", l.Settings)

	ok = l.HasPlayer(username)
	session := lobby.NewWsClient(ws, lm.LobbyManager, l, lm.logger, username, l.ID)

	l.RegisterChan <- session

    go session.SubscribeToLobby()
    go session.ReadPump()
	// go session.WritePump()


}

func (lm *LobbyHandler) handleCreateGameLobby(w http.ResponseWriter, r *http.Request) {
	// get the settings
	numberOfPlayersString := r.FormValue("num_of_players")
	maxHandSizeString := r.FormValue("max_hand_size")

	numOfPlayers, err := strconv.Atoi(numberOfPlayersString)
	if err != nil {
		return
	}
	maxHandSize, err := strconv.Atoi(maxHandSizeString)
	if err != nil {
		return
	}

	// create the lobby
	lobbyId := lm.LobbyManager.NewLobby(game.Settings{
		NumOfPlayers: numOfPlayers,
		MaxHandSize:  maxHandSize,
	})

	lm.logger.Info("New game lobby", slog.String("lobby-id", lobbyId))

	// Redirect to the lobby page after it has been created
	w.Header().Set("HX-Redirect", fmt.Sprintf("/lobby/%s", lobbyId))

}

func (lm *LobbyHandler) handleJoinLobby(w http.ResponseWriter, r *http.Request) {
	lobbyID := r.FormValue("lobby-id")

	l, exists := lm.LobbyManager.LobbyExists(lobbyID)
	if !exists {
		content := "make sure you entered a valid lobby id"
		topic := "Lobby not found"
		components.ToastComponent(topic, content).Render(r.Context(), w)
		return
	}

	if len(l.Players) == l.Settings.NumOfPlayers {
		topic := "Lobby full"
		content := "cannot join lobby, already at max capacity"
		components.ToastComponent(topic, content).Render(r.Context(), w)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/lobby/%v", lobbyID))
	render.Text(w, http.StatusSeeOther, "")

}

// GenerateUsername generates a username and stores the value in the session.
func (lm *LobbyHandler) handleGenerateUsername(w http.ResponseWriter, r *http.Request) {
	_, userCookie := generateUserCookie()

	http.SetCookie(w, userCookie)
	// add the username to the session

	w.Header().Set("HX-Redirect", "/")

	// send the response back to the client
	render.Text(w, http.StatusSeeOther, "")
}

func (lm *LobbyHandler) handlePromptUserToGenerateUsername(w http.ResponseWriter, r *http.Request) {
	topic := "Generate a username first."
	content := `this site work better when you have a username click on "generate username" to get yours`
	components.ToastComponent(topic, content).Render(r.Context(), w)

}

func (lm *LobbyHandler) handleSendJoinLobbyModal(w http.ResponseWriter, r *http.Request) {
	components.JoinLobbyModal().Render(r.Context(), w)

}
