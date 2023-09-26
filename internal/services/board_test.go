package services

import (
	"fmt"
	"testing"
)

const TestPath = "testdata/board_cells.json"

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

	// Make sure the board has 100 cells
	if len(board) != 100 {
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
		cellName := fmt.Sprintf("Corner_%d_%d", cp.X, cp.Y)
		cell := board[cellName]

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
