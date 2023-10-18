const chatForm = document.querySelector<HTMLFormElement>("#chat-form")
const chatInput = document.querySelector<HTMLTextAreaElement>("#chat-input")
const username = document.querySelector<HTMLDivElement>("#username")?.dataset["username"]
const lobbyId = document.querySelector<HTMLDivElement>("#lobby-id")?.dataset["lobbyId"]

document.body.addEventListener("htmx:wsOpen", function(e) {
    const message = {
        action: "join_lobby",
        username: username,
        lobby_id: lobbyId
    }
    //@ts-ignore
    e.detail.socketWrapper.send(JSON.stringify(message), e.detail.elt)
})

document.body.addEventListener("htmx:wsClose", function(e) {
    const message = {
        action: "left",
        username: username,
        lobby_id: lobbyId
    }
    //@ts-ignore
    e.detail.socketWrapper.send(JSON.stringify(message), e.detail.elt)

})

chatInput?.addEventListener("keydown", function(e) {

    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        if (!chatInput?.value) return false
        chatInput?.addEventListener("htmx:wsConfigSend", function(e) {
            console.log(e);
            //@ts-ignore
            e.detail.parameters = {
                action: "chat-message",
                message: chatInput!.value,
                username: username
            }

        })
    }
})

document.body.addEventListener('htmx:wsAfterSend', function() {
    chatInput!.value = ""
})

