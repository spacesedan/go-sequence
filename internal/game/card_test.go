package game

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewCardService(t *testing.T) {
	cs := NewCardService()
	if cs == nil {
		t.Error("Expected card service to not be nil")
	}
}

// TestNewDeck check to see that a new deck is created with 104 cards
func TestNewDeck(t *testing.T) {
	cardService := NewCardService()
	cardService.NewDeck()
	deck := cardService.getDeck()

	if len(deck) != 104 {
		t.Errorf("deck should have 104 cards, instead it has %d", len(deck))
	}
}

// TestShuffleDeck check to make sure deck has been shuffled
func TestShuffleDeck(t *testing.T) {
	cardService := NewCardService()
	cardService.NewDeck()
	deck := cardService.getDeck()

	cardOne := deck[0]
	cardTwo := deck[1]

	cardService.shuffleDeck()

	deck = cardService.getDeck()

	// Keep in mind that there is a chance for this check to pass even though
	// the deck has been shuffled
	// there is a 1 in 10816 chance that this could happen
	if cardOne == deck[0] && cardTwo == deck[1] {
		t.Errorf("Expected deck to be shuffled")
	}
}

// TestDealOneCard dealing a card should reduce the size of teh deck by one
func TestDealOneCard(t *testing.T) {
	cs := NewCardService()
	cs.NewDeck()

	deckLengthBefore := len(cs.getDeck())

	cs.dealOneCard()

	deckLengthAfter := len(cs.getDeck())

	if deckLengthBefore == deckLengthAfter {
		t.Errorf("Expected deck to be smaller after dealing a card")
	}

	if deckLengthAfter != deckLengthBefore-1 {
		t.Errorf("Expected deck to be 1 less after dealing a card")
	}

}

// TestAddingToDiscardPile once a card gets played it is added to the discard pile
func TestAddingToDiscardPile(t *testing.T) {
	cs := NewCardService()

	cs.NewDeck()

	card := cs.dealOneCard()
	cs.AddToDiscardPile(card)

	discardPile := cs.getDiscardPile()

	if len(discardPile) != 1 {
		t.Errorf("Expected size of discard pile to be one")
	}
}

// TestResettingDeck as the size of the deck gets closer to zero the size of the
// discard pile increases, once the deck is empty the discard and the deck switch
func TestResettingDeck(t *testing.T) {
	cs := NewCardService()

	cs.NewDeck()

	deck := cs.getDeck()

	for i := 0; i < len(deck)+1; i++ {
		card := cs.dealOneCard()
		cs.AddToDiscardPile(card)
	}

	if len(cs.getDeck()) != 103 {
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

	cs := NewCardService()

	cs.NewDeck()

	deckLengthBeforeDealing := len(cs.getDeck())

	handSize := 7

	cs.DealCards(players, handSize)

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

	deckLengthAfterDealing := len(cs.getDeck())

	// Check to see if the size of the deck changed after dealing cards to players
	if deckLengthAfterDealing == deckLengthBeforeDealing {
		t.Errorf("Expected the deck size to reflect the number of cards drawn; but go a size of %v", deckLengthAfterDealing)
	}

}
