import path from "path";
import { defineConfig } from "vite";

export default defineConfig({
  // @ts-ignore
  test: {
    globalSetup: "./test/setup.ts",
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, ".."),
      "~~": path.resolve(__dirname, ".."),
    },
  },
});
