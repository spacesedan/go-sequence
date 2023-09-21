get_deck:
	curl -s http://localhost:42069/game


get_players:
	curl -s http://localhost:42069/game/p


add_player_1:
	curl -s http://localhost:42069/game/p/add -X "POST" -d '{"name" : "player 1", "color": "blue"}'


add_player_2:
	curl -s http://localhost:42069/game/p/add -X "POST" -d '{"name" : "player 2", "color": "red"}'


deal_cards:
	curl -s http://localhost:42069/game/deal
