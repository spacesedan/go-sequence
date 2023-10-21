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
            //@ts-ignore
            e.detail.parameters = {
                action: "chat_message",
                message: chatInput!.value,
                username: username,
            }

        })
    }
})

document.body.addEventListener('htmx:wsAfterSend', function() {
    chatInput!.value = ""
})

//@ts-ignore
htmx.onLoad(function(content) {
    const red = document.body.querySelector<HTMLDivElement>("#red")
    const blue = document.body.querySelector<HTMLDivElement>("#blue")
    const green = document.body.querySelector<HTMLDivElement>("#green")

    red?.addEventListener("click", function() {
        //@ts-ignore
        htmx.trigger("#red", "htmx:wsConfigSend", {})
    })

    red?.addEventListener("htmx:wsConfigSend", function(e) {
        //@ts-ignore
        e.detail.parameters = {
            action: "choose_color",
            message: "red",
            username: username,
            enabled: red.dataset["enabled"] === "true" ? true : false
        }

    })

    red?.addEventListener("htmx:wsError", function(e) {
        console.log(e);

    })


    blue?.addEventListener("click", function() {
        //@ts-ignore
        htmx.trigger("#blue", "htmx:wsConfigSend", {})
    })

    blue?.addEventListener("htmx:wsConfigSend", function(e) {
        //@ts-ignore
        e.detail.parameters = {
            action: "choose_color",
            message: "blue",
            username: username,
            enabled: blue.dataset["enabled"] === "true" ? true : false
        }

    })


    green?.addEventListener("click", function() {
        //@ts-ignore
        htmx.trigger("#green", "htmx:wsConfigSend", {})
    })

    green?.addEventListener("htmx:wsConfigSend", function(e) {

        //@ts-ignore
        e.detail.parameters = {
            action: "choose_color",
            message: "green",
            username: username,
            lobby_id: lobbyId,
            enabled: green.dataset["enabled"] === "true" ? true : false
        }

    })


})
