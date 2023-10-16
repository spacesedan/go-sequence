/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./internal/**/*.templ", ],
    theme: {
        extend: {
            gridTemplateRows: {
                // Simple 10 row grid
                '10': 'repeat(10, minmax(0, 1fr))',
                // Lobby row config
                lobby_grid: '5vh 1fr',
            }
        },
    },
    plugins: [],
}

