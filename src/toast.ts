//@ts-ignore
htmx.onLoad(function(content) {
    const toast = document.querySelector("#toast")
    const toastUnderlay = document.querySelector("#toast-underlay")

    function closeToast() {
        toast?.classList.add("closing")
        toast?.addEventListener("animationend", function() {
            toast.remove()
        })
    }

    toastUnderlay?.addEventListener("click", closeToast)

    setTimeout(closeToast, 5000)
})
