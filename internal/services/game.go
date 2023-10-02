package services

type GameService interface{}

type gameService struct {
    BoardService
    CardService
    PlayerService
}


func NewGameService(cs CardService, bs BoardService, ps PlayerService) GameService {
    cs.NewDeck()
    bs.NewBoard(BoardCellsJSONPath)

    return &gameService{
        BoardService: bs,
        CardService: cs,
        PlayerService: ps,
    }
}
