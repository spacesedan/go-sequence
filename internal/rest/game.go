package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/spacesedan/go-sequence/internal/services/game"
	"github.com/unrolled/render"
)

type GameHandler struct {
	render *render.Render
	svc    game.GameServiceInterface
}

func NewGameHandler(svc game.GameServiceInterface) *GameHandler {
	render := render.New()
	return &GameHandler{
		svc:    svc,
		render: render,
	}
}

func (g *GameHandler) Register(r *chi.Mux) {
	r.Route("/game", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			g.svc.ShuffleDeck()
			d := g.svc.GetDeck()
			g.render.JSON(w, http.StatusOK, map[string]interface{}{
				"deck": d,
			})
		})
	})
}
