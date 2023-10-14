const chatForm = document.querySelector<HTMLFormElement>("#chat-form")
const chatInput = document.querySelector<HTMLInputElement>("#chat-input")

// document.body.addEventListener("htmx:wsOpen", function(e) {
//     const message = {
//         action: "join_lobby",
//     }
//     //@ts-ignore
//     e.detail.socketWrapper.send(JSON.stringify(message), e.detail.elt)
// })

type wsPayload = {
    action: string
    message: string
    username: string
}

chatForm?.addEventListener('submit', function() {
    chatForm.addEventListener('htmx:wsConfigSend', function(e) {
        //@ts-ignore
        e.detail.parameters = {
            action: "chat-message",
            message: chatInput!.value,
            username: ""
        }

    })
})

document.body.addEventListener('htmx:wsAfterSend', function() {
    chatInput!.value = ""
})

