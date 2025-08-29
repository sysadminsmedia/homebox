import path from "path";
import { defineConfig } from "vite";

export default defineConfig({
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  test: {
    globalSetup: "./test/setup.ts",
    include: ["**/*.test.ts"],
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, ".."),
      "~~": path.resolve(__dirname, ".."),
    },
  },
});
