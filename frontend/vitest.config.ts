import path from "node:path";
import { configDefaults, defineConfig } from "vitest/config";

export default defineConfig({
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    css: true,
    include: ["src/**/*.{test,spec}.{ts,tsx}"],
    exclude: [
      ...configDefaults.exclude,
      "e2e/**",
      "playwright-report/**",
      "test-results/**",
    ],
    coverage: {
      provider: "v8",
      reporter: ["text", "json"],
      reportsDirectory: "./coverage",
      exclude: [
        "node_modules/**",
        "dist/**",
        "e2e/**",
        "**/*.d.ts",
        "src/main.tsx",
        "*.config.{ts,js,mjs}",
        "test-build.mjs",
        "verify_*.{ts,js}",
      ],
      reportOnFailure: true,
    },
  },
});
