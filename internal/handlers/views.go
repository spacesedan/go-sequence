package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kataras/blocks"
)

type ViewHandler struct {
	Views *blocks.Blocks
}

func NewViewHandler() *ViewHandler {
	views := blocks.New("./views").
		Reload(true)

	err := views.Load()
	if err != nil {
		panic(err)
	}

	return &ViewHandler{
		Views: views,
	}
}

// valid lobby ids are made up of 4 characters that contain any configuration
// of this regex
const lobbyIdRegex string = `[0-9A-Z]{4}`

func (v ViewHandler) Register(r *chi.Mux) {
	r.Get("/", v.HomePage)
	r.Get("/lobby-create", v.CreateLobbyPage)
	r.Get(fmt.Sprintf("/lobby/{lobbyID:%s}", lobbyIdRegex), v.LobbyPage)
}

func (v ViewHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content/Type", "text/html; charset=utf-8")

	data := map[string]interface{}{
		"Title": "Sequence Web",
	}

	err := v.Views.ExecuteTemplate(w, "home", "main", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (v ViewHandler) CreateLobbyPage(w http.ResponseWriter, r *http.Request) {

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
	lobbyID := chi.URLParam(r, "lobbyID")

	fmt.Println(lobbyID)
	data := map[string]interface{}{
		"Title": fmt.Sprintf("Lobby %s", lobbyID),
	}

	err := v.Views.ExecuteTemplate(w, "index", "with_ws", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
