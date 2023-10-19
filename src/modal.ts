

const lobbyIdRegex = /[A-Z0-9]{4}/

//@ts-ignore
htmx.onLoad(function(content) {
    const modal = document.querySelector<HTMLDivElement>("#modal")
    const modalUnderlay = document.querySelector<HTMLDivElement>("#modal-underlay")
    const closeModalBtn = document.querySelector<HTMLButtonElement>("#modal-btn")
    const lobbyIdInput = document.querySelector<HTMLInputElement>("#lobby-id")
    const lobbyIdLabel = document.querySelector<HTMLLabelElement>("#lobby-id-label")
    const lobbyForm = document.querySelector<HTMLFormElement>("#join-lobby-form")


    // closeModal
    function closeModal() {
        // add the closing animation to the modal element
        modal?.classList.add("closing")
        // wait for the animation to end then remove the element
        modal?.addEventListener("animationend", function() {
            modal.remove()
        })
    }

    closeModalBtn?.addEventListener('click', function() {
        closeModal()
    })

    modalUnderlay?.addEventListener("click", function() {
        closeModal()
    })

    lobbyForm?.addEventListener('submit', function() {

        switch (true) {
            case !lobbyIdInput?.value:
                console.log(1);
                lobbyIdInput!.classList.remove("border-gray-300")
                lobbyIdInput!.style.borderColor = 'red'
                lobbyIdLabel!.innerText = "no lobby id"
                lobbyIdInput!.innerText = ""
                return
            case !lobbyIdInput?.value.match(lobbyIdRegex):
                console.log(2);
                lobbyIdInput!.classList.remove("border-gray-300")
                lobbyIdInput!.style.borderColor = 'red'
                lobbyIdLabel!.innerText = "invalid lobby id"
                lobbyIdInput!.innerText = ""
                return
            default:
                //@ts-ignore
                htmx.ajax('POST', `/lobby/join?lobby_id=${lobbyIdInput?.value}`, { target: '#body', swap: 'beforeend' })
                lobbyIdInput!.innerText = ""
                closeModal()
                return
        }

        // if (lobbyIdInput?.value == "") {
        //     return
        // }
        //
        // if (!lobbyIdInput?.value.match(lobbyIdRegex)) {
        //     lobbyIdInput!.classList.remove("border-gray-300")
        //     lobbyIdInput!.style.borderColor = 'red'
        //     lobbyIdLabel!.innerText = "invalid lobby id"
        //     lobbyIdInput!.innerText = ""
        //     return
        // }



    })
})


