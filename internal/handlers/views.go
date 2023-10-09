package handlers

import (
	"fmt"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/kataras/blocks"
)

type ViewHandler struct {
	Views *blocks.Blocks
	sm    *scs.SessionManager
}

func NewViewHandler(sm *scs.SessionManager) *ViewHandler {
	views := blocks.New("./views").
		Reload(true)

	err := views.Load()
	if err != nil {
		panic(err)
	}

	return &ViewHandler{
		sm:    sm,
		Views: views,
	}
}

// valid lobby ids are made up of 4 characters that contain any configuration
// of this regex
const lobbyIdRegex string = `[0-9A-Z]{4}`

func (v ViewHandler) Register(r *chi.Mux) {
	r.Get("/", v.IndexPage)
	r.Get("/lobby-create", v.CreateLobbyPage)
	r.Get(fmt.Sprintf("/lobby/{lobbyID:%s}", lobbyIdRegex), v.LobbyPage)
}

func (v ViewHandler) IndexPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content/Type", "text/html; charset=utf-8")

	var userName string
	userName = v.sm.GetString(r.Context(), "username")

	fmt.Println("username", userName)


	data := map[string]interface{}{
		"Title": "Sequence Web",
		"UserName": userName,
	}

	err := v.Views.ExecuteTemplate(w, "index", "main", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (v ViewHandler) CreateLobbyPage(w http.ResponseWriter, r *http.Request) {
    username := v.sm.GetString(r.Context(), "username")

    fmt.Println("username", username)

	w.Header().Set("Content/Type", "text/html; charset=utf-8")

	data := map[string]interface{}{
		"Title": "Create a new lobby",
	}

	err := v.Views.ExecuteTemplate(w, "create_lobby", "with_ws", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (v ViewHandler) LobbyPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Will this print")
	lobbyID := chi.URLParam(r, "lobbyID")

	data := map[string]interface{}{
		"Title":   fmt.Sprintf("Lobby %s", lobbyID),
		"LobbyID": lobbyID}

	err := v.Views.ExecuteTemplate(w, "lobby", "with_ws", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
