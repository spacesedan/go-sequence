const userName = document.querySelector<HTMLHeadingElement>("#generated_username")!
const generateBtn = document.querySelector<HTMLInputElement>("#generate-btn")!

if (generateBtn) {
    generateBtn.addEventListener('click', function() {
        generateBtn.disabled = true
        generateBtn.style.display = 'none'
    })
}

userName.addEventListener('click', function() {
    navigator.clipboard.writeText(userName.outerText)
})
