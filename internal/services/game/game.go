package game

import (
	"log"
	"math/rand"

	"github.com/google/uuid"
)

type GameServiceInterface interface {
	AddPlayer(p *Player)
	RemovePlayer(id uuid.UUID)
	DealCards(n int)
	HandlePlayerTurn(tp TurnPayload)
	PlayCard(p Player, index int) error
	DrawOneCard(playerId uuid.UUID)

	// Debug funcs
	GetDeck() Deck
	GetPlayers() Players
}

// GameService game state
type GameService struct {
	Deck        Deck
	DiscardPile DiscardPile
	Players     Players
	// HandsDealt checks to see if cards have been dealt
	HandsDealt bool
	// MaxHandSize defines the max hand size for the given amount of players
	MaxHandSize int
	// MaxPlayers defines the max amount of players in a game
	MaxPlayers int
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
	ID    uuid.UUID
	Name  string
}

type Deck []Card
type DiscardPile []Card
type Players map[uuid.UUID]*Player

// NewGame creates a new game
func NewGame() GameServiceInterface {
	deck := newDeck()
	deck = shuffleDeck(deck)
	return &GameService{
		Deck:       deck,
		Players:    make(Players),
		HandsDealt: false,
        MaxPlayers: 2,
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
			// add two copies of every card to the deck
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

// AddPlayer adds a player to the game
func (g *GameService) AddPlayer(player *Player) {
	if len(g.Players) < g.MaxPlayers {
        player.ID = uuid.New()
		g.Players[player.ID] = player
    } else {
        log.Println("No room for other players")
    }

}

// RemovePlayer removes a player from the game
func (g *GameService) RemovePlayer(playerId uuid.UUID) {
	delete(g.Players, playerId)
}

func (g GameService) GetDeck() Deck {
	return g.Deck
}

func (g GameService) GetPlayers() Players {
	return g.Players
}

// ShuffleDeck shuffle deck
// TODO: add logic to shuffle DiscardPile when the deck has been spent
func shuffleDeck(d Deck) Deck {
	for i := 1; i < len(d); i++ {
		r := rand.Intn(i + 1)
		if i != r {
			d[r], d[i] = d[i], d[r]
		}
	}

	return d
}

// DealCards deals cards to every player based on the input
func (g *GameService) DealCards(n int) {
    //if hands have already been dealt do nothing
	if g.HandsDealt {
		return
	}

	for i := 0; i < n; i++ {
		for _, player := range g.Players {
			card := g.DealOneCard()
			player.Hand = append(player.Hand, card)
		}
	}

    // Set HandsDealt to true to prevent dealing more cards
	g.HandsDealt = true
}

// DealOneCard deals a single card to a player
func (g *GameService) DealOneCard() Card {
	if len(g.Deck) == 0 {
		return Card{}
	}

	// Deal a card from the top
	card := g.Deck[len(g.Deck)-1]

	// upate the game deck to reflect the removed card
	g.Deck = g.Deck[:len(g.Deck)-1]

	return card

}

type TurnPayload struct {
	Player    Player
	HandIndex int
}

func (g *GameService) HandlePlayerTurn(tp TurnPayload) {
	g.PlayCard(tp.Player, tp.HandIndex)
	g.DrawOneCard(tp.Player.ID)
}

// PlayCard player plays a card from their hand and adds it to the discard pile
func (g *GameService) PlayCard(p Player, index int) error {

	var cardPlayed Card
	var newHand []Card

	for i := 0; i < len(p.Hand); i++ {
		if i == index {
			// get the value of the card played
			cardPlayed = p.Hand[i]
		} else {
			// create a new slice with cards that are still in the players hand
			newHand = append(newHand, p.Hand[i])
		}
	}

	// add the played card to the discard pile
	g.DiscardPile = append(g.DiscardPile, cardPlayed)

	// update player hand
	p.Hand = newHand
	return nil
}

// DrawOneCard player draws one card from the top of the deck and adds it to
// thier hand
func (g *GameService) DrawOneCard(playerId uuid.UUID) {
	// draw a card from the top of the deck
	cardDrawn := g.Deck[len(g.Deck)-1]

	// update deck to reflect the drawn card
	g.Deck = g.Deck[:len(g.Deck)-1]

	// add the drawn card to hand
	g.Players[playerId].Hand = append(g.Players[playerId].Hand, cardDrawn)

}
