package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const boardCellsJSONPath = "data/board_cells.json"

type BoardService interface {
	NewBoard()
}

type boardService struct {
	Board
}

// BoardLocation are the spots inside of the board
type BoardCell struct {
	Type       string `json:"type"`
	Suit       string `json:"suit"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	IsCorner   bool   `json:"is_corner"`
	ChipColor  string `json:"chip_color"`
	ChipPlaced bool   `json:"chip_placed"`
}

type BoardCells []BoardCell
type Board map[string]BoardCell

func NewBoardService() BoardService {
	return &boardService{
		Board: make(Board),
	}
}

// NewBoard creates a new game board
func (b *boardService) NewBoard() {
	// cells is going to hold the cells array loaded from file
	var cells BoardCells

	// openn the cells file
	file, err := os.Open(boardCellsJSONPath)
	if err != nil {
		log.Println(err.Error())
	}

	// close the file once the function is executed
	defer file.Close()

	// create a reading from teh file
	r := bufio.NewReader(file)

	// decode the loaded file into a usable struct
	err = json.NewDecoder(r).Decode(&cells)
	if err != nil {
		log.Println(err.Error())
	}

	for _, cell := range cells {
		cellName := fmt.Sprintf("%v_%v", cell.X, cell.Y)
		b.Board[cellName] = cell

		// If the cell is a corner
		if cell.X == 0 && cell.Y == 0 || cell.X == 9 && cell.Y == 0 || cell.X == 0 && cell.Y == 9 || cell.X == 9 && cell.Y == 9 {
			updatedCell := b.Board[cellName]
			updatedCell.IsCorner = true
			updatedCell.ChipPlaced = true
			updatedCell.ChipColor = "Any"

			b.Board[cellName] = updatedCell
		}
	}
}
