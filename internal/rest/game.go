package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/spacesedan/go-sequence/internal/services/game"
	"github.com/unrolled/render"
)

type GameHandler struct {
	render *render.Render
	svc    game.GameServiceInterface
}

func NewGameHandler(svc game.GameServiceInterface) *GameHandler {
	return &GameHandler{
		svc: svc,
	}
}

func (g *GameHandler) Register(r *chi.Mux) {
	r.Route("/game", func(r chi.Router) {
		r.Get("/", g.GetDeck)
		r.Get("/p", g.GetPlayers)
        r.Get("/deal", g.DealCards)
		r.Post("/p/add", g.AddPlayer)
		r.Delete("/p/remove/{id}", g.RemovePlayer)
	})
}

func (g GameHandler) GetDeck(w http.ResponseWriter, r *http.Request) {
	d := g.svc.GetDeck()
	rndr.JSON(w, http.StatusOK, Response{
		"deck": d,
	})
}

func (g GameHandler) GetPlayers(w http.ResponseWriter, r *http.Request) {
	players := g.svc.GetPlayers()
	rndr.JSON(w, http.StatusOK, Response{
		"players": players,
	})
}

type PlayerRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (g GameHandler) AddPlayer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req PlayerRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	g.svc.AddPlayer(&game.Player{
		Name:  req.Name,
		Color: req.Color,
	})

	rndr.JSON(w, http.StatusOK, Response{
		"player": req,
	})

}

func (g GameHandler) RemovePlayer(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    fmt.Println("Deleting player", id)
    g.svc.RemovePlayer(uuid.MustParse(id))
    rndr.Text(w, http.StatusOK, id)
}


func (g GameHandler) DealCards(w http.ResponseWriter, r *http.Request) {
    g.svc.DealCards(7)
    rndr.JSON(w, http.StatusOK, Response{
        "players": g.svc.GetPlayers(),
    })
}
