const numOfPlayersInput = document.querySelector<HTMLInputElement>("#num_of_players")
const maxHandSizeInput = document.querySelector<HTMLInputElement>("#max_hand_size")
const createLobbyForm = document.querySelector<HTMLFormElement>("#create-lobby-form")

createLobbyForm?.addEventListener('submit', function(e) {
    e.preventDefault()
    switch (true) {
        case !numOfPlayersInput!.value:
            numOfPlayersInput?.focus()
            return
        case !maxHandSizeInput!.value:
            maxHandSizeInput?.focus()
            return
        case numOfPlayersInput!.value !== "" && maxHandSizeInput!.value !== "":
            //@ts-ignore
            htmx.ajax('POST', `/lobby/create?num_of_players=${numOfPlayersInput!.value}&max_hand_size=${maxHandSizeInput!.value}`, "")
            numOfPlayersInput!.value = ""
            maxHandSizeInput!.value = ""
            return
    }

})

