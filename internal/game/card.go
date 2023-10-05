package game

// Card holds the Suit and the value of a card
type Card struct {
	Suit string
	Type string
}

// Slice of cards where plays get dealt cards and draw from
type Deck []Card

// Slice of cards where plays put played cards, get reshuffled and made into
// the new playing card
type DiscardPile []Card
