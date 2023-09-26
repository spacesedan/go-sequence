package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

const boardCellsJSONPath = "data/board_cells.json"

type BoardService interface {
	NewBoard()
	GetBoard() Board
	AddPlayerChip(p Player, c Card, pos Position)
	RemovePlayerChip(pos Position)
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
	ChipPlaced bool   `json:"chip_placed"`
	ChipColor  string `json:"chip_color"`
}

type BoardCells []BoardCell

type Board map[string]*BoardCell

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

		// If the cell is a corner
		if cell.X == 0 && cell.Y == 0 || cell.X == 9 && cell.Y == 0 || cell.X == 0 && cell.Y == 9 || cell.X == 9 && cell.Y == 9 {
			cellName := fmt.Sprintf("Corner_%v_%v", cell.X, cell.Y)

			// set the values for a corner
			cell.IsCorner = true
			cell.ChipPlaced = true
			cell.ChipColor = "Any"

			cellPointer := newBoardCell(cell)

			b.Board[cellName] = cellPointer

		} else {
			cellName := fmt.Sprintf("%s_%s_%v_%v", cell.Suit, cell.Type, cell.X, cell.Y)

			cellPointer := newBoardCell(cell)

			b.Board[cellName] = cellPointer
		}

	}

}

func (b boardService) GetBoard() Board {
	return b.Board
}

type Position struct {
	X int
	Y int
}

func (b boardService) AddPlayerChip(player Player, card Card, pos Position) {
	cellName := fmt.Sprintf("%s_%s_%d_%d", card.Suit, card.Type, pos.X, pos.Y)

	cell := b.Board[cellName]

	// check to see if the cell is already occupied
	if cell.ChipPlaced {
		// add more information later
		log.Println("Cell taken")
		return
	}

	cell.ChipColor = player.Color
	cell.ChipPlaced = true

}

func (b boardService) RemovePlayerChip(pos Position) {
	// create a substring to help find the board
	subStr := fmt.Sprintf("_%d_%d", pos.X, pos.Y)

	for name, cell := range b.Board {
		// find the board to update
		if strings.Contains(name, subStr) {
			// if the cell does not have a chip set ignore this move and try again
			if !cell.ChipPlaced {
				log.Println("no chip to remove, choose another position")
			}

			// remove the placed chip
			cell.ChipPlaced = false
			// remove the teams chip color from the cell
			cell.ChipColor = ""

		}
	}

}

// newBoardCell creates a new pointer to a BoardCell
func newBoardCell(b BoardCell) *BoardCell {
	return &BoardCell{
		Type:       b.Type,
		Suit:       b.Suit,
		X:          b.X,
		Y:          b.Y,
		IsCorner:   b.IsCorner,
		ChipColor:  b.ChipColor,
		ChipPlaced: b.ChipPlaced,
	}
}
