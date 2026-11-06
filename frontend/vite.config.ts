import react from '@vitejs/plugin-react';
import {defineConfig} from 'vitest/config';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    strictPort: true,
    proxy: {'/api': 'http://127.0.0.1:6666'},
  },
  test: {
    environment: 'jsdom',
    server: {deps: {inline: ['@gravity-ui/uikit']}},
    setupFiles: ['./src/test-setup.ts'],
    restoreMocks: true,
    clearMocks: true,
  },
});
