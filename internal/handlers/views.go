package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kataras/blocks"
	"github.com/spacesedan/go-sequence/internal/services"
)

type ViewHandler struct {
	Views     *blocks.Blocks
	GameCells Cells
}

type Cells []*services.BoardCell

func NewViewHandler(gb services.Board) *ViewHandler {
	views := blocks.New("./views").
		Reload(true)

	var BoardCells Cells

	for i := 0; i < services.BoardSize; i++ {
		for j := 0; j < services.BoardSize; j++ {
			BoardCells = append(BoardCells, gb[i][j])
		}
	}

	err := views.Load()
	if err != nil {
		panic(err)
	}

	return &ViewHandler{
		Views:     views,
		GameCells: BoardCells,
	}
}

func (v ViewHandler) Register(r *chi.Mux) {
	r.Get("/", v.HomePage)
}

func (v ViewHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content/Type", "text/html; charset=utf-8")

	data := map[string]interface{}{
		"Title": "Sequence Web",
        "Cells": v.GameCells,
	}

	err := v.Views.ExecuteTemplate(w, "index", "main", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
