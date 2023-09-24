package services

import (
	"math/rand"
)

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

// CardService
type CardService interface {
	NewDeck()
	DealCards(p Players, n int)
	AddToDiscardPile(card Card)
	DrawCard(p Player, n int)

	// utility functions
	shuffleDeck()
	dealOneCard() Card
	getDeck() Deck
	getDiscardPile() DiscardPile
}

type cardService struct {
	Deck        Deck
	DiscardPile DiscardPile
}

func NewCardService() CardService {
	return &cardService{}
}

// NewDeck creates a new deck
func (c *cardService) NewDeck() {
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

	c.Deck = deck
}

// ShuffleDeck shuffles the deck
func (c cardService) shuffleDeck() {
	d := c.Deck
	for i := 1; i < len(d); i++ {
		r := rand.Intn(i + 1)
		if i != r {
			d[r], d[i] = d[i], d[r]
		}
	}
}

// dealOneCard get a single card from the deck and update the deck
func (c *cardService) dealOneCard() Card {

	// if the deck size reaches zero
	if len(c.Deck) == 0 {
		// switch the deck with the discard pile
		c.Deck, c.DiscardPile = Deck(c.DiscardPile), DiscardPile(c.Deck)
		// reshuffle the deck
		c.shuffleDeck()
	}

	// Deal a card from the top
	card := c.Deck[len(c.Deck)-1]

	// upate the game deck to reflect the removed card
	c.Deck = c.Deck[:len(c.Deck)-1]

	return card

}

// DealCards deals cards to the players based on a given hand size
func (c *cardService) DealCards(players Players, handSize int) {
	for i := 0; i < handSize; i++ {
		for _, player := range players {
			card := c.dealOneCard()
			player.Hand = append(player.Hand, card)
		}
	}
}

// DrawCard Draw a card from the deck and add it to the players hand
func (c *cardService) DrawCard(player Player, maxHandSize int) {
	if len(player.Hand) < maxHandSize {
		card := c.dealOneCard()
		player.Hand = append(player.Hand, card)
	}
}

// AddToDiscardPile adds card to the discard pile
func (c *cardService) AddToDiscardPile(card Card) {
	c.DiscardPile = append(c.DiscardPile, card)
}

// getDeck returns the current Deck
func (c cardService) getDeck() Deck {
	return c.Deck
}

// getDiscardPile returns the discard pile
func (c cardService) getDiscardPile() DiscardPile {
	return c.DiscardPile
}
