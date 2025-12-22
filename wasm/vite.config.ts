import { defineConfig } from "rolldown-vite";

export default defineConfig({
  root: __dirname,
  base: "./",
  server: {
    port: 5173,
  },
  build: {
    outDir: "dist",
  },
  bundler: "rolldown",
});
