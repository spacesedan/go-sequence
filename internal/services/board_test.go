package services

import (
	"testing"
)

const TestPath = "testdata/board_cells.json"
const TestPathBadJSON = "testdata/bad.json"

func TestNewBoardService(t *testing.T) {
	bs := NewBoardService()

	if bs == nil {
		t.Error("Expected board service to not be nil")
	}
}

func TestNewBoard(t *testing.T) {
	bs := NewBoardService()
	bs.NewBoard(TestPath)

	board := bs.GetBoard()

	// number of cell
	n := 0

	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			n++
		}
	}

	// Make sure the board has 100 cells
	if n != 100 {
		t.Error("Expected board to contain 100 cells")
	}

	cornerPositions := []Position{
		{X: 0, Y: 0},
		{X: 9, Y: 0},
		{X: 0, Y: 9},
		{X: 9, Y: 9},
	}

	// Make sure the board corners are in the proper location
	for _, cp := range cornerPositions {
		cell := board[cp.X][cp.Y]

		if !cell.IsCorner {
			t.Errorf("Expected cell at postion X: %d Y: %d, to be a corner", cp.X, cp.Y)
		}

		if !cell.ChipPlaced {
			t.Error("Expected a chip to be placed in this location")
		}

		if cell.ChipColor != "Any" {
			t.Error("Expected cell chip color to be Any")
		}
	}
}

func TestNewBoardFailCases(t *testing.T) {
	testCases := []struct {
		name     string
		fileName string
	}{
		{name: "Bad path", fileName: "bad_path"},
		{name: "Bad JSON", fileName: TestPathBadJSON},
	}

	for _, tc := range testCases {
		bs := NewBoardService()
		err := bs.NewBoard(tc.fileName)

		if err == nil {
			t.Error("Expected but got none")
		}

	}
}

func TestNewBoardBadPath(t *testing.T) {
	bs := NewBoardService()
	err := bs.NewBoard("bad_path")

	if err == nil {
		t.Error("Expected an error from entering a bad path")
	}
}

// TestAddPlayerChip check to see if a chip and color are added to a cell when
// a player plats a card
func TestAddPlayerChip(t *testing.T) {
	bs := NewBoardService()
	bs.NewBoard(TestPath)

	board := bs.GetBoard()

	// player color
	color := "green"

	// player stub
	player := Player{
		Name:  "Player 1",
		Color: color,
	}

	// card stub
	card := Card{
		Suit: "Spade",
		Type: "Four",
	}

	// postion stub
	pos := Position{
		X: 6,
		Y: 0,
	}

	// Add a player chip using the stubbed data
	bs.AddPlayerChip(player, card, pos)

	// get cell from the board
	cell := board[pos.X][pos.Y]

	// check to see if the color of the cell is equal to the player color
	if cell.ChipColor != color {
		t.Error("Expected chip color to be player color")
	}

	// check to see if a chip has been placed on this cell
	if !cell.ChipPlaced {
		t.Error("Expected chip placed to be true")
	}
}

func TestAddPlayerChipIllegalMove(t *testing.T) {

	bs := NewBoardService()
	bs.NewBoard(TestPath)

	// player color
	color := "green"

	// player stub
	player := Player{
		Name:  "Player 1",
		Color: color,
	}

	// card stub
	card := Card{
		Suit: "Spade",
		Type: "Four",
	}

	// postion stub
	pos := Position{
		X: 6,
		Y: 0,
	}

	// Add a player chip using the stubbed data
	_, err := bs.AddPlayerChip(player, card, pos)
	if err != nil {
		t.Error("Expected cell to be updated with player information")
	}
	_, err = bs.AddPlayerChip(player, card, pos)
	if err == nil {
		t.Errorf("Expected an error, chip has already been taken")
	}
}

// TestRemovePlayerChip test removal of a player chip from a cell
func TestRemovePlayerChip(t *testing.T) {
	bs := NewBoardService()
	bs.NewBoard(TestPath)

	board := bs.GetBoard()

	color := "green"

	player := Player{
		Name:  "Player 1",
		Color: color,
	}

	card := Card{
		Suit: "Spade",
		Type: "Four",
	}

	pos := Position{
		X: 6,
		Y: 0,
	}

	bs.AddPlayerChip(player, card, pos)

	bs.RemovePlayerChip(pos)

	cell := board[pos.X][pos.Y]

	if cell.ChipColor != "" {
		t.Error("Expected chip to be empty")
	}

	if cell.ChipPlaced {
		t.Error("Expected not to have a chip placed on this cell")
	}

}

func TestRemovePlayerChipCellNotTaken(t *testing.T) {
	bs := NewBoardService()
	bs.NewBoard(TestPath)

	pos := Position{
		X: 6,
		Y: 0,
	}

	err := bs.RemovePlayerChip(pos)
	if err == nil {
		t.Error("Expected an error, there is no chip on this cell")
	}

}

func TestRemovePlayerChipCellLocked(t *testing.T) {
	bs := NewBoardService()
	bs.NewBoard(TestPath)

	color := "green"

	player := Player{
		Name:  "Player 1",
		Color: color,
	}

	card := Card{
		Suit: "Spade",
		Type: "Four",
	}

	pos := Position{
		X: 6,
		Y: 0,
	}

	cell, _ := bs.AddPlayerChip(player, card, pos)

	cell.CellLocked = true

	err := bs.RemovePlayerChip(pos)
	if err == nil {
		t.Error("Expected error while removing a locked cell, but got none")
	}
}
