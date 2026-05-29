import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: '../internal/web/static',
    emptyOutDir: true,
    chunkSizeWarningLimit: 600,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return undefined
          // CodeMirror ecosystem — only loaded by the lazy editor page (~504 kB).
          if (id.includes('@codemirror') || id.includes('@lezer')) {
            return 'vendor-codemirror'
          }
          // Markdown rendering pipeline — only loaded by the vault page.
          if (
            id.includes('react-markdown') ||
            id.includes('remark-') ||
            id.includes('rehype-') ||
            id.includes('unified') ||
            id.includes('unist-') ||
            id.includes('mdast-') ||
            id.includes('hast-') ||
            id.includes('micromark') ||
            id.includes('character-entities') ||
            id.includes('decode-named-character-reference')
          ) {
            return 'vendor-markdown'
          }
          // React runtime — shared across all pages.
          if (id.includes('react') || id.includes('react-dom') || id.includes('scheduler')) {
            return 'vendor-react'
          }
          // Remaining node_modules (cva, clsx, tailwind-merge, etc.) are small
          // enough to fall through to Vite's default chunking.
        },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
