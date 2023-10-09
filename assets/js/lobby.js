const lobbyID = document.querySelector("#lobby_id")


document.body.addEventListener("htmx:wsOpen", function(e) {

    console.log(lobbyID.dataset["lobby_id"])
    const message = {
        action: "join_lobby",
        lobby_id: lobbyID.dataset["lobby_id"]
    }
    e.detail.socketWrapper.send(JSON.stringify(message), e.detail.elt)
})

document.body.addEventListener("htmx:wsClose", function(e) {
    console.log(e)
    const message = {
        action: "leave_lobby",
        lobbyID: lobbyID.dataset["lobby_id"]
    }

    e.detail.socketWrapper.send(JSON.stringify(message), e.detail.elt)
})

document.body.addEventListener("htmx:afterSwap", function(e) {
    console.log(e)
})
