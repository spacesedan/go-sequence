package internal

// Current state ... are the players in the lobby still choosing thier colors,
// or are they in the game
type CurrentState uint

const (
	Unknown CurrentState = iota
	InLobby
	InGame
)

// String get a stringified version of the current game state
func (c CurrentState) String() string {
	switch c {
	case InLobby:
		return "lobby"
	case InGame:
		return "game"
	default:
		return "unknown"

	}
}

type Lobby struct {
	ID              string
	CurrentState    CurrentState
	Players         map[string]struct{}
	ColorsAvailable map[string]bool
	Settings        Settings
}
