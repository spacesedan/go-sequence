package game

import (
	"bufio"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"os"

	"github.com/google/uuid"
	"github.com/spacesedan/go-sequence/internal/services"
)

type GameService interface {
	// Deck & Discard Pile
	DealCards() error
	DrawCard(*Player) Card
	AddToDiscardPile(Card)
	GetDeck() Deck
	GetDiscardPile() DiscardPile

	// Board
	GetBoard() Board
	AddPlayerChip(*Player, Card, CellPosition) (*BoardCell, error)
	RemovePlayerChip(CellPosition) error

	// Player
	AddPlayer(*Player) error
	RemovePlayer(uuid.UUID) error
	GetPlayer(uuid.UUID) (*Player, error)
	GetPlayers() Players

	PlayerPlayCardFromHand(*Player, int) (Card, error)
	PlayerAddCardTooHand(*Player, Card)
}

type gameService struct {
	Deck          Deck
	DiscardPile   DiscardPile
	Board         Board
	Players       Players
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

func NewGameService() GameService {
	board, err := NewBoard(BoardCellsJSONPath)
	if err != nil {
		panic(err)
	}

	return &gameService{
		Deck:        shuffleDeck(NewDeck()),
		DiscardPile: DiscardPile{},
		Board:       board,
		Players:     make(Players),
	}
}

// DECK & DISCARD PILE LOGIC -------------------------------------------

// NewDeck creates a new deck
func NewDeck() Deck {
	var deck Deck

	// Create a new deck
	types := []string{"Two", "Three", "Four", "Five", "Six", "Seven",
		"Eight", "Nine", "Ten", "Jack", "Queen", "King", "Ace",
	}
	suits := []string{"Spade", "Heart", "Club", "Diamond"}

	for i := 0; i < len(types); i++ {
		for n := 0; n < len(suits); n++ {
			card := Card{
				Type: types[i],
				Suit: suits[n],
			}

			// add two copies of every card to the deck
			deck = append(deck, card)
			deck = append(deck, card)

		}
	}

	return deck

}

// shuffleDeck shuffles a deck and returns it
func shuffleDeck(d Deck) Deck {
	for i := 1; i < len(d); i++ {
		r := rand.Intn(i + 1)
		if i != r {
			d[r], d[i] = d[i], d[r]
		}
	}

	return d
}

// dealOneCard get a single card from the deck and update the deck
func (g *gameService) dealOneCard() Card {

	// if the deck size reaches zero
	if len(g.Deck) == 0 {
		// switch the deck with the discard pile
		g.Deck, g.DiscardPile = Deck(g.DiscardPile), DiscardPile(g.Deck)
		// reshuffle the deck
		g.Deck = shuffleDeck(g.Deck)
	}

	// Deal a card from the top
	card := g.Deck[len(g.Deck)-1]

	// upate the game deck to reflect the removed card
	g.Deck = g.Deck[:len(g.Deck)-1]

	return card
}

// DealCards deals cards to the players based on a given hand size
func (g *gameService) DealCards() error {
	// return an error if there are no players to deal cards too.
	if len(g.Players) == 0 {
		return services.WrapErrorf(
			errors.New("Illegal move; no players to deal cards too"),
			services.ErrorCodeIllegalMove,
			"gameService.DealCards")
	}

	// Deal a single card to every player until the desired hand size is reached
	for i := 0; i < HandSize; i++ {
		for _, player := range g.Players {
			card := g.dealOneCard()
			player.Hand = append(player.Hand, card)
		}
	}

	return nil
}

// DrawCard Draw a card from the deck and add it to the players hand
func (g *gameService) DrawCard(player *Player) Card {
	if len(player.Hand) < HandSize {
		card := g.dealOneCard()
		return card
	}

	return Card{}
}

// AddToDiscardPile adds card to the discard pile
func (g *gameService) AddToDiscardPile(card Card) {
	g.DiscardPile = append(g.DiscardPile, card)
}

// getDeck returns the current Deck
func (g gameService) GetDeck() Deck {
	return g.Deck
}

// getDiscardPile returns the discard pile
func (g gameService) GetDiscardPile() DiscardPile {
	return g.DiscardPile
}

// BOARD LOGIC -------------------------------------------

// NewBoard creates a new game board
func NewBoard(fileName string) (Board, error) {
	var board Board

	cells, err := boardCellsFromFile(fileName)
	if err != nil {
		return Board{}, services.WrapErrorf(err, services.ErrorCodeNotFound, "services.NewBoard")
	}

	for _, cell := range cells {

		// If the cell is a corner
		if cell.X == 0 && cell.Y == 0 || cell.X == 9 && cell.Y == 0 || cell.X == 0 && cell.Y == 9 || cell.X == 9 && cell.Y == 9 {

			// set the values for a corner
			cell.IsCorner = true
			cell.ChipPlaced = true
			cell.ChipColor = "Any"

			cellPointer := newBoardCell(cell)

			board[cell.X][cell.Y] = cellPointer

		} else {
			cellPointer := newBoardCell(cell)

			board[cell.X][cell.Y] = cellPointer
		}

	}

	return board, nil

}

// GetBoard returns the game board
func (g gameService) GetBoard() Board {
	return g.Board
}

// AddPlayerChip adds a chip to a cell on the board using a card and a cell position
func (g *gameService) AddPlayerChip(player *Player, card Card, pos CellPosition) (*BoardCell, error) {

	cell := g.Board[pos.X][pos.Y]

	// check to see if the cell is already occupied
	if cell.ChipPlaced {
		// add more information later
		return nil, services.WrapErrorf(errors.New("Illegal Move, cell is taken"),
			services.ErrorCodeIllegalMove,
			"boardService.AddPlayerChip")
	}

	player.Cells[pos.X][pos.Y] = cell

	cell.ChipColor = player.Color
	cell.ChipPlaced = true
	cell.Player = player

	return cell, nil
}

// RemovePlayerChip removes chip and color set on a cell
func (g gameService) RemovePlayerChip(pos CellPosition) error {
	cell := g.Board[pos.X][pos.Y]

	if !cell.ChipPlaced {
		return services.WrapErrorf(
			errors.New("Illegal Move: cell not taken"),
			services.ErrorCodeIllegalMove,
			"boardService.RemovePlayerChip")

	}

	if cell.CellLocked {
		return services.WrapErrorf(
			errors.New("Illegal Move: cell is a part of a sequence"),
			services.ErrorCodeIllegalMove,
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
		return BoardCells{}, services.WrapErrorf(err, services.ErrorCodeNotFound, "os.Open")
	}

	// close the file once the function is executed
	defer file.Close()

	// create a reading from teh file
	r := bufio.NewReader(file)

	// decode the loaded file into a usable struct
	err = json.NewDecoder(r).Decode(&cells)
	if err != nil {
		return BoardCells{}, services.WrapErrorf(err, services.ErrorCodeUnknown, "json.NewDecoder")
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

// PLAYER LOGIC -------------------------------------------

// AddPlayer add a player to the player list
func (g *gameService) AddPlayer(player *Player) error {
	if _, ok := g.Players[player.ID]; ok {
		return services.WrapErrorf(errors.New("Error: Player already exists"),
			services.ErrorCodeInvalidArgument,
			"playerService.AddPlayer",
		)
	}
	player.Cells = PlayerCells{}
	g.Players[player.ID] = player

	return nil
}

// RemovePlayer removes a player from the player list
func (g *gameService) RemovePlayer(playerId uuid.UUID) error {
	// check to see if the player exists
	if _, ok := g.Players[playerId]; !ok {
		// If not found return an error
		return services.WrapErrorf(errors.New("No player found"), services.ErrorCodeNotFound, "playerService.RemovePlayer")
	}

	// if the player does exist remove them from the player list
	delete(g.Players, playerId)

	return nil
}

func (g *gameService) GetPlayers() Players {
	return g.Players
}

// Get a player form the player list using their id
func (g gameService) GetPlayer(playerId uuid.UUID) (*Player, error) {
	// check to see if the player exists
	player, ok := g.Players[playerId]
	if !ok {
		// If not found return an error
		return nil, services.WrapErrorf(
			errors.New("No player found"),
			services.ErrorCodeNotFound,
			"playerService.RemovePlayer")
	}

	return player, nil
}

// PlayerPlayCardFromHand player plays a card from their hand using a card index,
// it returns the card a the players played or an error
func (g *gameService) PlayerPlayCardFromHand(player *Player, cardIndex int) (Card, error) {
	// check to see if the card played is in the players hand
	if cardIndex > len(player.Hand) {
		return Card{}, services.WrapErrorf(
			errors.New("Illegal move; cannot play card that is not in your hand"),
			services.ErrorCodeIllegalMove,
			"playerService.PlayerPlayCardFromHand")

	}

	// newHand holds the value of the player hand minus the card that was played
	var newHand []Card

	// cardPlayed is the card a player want to play
	cardPlayed := player.Hand[cardIndex]

	// Update the player hand
	for i := 0; i < len(player.Hand); i++ {
		// we don't want to add this card back to the player hand so we ignore it
		// in the loop
		if cardPlayed == player.Hand[i] {
			continue
		}

		newHand = append(newHand, player.Hand...)
		// update the player hand with the new hand
		player.Hand = newHand
	}

	return cardPlayed, nil
}

// PlayerAddCardTooHand adds a card to the players hand
func (g *gameService) PlayerAddCardTooHand(player *Player, card Card) {
	player.Hand = append(player.Hand, card)
}
