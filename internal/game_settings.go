package internal


type Settings struct {
	NumOfPlayers int `json:"num_of_players"`
	MaxHandSize  int `json:"max_hand_size"`
	Teams        bool
}
