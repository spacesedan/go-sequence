import "./modal.ts"

const userName = document.querySelector<HTMLHeadingElement>("#generated_username")!
const generateBtn = document.querySelector<HTMLInputElement>("#generate-btn")!

if (generateBtn) {
    generateBtn.addEventListener('click', function() {
        generateBtn.disabled = true
        generateBtn.style.display = 'none'
    })
}

if (userName) {
    userName.addEventListener('click', function() {
        navigator.clipboard.writeText(userName.outerText)
    })
}
