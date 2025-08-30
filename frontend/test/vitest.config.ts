export default async () => {
  const { defineConfig } = await import("vitest/config");
  const path = await import("path");

  return defineConfig({
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
};
