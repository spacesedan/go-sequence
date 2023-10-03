package services

import "log/slog"

type GameService struct {
	BoardService
	CardService
	PlayerService
	logger        *slog.Logger
	GameOver      bool
	CurrentPlayer int
}

// path to file which contains board game cell information
const (
	BoardSize          = 10
	NumOfPlayers       = 2
	HandSize           = 7
	SequenceSize       = 5
	BoardCellsJSONPath = "data/board_cells.json"
)

func NewGameService(logger *slog.Logger) GameService {
	// initialize our services
	cs := NewCardService()
	bs := NewBoardService()
	ps := NewPlayerService()

	// create a new deck
	cs.NewDeck()
	// create the board
	bs.NewBoard(BoardCellsJSONPath)

	return GameService{
		BoardService:  bs,
		CardService:   cs,
		PlayerService: ps,
	}
}
