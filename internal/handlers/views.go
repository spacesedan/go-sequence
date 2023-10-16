package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/spacesedan/go-sequence/internal/lobby"
	"github.com/spacesedan/go-sequence/internal/views"
)

type ViewHandler struct {
	LobbyManager *lobby.LobbyManager
	sm           *scs.SessionManager
}

func NewViewHandler(sm *scs.SessionManager, lm *lobby.LobbyManager) *ViewHandler {
	return &ViewHandler{
		sm:           sm,
		LobbyManager: lm,
	}
}

// valid lobby ids are made up of 4 characters that contain any configuration
// of this regex
const lobbyIdRegex string = `[0-9A-Z]{4}`

func (v ViewHandler) Register(r *chi.Mux) {
	r.Get("/", v.IndexPage)

	lobbyGroup := r.Group(nil)
	lobbyGroup.Use(CheckUsernameCookie)
	lobbyGroup.Get("/lobby-create", v.CreateLobbyPage)
	lobbyGroup.Get(fmt.Sprintf("/lobby/{lobbyID:%s}", lobbyIdRegex), v.LobbyPage)
}

func (v ViewHandler) IndexPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content/Type", "text/html; charset=utf-8")

	userCookie, _ := r.Cookie("username")

	var userName string

	if userCookie != nil {
		userName = userCookie.Value
	}

	err := views.
		MainLayout("Sequence Web", views.IndexPage(userName)).
		Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateLobbyPage
func (v ViewHandler) CreateLobbyPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content/Type", "text/html; charset=utf-8")

	err := views.
		MainLayout("Sequence Web", views.CreateLobbyPage()).
		Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (v ViewHandler) LobbyPage(w http.ResponseWriter, r *http.Request) {
	username, err := getUsernameFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	lobbyID := chi.URLParam(r, "lobbyID")
    lobbyID = strings.Trim(lobbyID, " ")
	exists := v.LobbyManager.LobbyExists(lobbyID, username)

	if !exists {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

    connectionUrl := createWebsocketConnectionString(lobbyID)

	err = views.
		MainLayoutWithWs(fmt.Sprintf("Lobby %s", lobbyID), views.LobbyPage(connectionUrl, lobbyID, username)).
		Render(context.Background(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
