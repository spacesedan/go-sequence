/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./views/**/*.{html, js}", "./internal/partials/*.templ"],
    theme: {
        extend: {
            gridTemplateRows: {
                '10': 'repeat(10, minmax(0, 1fr))'
            }
        },
    },
    plugins: [],
}

