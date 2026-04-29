import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  plugins: [svelte()],
  // VITE_BASE=/bajet/ when building for GitHub Pages; default to / elsewhere
  base: process.env.VITE_BASE ?? '/',
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
