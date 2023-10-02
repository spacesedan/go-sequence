package services

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
)

type BoardService interface {
	NewBoard(fn string) error
	GetBoard() Board
	AddPlayerChip(p *Player, c Card, pos Position) (*BoardCell, error)
	RemovePlayerChip(pos Position) error
}

type boardService struct {
	Board
}

// BoardLocation are the spots inside of the board
type BoardCell struct {
	Type       string  `json:"type"`
	Suit       string  `json:"suit"`
	X          int     `json:"x"`
	Y          int     `json:"y"`
	CellLocked bool    `json:"cell_locked"`
	IsCorner   bool    `json:"is_corner"`
	ChipPlaced bool    `json:"chip_placed"`
	ChipColor  string  `json:"chip_color"`
	Player     *Player `json:",omitempty"`
}

type BoardCells []BoardCell

type Board [BoardSize][BoardSize]*BoardCell

func NewBoardService() BoardService {
	return &boardService{
		Board: Board{},
	}
}

// NewBoard creates a new game board
func (b *boardService) NewBoard(fileName string) error {
	cells, err := boardCellsFromFile(fileName)
	if err != nil {
		return WrapErrorf(err, ErrorCodeNotFound, "services.NewBoard")
	}

	for _, cell := range cells {

		// If the cell is a corner
		if cell.X == 0 && cell.Y == 0 || cell.X == 9 && cell.Y == 0 || cell.X == 0 && cell.Y == 9 || cell.X == 9 && cell.Y == 9 {

			// set the values for a corner
			cell.IsCorner = true
			cell.ChipPlaced = true
			cell.ChipColor = "Any"

			cellPointer := newBoardCell(cell)

			b.Board[cell.X][cell.Y] = cellPointer

		} else {
			cellPointer := newBoardCell(cell)

			b.Board[cell.X][cell.Y] = cellPointer
		}

	}

	return nil

}

// GetBoard returns the game board
func (b boardService) GetBoard() Board {
	return b.Board
}

// Position
type Position struct {
	X int
	Y int
}

// AddPlayerChip adds a chip to a cell on the board using a card and a cell position
func (b boardService) AddPlayerChip(player *Player, card Card, pos Position) (*BoardCell, error) {

	cell := b.Board[pos.X][pos.Y]

	// check to see if the cell is already occupied
	if cell.ChipPlaced {
		// add more information later
		return nil, WrapErrorf(errors.New("Illegal Move, cell is taken"),
			ErrorCodeIllegalMove,
			"boardService.AddPlayerChip")
	}

	player.Cells[pos.X][pos.Y] = cell

	cell.ChipColor = player.Color
	cell.ChipPlaced = true
	cell.Player = player

	return cell, nil
}

// RemovePlayerChip removes chip and color set on a cell
func (b boardService) RemovePlayerChip(pos Position) error {
	cell := b.Board[pos.X][pos.Y]

	if !cell.ChipPlaced {
		return WrapErrorf(
			errors.New("Illegal Move: cell not taken"),
			ErrorCodeIllegalMove,
			"boardService.RemovePlayerChip")

	}

	if cell.CellLocked {
		return WrapErrorf(
			errors.New("Illegal Move: cell is a part of a sequence"),
			ErrorCodeIllegalMove,
			"boardService.RemovePlayerChip",
		)
	}

	// Remove the cell from the player
	cell.Player.Cells[pos.X][pos.Y] = nil

	// remove the placed chip
	cell.ChipPlaced = false
	// remove the teams chip color from the cell
	cell.ChipColor = ""
	cell.Player = nil

	return nil

}

// boardCellsFromFile returns board cells from a file
func boardCellsFromFile(fileName string) (BoardCells, error) {
	// cells is going to hold the cells array loaded from file
	var cells BoardCells

	// openn the cells file
	file, err := os.Open(fileName)
	if err != nil {
		return BoardCells{}, WrapErrorf(err, ErrorCodeNotFound, "os.Open")
	}

	// close the file once the function is executed
	defer file.Close()

	// create a reading from teh file
	r := bufio.NewReader(file)

	// decode the loaded file into a usable struct
	err = json.NewDecoder(r).Decode(&cells)
	if err != nil {
		return BoardCells{}, WrapErrorf(err, ErrorCodeUnknown, "json.NewDecoder")
	}

	return cells, nil
}

// newBoardCell creates a new pointer to a BoardCell
func newBoardCell(b BoardCell) *BoardCell {
	return &BoardCell{
		Type:       b.Type,
		Suit:       b.Suit,
		X:          b.X,
		Y:          b.Y,
		IsCorner:   b.IsCorner,
		CellLocked: b.CellLocked,
		ChipColor:  b.ChipColor,
		ChipPlaced: b.ChipPlaced,
	}
}
