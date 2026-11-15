import react from '@vitejs/plugin-react';
import { defineConfig, loadEnv } from 'vitest/config';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '.', '');
  const apiProxy = {
    '/api': {
      target: env.BACKEND_PROXY_TARGET || 'http://localhost:8080',
      changeOrigin: true,
      rewrite: (path: string) => path.replace(/^\/api/, ''),
    },
  };

  return {
    plugins: [react()],
    server: {
      host: '0.0.0.0',
      port: 3000,
      strictPort: true,
      proxy: apiProxy,
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
  };
});
