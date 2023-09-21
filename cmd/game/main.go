package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

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
	Deck        []Card
	DiscardPile []Card
	Players     []Player
	MaxHandSize int
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func main() {
	game := NewGame([]Player{
		{Color: "Blue"},
		{Color: "Green"},
		{Color: "Red"},
	})

	game.DebugDeck()
	game.ShuffleDeck()
	game.DebugDeck()
	game.Deal(10)
	game.PlayCard(game.Players[0], 4)
	fmt.Printf("%#v\n", game.Deck)

	fmt.Println("--------------")
	fmt.Printf("%#v\n", game.DiscardPile)
	fmt.Println("--------------")
	fmt.Printf("%#v\n", game.Players)
}

func NewGame(players []Player) Game {
	var game Game

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
			game.Deck = append(game.Deck, card)
			game.Deck = append(game.Deck, card)
		}
	}

	for _, player := range players {
		game.Players = append(game.Players, player)
	}

	return game
}

func (g Game) ShuffleDeck() {
	d := g.Deck
	for i := 1; i < len(d); i++ {
		r := rand.Intn(i + 1)
		if i != r {
			d[r], d[i] = d[i], d[r]
		}
	}
}

func (g Game) DebugDeck() {
	d := g.Deck
	if os.Getenv("DEBUG") != "" {
		for i := 0; i < len(d); i++ {
			fmt.Printf("Card #%d is a %s of %ss\n", i+1, d[i].Type, d[i].Suit)
		}

	}
}

// Deal a card to every player
func (g *Game) Deal(n int) {
	for i := 0; i < n*len(g.Players); i++ {
		g.Players[i%len(g.Players)].Hand =
			append(g.Players[i%len(g.Players)].Hand, g.Deck[i])
	}

	// remove the cards dealt from the deck
	g.Deck = g.Deck[n*len(g.Players):]
}

// PlayCard handles what happens when a card is played
// it takes the payer and an index and updates
func (g *Game) PlayCard(player Player, index int) {
	var cardPlayed Card
	var newHand []Card

	for i := 0; i < len(player.Hand); i++ {
		if i == index {
			cardPlayed = player.Hand[i]
		} else {
			newHand = append(newHand, player.Hand[i])
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
	player.Hand = newHand

}
