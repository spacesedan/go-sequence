document.body.addEventListener("htmx:wsConfigSend", function(e) {
    const numOfPlayers = e.detail.parameters["num_of_players"]
    const maxHandSize = e.detail.parameters["max_hand_size"]

    e.detail.parameters = {
        action: "create_lobby",
        settings: {
            num_of_players: numOfPlayers,
            max_hand_size: maxHandSize
        }
    }
})

document.body.addEventListener("htmx:wsAfterMessage", function(e){
    console.log(e)
})
