const userName = document.querySelector("#generated_username")
const generateBtn = document.querySelector("#generate-btn")


generateBtn.addEventListener('click', function(e) {
    e.target.disabled = true
    e.target.style.display = 'none'
})

userName.addEventListener('click', function(e) {
    navigator.clipboard.writeText(e.target.outerText)
})
