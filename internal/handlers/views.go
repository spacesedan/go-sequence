package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kataras/blocks"
)

type ViewHandler struct {
	Views     *blocks.Blocks
}


func NewViewHandler() *ViewHandler {
	views := blocks.New("./views").
		Reload(true)



	err := views.Load()
	if err != nil {
		panic(err)
	}

	return &ViewHandler{
		Views:     views,
	}
}

func (v ViewHandler) Register(r *chi.Mux) {
	r.Get("/", v.HomePage)
}

func (v ViewHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content/Type", "text/html; charset=utf-8")

	data := map[string]interface{}{
		"Title": "Sequence Web",
	}

	err := v.Views.ExecuteTemplate(w, "index", "main", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
