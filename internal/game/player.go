package game

import "github.com/google/uuid"

// Player contains information for a single player
type Player struct {
	Hand  []Card
	Cells PlayerCells
	Color string
	ID    uuid.UUID
	Name  string
}

type PlayerCells [BoardSize][BoardSize]*BoardCell

type Players map[uuid.UUID]*Player
