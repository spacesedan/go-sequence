//@ts-ignore
htmx.onLoad(function(content) {
    const modal = document.querySelector<HTMLDivElement>("#modal")
    const modalUnderlay = document.querySelector<HTMLDivElement>("#modal-underlay")
    const closeModalBtn = document.querySelector<HTMLButtonElement>("#modal-btn")

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
})


