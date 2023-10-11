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

	data := map[string]interface{}{
		"Title":    "Sequence Web",
		"Username": userName,
	}

	err := v.Views.ExecuteTemplate(w, "index", "main", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateLobbyPage
func (v ViewHandler) CreateLobbyPage(w http.ResponseWriter, r *http.Request) {
	username, err := getUsernameFromCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Println(username)

	w.Header().Set("Content/Type", "text/html; charset=utf-8")

	data := map[string]interface{}{
		"Title": "Create a new lobby",
		"Username": username,
	}

	err = v.Views.ExecuteTemplate(w, "create_lobby", "with_ws", data)
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

	data := map[string]interface{}{
		"Title":   fmt.Sprintf("Lobby %s", lobbyID),
		"LobbyID": lobbyID,
		"Username": username,

	}

	err = v.Views.ExecuteTemplate(w, "lobby", "with_ws", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
