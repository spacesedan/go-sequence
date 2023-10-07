package handlers

import (
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

func (v ViewHandler) Register(r *chi.Mux) {
	r.Get("/", v.HomePage)
	r.Route("/lobby", func(r chi.Router) {
        r.Get("/create", v.CreateLobbyPage)
	})
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

	err := v.Views.ExecuteTemplate(w, "create_lobby", "main", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
