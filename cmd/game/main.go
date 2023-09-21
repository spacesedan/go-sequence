package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Card struct {
	Suit string
	Type string
}

type Player struct {
	Hand  []Card
	Color string
}

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
	game.Deal(6)
	game.PlayCard(game.Players[0], 4)
	fmt.Println("discard pile length", len(game.DiscardPile))
	fmt.Println("deck size", len(game.Deck))
	game.PlayCard(game.Players[1], 2)
	fmt.Println("discard pile length", len(game.DiscardPile))
	fmt.Println("deck size", len(game.Deck))
	game.PlayCard(game.Players[2], 3)

	fmt.Println("discard pile length", len(game.DiscardPile))
	fmt.Println("deck size", len(game.Deck))

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
		g.Players[i%len(g.Players)].Hand = append(g.Players[i%len(g.Players)].Hand, g.Deck[i])
		fmt.Println(g.Deck[i], i)
	}

	// remove the cards dealt from the deck
	g.Deck = g.Deck[n*len(g.Players):]
}

// PlayCard handles what happens when a card is played
// it takes the payer and an index and updates
func (g *Game) PlayCard(player Player, index int) Card {
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

	// updated deck slice
	updatedDeck := g.Deck[:len(g.Deck)-1]

	// add drawnCard to the Players Hand
	newHand = append(newHand, drawnCard)

	// update deck and discard pile
	g.Deck = updatedDeck
	g.DiscardPile = append(g.DiscardPile, cardPlayed)

	// update player hand
	player.Hand = newHand


	return cardPlayed
}
