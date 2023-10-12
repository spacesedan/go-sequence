

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

    lobbyForm?.addEventListener('submit', function(e) {
        e.preventDefault()

        if (lobbyIdInput?.value == "") {
            lobbyIdInput.classList.remove("border-gray-300")
            lobbyIdInput.style.borderColor = 'red'
            lobbyIdLabel!.innerText = "no lobby id"
            return
        }

        if (!lobbyIdInput?.value.match(lobbyIdRegex)) {
            lobbyIdInput!.classList.remove("border-gray-300")
            lobbyIdInput!.style.borderColor = 'red'
            lobbyIdLabel!.innerText = "invalid lobby id"
            return
        }

    })
})


