package services

import "log/slog"

type GameService interface{}

type gameService struct {
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

func NewGameService(cs CardService, bs BoardService, ps PlayerService) GameService {
	cs.NewDeck()
	bs.NewBoard(BoardCellsJSONPath)

	return &gameService{
		BoardService:  bs,
		CardService:   cs,
		PlayerService: ps,
	}
}
