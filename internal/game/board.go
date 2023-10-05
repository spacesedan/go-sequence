package game

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

// Position
type CellPosition struct {
	X int
	Y int
}
