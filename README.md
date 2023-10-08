# Sequence

The game of sequence but in a browser


## Parts of the game

- Cards
    - Deck
        - Shuffle cards
        - Deal cards to every player
    - Discard Pile
        - Add to pile after every turn
        - If deck reaches zero then the discard pile should be shuffled and used
        as the deck
- Board
    - 100 spaces to be used to create sequences
    - 4 corners are "Free Spaces"
    - Add games pieces to fill in unused spaces
    - Remove game pieces in certain circumstances
- Players
    - Play cards from their hand
    - Place pieces on the board corresponding the card played
    - Draw cards from the deck at the start of every turn
}


## Lobby implementation

### What i have right now
- a create lobby page that is connected to my lobby websocket. the idea with
it is to define the game seting for that lobby then send the configuration to the
server and have the server create a new lobby with the game settings
