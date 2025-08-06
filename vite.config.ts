import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import path from "path"
import { defineConfig } from "vite"

export default defineConfig({
    plugins: [
        react({
            babel: {
                plugins: ["babel-plugin-react-compiler", {}],
            },
        }),
        tailwindcss(),
    ],
    root: "ui",
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./ui/src"),
            proto: path.resolve(__dirname, "./proto"),
        },
    },
    server: {
        host: true,
    },
    build: {
        outDir: "../gapi/src/server/ui",
        emptyOutDir: true,
    },
})
