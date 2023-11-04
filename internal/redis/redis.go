package redis

type LobbyState struct {
	CurrentState    CurrentState
	Players         map[string]*PlayerState
	ColorsAvailable map[string]bool
	Settings        Settings
}

