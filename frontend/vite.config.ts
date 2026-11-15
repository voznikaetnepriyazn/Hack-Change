import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

const apiProxy = {
  '/api': {
    target: 'http://localhost:8080',
    changeOrigin: true,
    rewrite: (path: string) => path.replace(/^\/api/, ''),
  },
};

export default defineConfig({
  plugins: [react()],

  server: {
    port: 3000,
    proxy: apiProxy,
    host: '0.0.0.0',
    strictPort: true,
  },

  preview: {
    host: '0.0.0.0',
    port: 3000,
    strictPort: true,
    proxy: apiProxy,
  },

  test: {
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',

    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: [
        'src/**/*.d.ts',
        'src/main.tsx',
        'src/test/**',
        'src/**/*.test.{ts,tsx}',
      ],
      thresholds: {
        statements: 80,
        branches: 80,
        functions: 80,
        lines: 80,
      },
    },
  },
});
