import { defineConfig } from 'vitest/config';

export default defineConfig({
  resolve: {
    preserveSymlinks: true,
  },
  test: {
    include: ['src/**/*.test.ts'],
    globals: true,
  },
});
