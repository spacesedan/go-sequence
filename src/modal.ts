//@ts-ignore
htmx.onLoad(function(content) {
    const modal = document.querySelector<HTMLDivElement>("#modal")
    const modalUnderlay = document.querySelector<HTMLDivElement>("#modal-underlay")
    const closeModalBtn = document.querySelector<HTMLButtonElement>("#modal-btn")
    const lobbyIdInput = document.querySelector<HTMLInputElement>("#lobby-id")
    const joinLobbyBtn = document.querySelector<HTMLButtonElement>("#join-lobby-btn")
    const lobbyForm = document.querySelector<HTMLFormElement>("#join-lobby-form")

    console.log(modal);
    console.log(lobbyForm);



    // closeModal
    function closeModal() {
        // add the closing animation to the modal element
        modal?.classList.add("closing")
        // wait for the animation to end then remove the element
        modal?.addEventListener("animationend", function() {
            modal.remove()
        })
    }

    lobbyForm?.addEventListener('submit', function(e) {
        e.preventDefault()

        if (lobbyIdInput?.innerText == "") {
            console.log("HIT");

            lobbyIdInput.classList.remove("border-gray-300")
            lobbyIdInput.style.borderColor = 'red'
        }

    })




    closeModalBtn?.addEventListener('click', function() {
        closeModal()
    })

    modalUnderlay?.addEventListener("click", function() {
        closeModal()
    })
})


