package game

import (
	"testing"

	"github.com/google/uuid"
)

// TestNewGameService test the creation of the game service
func TestNewGameService(t *testing.T) {
	gs := NewGameService(TestPath)
	if gs == nil {
		t.Error("Expected gameService to not be nil")
	}
}

// CARD & DISCARD PILE TESTS -------------------------------------------------

// TestNewDeck check to see that a new deck is created with 104 cards
func TestNewDeck(t *testing.T) {
	deck := NewDeck()

	if len(deck) != 104 {
		t.Errorf("deck should have 104 cards, instead it has %d", len(deck))
	}
}

// TestShuffleDeck check to make sure deck has been shuffled
func TestShuffleDeck(t *testing.T) {
	deck := NewDeck()

	cardOne := deck[0]
	cardTwo := deck[1]

	deck = shuffleDeck(deck)

	// Keep in mind that there is a chance for this check to pass even though
	// the deck has been shuffled
	// there is a 1 in 10816 chance that this could happen
	if cardOne == deck[0] && cardTwo == deck[1] {
		t.Errorf("Expected deck to be shuffled")
	}
}

// TestDealOneCard dealing a card should reduce the size of teh deck by one
func TestDealOneCard(t *testing.T) {
	gs := NewGameService(TestPath)

	deckLengthBefore := len(gs.GetDeck())

	gs.DealOneCard()

	deckLengthAfter := len(gs.GetDeck())

	if deckLengthBefore == deckLengthAfter {
		t.Errorf("Expected deck to be smaller after dealing a card")
	}

	if deckLengthAfter != deckLengthBefore-1 {
		t.Errorf("Expected deck to be 1 less after dealing a card")
	}

}

// TestAddingToDiscardPile once a card gets played it is added to the discard pile
func TestAddingToDiscardPile(t *testing.T) {
	gs := NewGameService(TestPath)

	card := gs.DealOneCard()
	gs.AddToDiscardPile(card)

	discardPile := gs.GetDiscardPile()

	if len(discardPile) != 1 {
		t.Errorf("Expected size of discard pile to be one")
	}
}

// TestResettingDeck as the size of the deck gets closer to zero the size of the
// discard pile increases, once the deck is empty the discard and the deck switch
func TestResettingDeck(t *testing.T) {
	gs := NewGameService(TestPath)

	deck := gs.GetDeck()

	for i := 0; i < len(deck)+1; i++ {
		card := gs.DealOneCard()
		gs.AddToDiscardPile(card)
	}

	if len(gs.GetDeck()) != 103 {
		t.Error("Expected deck size to be 103 after dealing all the cards in the deck")
	}

}

func TestDealCards(t *testing.T) {
	playerOneId := uuid.New()
	playerTwoId := uuid.New()
	players := Players{
		playerOneId: {
			Name:  "Player 1",
			Color: "Blue",
			ID:    playerOneId,
		},
		playerTwoId: {
			Name:  "Player 2",
			Color: "Red",
			ID:    playerTwoId,
		},
	}

	gs := NewGameService(TestPath)

	// Add players to the game
	for _, p := range players {
		gs.AddPlayer(p)
	}

	deckLengthBeforeDealing := len(gs.GetDeck())

	handSize := 7

	gs.DealCards()

	playerOne := players[playerOneId]
	playerTwo := players[playerTwoId]

	// Check to see if the player hands equal to the set deal amout
	if len(playerOne.Hand) != handSize {
		t.Errorf("Expected player hand size to be equal to the set hand size; but got %v", len(playerOne.Hand))
	}

	// Check to see if players get the same size hands
	if len(playerOne.Hand) != len(playerTwo.Hand) {
		t.Error("Expected player hands to be the same size")
	}

	deckLengthAfterDealing := len(gs.GetDeck())

	// Check to see if the size of the deck changed after dealing cards to players
	if deckLengthAfterDealing == deckLengthBeforeDealing {
		t.Errorf("Expected the deck size to reflect the number of cards drawn; but go a size of %v", deckLengthAfterDealing)
	}

}

// BOARD TESTS ---------------------------------------------------------------

const TestPath = "testdata/board_cells.json"
const TestPathBadJSON = "testdata/bad.json"

func TestNewBoard(t *testing.T) {
	board, err := NewBoard(TestPath)

	if err != nil {
		t.Error("Expected board to be created without any issues")
	}

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

	cornerPositions := []CellPosition{
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
		_, err := NewBoard(tc.fileName)

		if err == nil {
			t.Error("Expected but got none")
		}

	}
}

// TestAddPlayerChip check to see if a chip and color are added to a cell when
// a player plats a card
func TestAddPlayerChip(t *testing.T) {
	gs := NewGameService(TestPath)

	board := gs.GetBoard()

	// player color
	color := "green"

	// player stub
	player := &Player{
		Name:  "Player 1",
		Color: color,
	}

	// card stub
	card := Card{
		Suit: "Spade",
		Type: "Four",
	}

	// postion stub
	pos := CellPosition{
		X: 6,
		Y: 0,
	}

	// Add a player chip using the stubbed data
	gs.AddPlayerChip(player, card, pos)

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

	if cell.Player != player {
		t.Error("Expected players to be the same")
	}
}

func TestAddPlayerChipIllegalMove(t *testing.T) {
	gs := NewGameService(TestPath)

	// player color
	color := "green"

	// player stub
	player := &Player{
		Name:  "Player 1",
		Color: color,
	}

    gs.AddPlayer(player)

	// card stub
	card := Card{
		Suit: "Spade",
		Type: "Four",
	}

	// postion stub
	pos := CellPosition{
		X: 6,
		Y: 0,
	}

	// Add a player chip using the stubbed data
	_, err := gs.AddPlayerChip(player, card, pos)
	if err != nil {
		t.Error("Expected cell to be updated with player information")
	}
	_, err = gs.AddPlayerChip(player, card, pos)
	if err == nil {
		t.Errorf("Expected an error, chip has already been taken")
	}
}

// TestRemovePlayerChip test removal of a player chip from a cell
func TestRemovePlayerChip(t *testing.T) {
    gs := NewGameService(TestPath)

	board := gs.GetBoard()

	color := "green"

	player := &Player{
		Name:  "Player 1",
		Color: color,
	}

    gs.AddPlayer(player)

	card := Card{
		Suit: "Spade",
		Type: "Four",
	}

	pos := CellPosition{
		X: 6,
		Y: 0,
	}

	gs.AddPlayerChip(player, card, pos)

	gs.RemovePlayerChip(pos)

	cell := board[pos.X][pos.Y]

	if cell.ChipColor != "" {
		t.Error("Expected chip to be empty")
	}

	if cell.ChipPlaced {
		t.Error("Expected not to have a chip placed on this cell")
	}

}

func TestRemovePlayerChipCellNotTaken(t *testing.T) {
    gs := NewGameService(TestPath)

	pos := CellPosition{
		X: 6,
		Y: 0,
	}

	err := gs.RemovePlayerChip(pos)
	if err == nil {
		t.Error("Expected an error, there is no chip on this cell")
	}

}

func TestRemovePlayerChipCellLocked(t *testing.T) {
    gs := NewGameService(TestPath)

	color := "green"

	player := &Player{
		Name:  "Player 1",
		Color: color,
	}

    gs.AddPlayer(player)

	card := Card{
		Suit: "Spade",
		Type: "Four",
	}

	pos := CellPosition{
		X: 6,
		Y: 0,
	}

	cell, _ := gs.AddPlayerChip(player, card, pos)

	cell.CellLocked = true

	err := gs.RemovePlayerChip(pos)
	if err == nil {
		t.Error("Expected error while removing a locked cell, but got none")
	}
}

// PLAYER TESTS --------------------------------------------------------------

func TestAddPlayer(t *testing.T) {
    gs := NewGameService(TestPath)

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}

	err := gs.AddPlayer(player)
	if err != nil {
		t.Error("Expected player to be added without any issues")
	}

	players := len(gs.GetPlayers())

	if players != 1 {
		t.Error("Expected players length to be 1 after adding a single player")
	}

}

func TestAddPlayerTwice(t *testing.T) {
    gs := NewGameService(TestPath)

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}

	gs.AddPlayer(player)
	err := gs.AddPlayer(player)
	if err == nil {
		t.Error("Expected an error after adding the same player twice")
	}
}

func TestRemovePlayer(t *testing.T) {
    gs := NewGameService(TestPath)

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}

	gs.AddPlayer(player)

	err := gs.RemovePlayer(player.ID)
	if err != nil {
		t.Error("Expected player to be removed without any errors")
	}

	players := len(gs.GetPlayers())

	if players != 0 {
		t.Error("Expected the number of players to be 0")
	}

}

func TestRemovePlayerTwice(t *testing.T) {
    gs := NewGameService(TestPath)

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}

	gs.AddPlayer(player)

	gs.RemovePlayer(player.ID)

	err := gs.RemovePlayer(player.ID)
	if err == nil {
		t.Error("Expected error after removing the same player twice")
	}

}

func TestGetPlayer(t *testing.T) {
    gs := NewGameService(TestPath)

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}

	gs.AddPlayer(player)

	p, err := gs.GetPlayer(player.ID)
	if err != nil {
		t.Error("Expected no error when tring to get a player that exists")
	}

	if p != player {
		t.Error("Expected the player recived to be that same player that we created")
	}

}

func TestGetPlayerThatDoesNotExist(t *testing.T) {
    gs := NewGameService(TestPath)

	badId := uuid.New()

	_, err := gs.GetPlayer(badId)
	if err == nil {
		t.Error("Expected an error when trying to get a player that does not exist")
	}
}

func TestPlayerAddCardToHand(t *testing.T) {
    gs := NewGameService(TestPath)

	card := Card{
		Suit: "Spade",
		Type: "Ace",
	}

	player := &Player{
		ID:    uuid.New(),
		Name:  "Player 1",
		Color: "green",
	}

	gs.AddPlayer(player)

	gs.PlayerAddCardToHand(player, card)

	player, _ = gs.GetPlayer(player.ID)

	if player.Hand[0] != card {
		t.Error("Expected this card to equal the one we just added")
	}

}
