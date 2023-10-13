import { resolve } from 'path'
import { defineConfig } from 'vite';

export default defineConfig({
    server: {
        proxy: {
            '/': 'http://localhost:42069/',
        }
    },
    css: {
        transformer: 'lightningcss',
    },
    build: {
        // required to create a manifest file
        manifest: true,
        emptyOutDir: true,
        cssMinify: 'lightningcss',
        minify: true,
        rollupOptions: {
            output: {
                assetFileNames: '[ext]/[name][extname]',
                entryFileNames: 'js/[name].js'
            },
            // specify your input files here, as stated in Vite config https://vitejs.dev/config/#build-rollupoptions
            input: {
                main_layout: resolve(__dirname, 'src/main_layout.ts'),
                index: resolve(__dirname, 'src/index.ts'),
                create_lobby: resolve(__dirname, 'src/create_lobby.ts'),
                lobby: resolve(__dirname, 'src/lobby.ts')
            }
        }
    }
})
