package game

import "math/rand"

type GameServiceInterface interface {
	WaitingRoom() error
	ShuffleDeck()
	DealCards(n int)
	PlayCard(tp TurnPayload) error
	// Debug funcs
	GetDeck() Deck
}

// GameService holds the Deck DiscardPile, # of Players, and the MaxHandSize
type GameService struct {
	Deck        Deck
	DiscardPile []Card
	Players     []Player
	MaxHandSize int
}

// Card holds the Suit and the value of a card
type Card struct {
	Suit string
	Type string
}

// Player holds the Current Hand and the color of the player
type Player struct {
	Hand  []Card
	Color string
}

// Game holds the Deck DiscardPile, # of Players, and the MaxHandSize for the given players
type Game struct {
	Deck        Deck
	DiscardPile []Card
	Players     []Player
	MaxHandSize int
}

type Deck []Card

// NewGame creates a new game
func NewGame() GameServiceInterface {
	deck := newDeck()
	return &GameService{
		Deck: deck,
	}
}

// newDeck create a new deck to be used during the game
func newDeck() Deck {
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
			deck = append(deck, card)
			deck = append(deck, card)
		}
	}

	return deck
}

// WaitingRoom wait for atleast two players before starting the game
func (g GameService) WaitingRoom() error {
	return nil
}

func (g GameService) GetDeck() Deck {
	return g.Deck
}

// ShuffleDeck shuffle deck
// TODO: add logic to shuffle DiscardPile when the deck has been spent
func (g GameService) ShuffleDeck() {
	d := g.Deck
	for i := 1; i < len(d); i++ {
		r := rand.Intn(i + 1)
		if i != r {
			d[r], d[i] = d[i], d[r]
		}
	}
}

// DealCards deals cards to every player
func (g *GameService) DealCards(n int) {
	for i := 0; i < n*len(g.Players); i++ {
		g.Players[i%len(g.Players)].Hand =
			append(g.Players[i%len(g.Players)].Hand, g.Deck[i])
	}

	// remove the cards dealt from the deck
	g.Deck = g.Deck[n*len(g.Players):]
}

type TurnPayload struct {
	Player    Player
	HandIndex int
}

// Play Card plays a card from the players hand, draws a new card from the deck
// updates the players hand
// add the played card to the discard pile
// updates the deck to reflect the player drawing a card
// TODO: function handles too much consider breaking it down
func (g *GameService) PlayCard(turnPayload TurnPayload) error {
	p := turnPayload.Player

	var cardPlayed Card
	var newHand []Card

	for i := 0; i < len(p.Hand); i++ {
		if i == turnPayload.HandIndex {
			// get the value of the card played
			cardPlayed = p.Hand[i]
		} else {
			// create a new slice with cards that are still in the players hand
			newHand = append(newHand, p.Hand[i])
		}
	}

	// draw a new card from the top
	drawnCard := g.Deck[len(g.Deck)-1]

	// create a copy of the deck to reflect the recently drawn card
	updatedDeck := g.Deck[:len(g.Deck)-1]

	// add drawnCard to the Players Hand
	newHand = append(newHand, drawnCard)

	// update deck and discard pile
	g.Deck = updatedDeck
	g.DiscardPile = append(g.DiscardPile, cardPlayed)

	// update player hand
	p.Hand = newHand
	return nil
}
