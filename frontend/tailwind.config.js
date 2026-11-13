/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  safelist: [
    "dark",
    "theme-aqua",
    "theme-black",
    "theme-bumblebee",
    "theme-cmyk",
    "theme-corporate",
    "theme-cupcake",
    "theme-cyberpunk",
    "theme-dracula",
    "theme-emerald",
    "theme-fantasy",
    "theme-forest",
    "theme-garden",
    "theme-halloween",
    "theme-light",
    "theme-lofi",
    "theme-luxury",
    "theme-pastel",
    "theme-retro",
    "theme-synthwave",
    "theme-valentine",
    "theme-wireframe",
    "theme-autumn",
    "theme-business",
    "theme-acid",
    "theme-lemonade",
    "theme-night",
    "theme-coffee",
    "theme-winter",
  ],
  prefix: "",

  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        // daisyUI has no tokens for these; it derives them from base-content opacity.
        border: "hsl(var(--foreground) / 0.1)",
        input: "hsl(var(--foreground) / 0.2)",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        "background-accent": "hsl(var(--background-accent))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--foreground) / 0.7)",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        sidebar: {
          DEFAULT: "hsl(var(--sidebar-background))",
          foreground: "hsl(var(--sidebar-foreground))",
        },
        "sidebar-primary": {
          DEFAULT: "hsl(var(--sidebar-primary))",
          foreground: "hsl(var(--sidebar-primary-foreground))",
        },
        "sidebar-accent": {
          DEFAULT: "hsl(var(--sidebar-foreground) / 0.1)",
          foreground: "hsl(var(--sidebar-accent-foreground))",
        },
        "sidebar-border": "hsl(var(--sidebar-foreground) / 0.1)",
        "sidebar-ring": "hsl(var(--ring))",
      },
      borderRadius: {
        xl: "calc(var(--radius) + 4px)",
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      keyframes: {
        "accordion-down": {
          from: { height: 0 },
          to: { height: "var(--reka-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--reka-accordion-content-height)" },
          to: { height: 0 },
        },
        "collapsible-down": {
          from: { height: 0 },
          to: { height: "var(--reka-collapsible-content-height)" },
        },
        "collapsible-up": {
          from: { height: "var(--reka-collapsible-content-height)" },
          to: { height: 0 },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
        "collapsible-down": "collapsible-down 0.2s ease-in-out",
        "collapsible-up": "collapsible-up 0.2s ease-in-out",
      },
      typography: {
        DEFAULT: {
          css: {
            "--tw-prose-headings": "hsl(var(--foreground))",
            "--tw-prose-lead": "hsl(var(--foreground) / 0.7)",
            "--tw-prose-bold": "hsl(var(--foreground))",
            "--tw-prose-counters": "hsl(var(--foreground) / 0.7)",
            "--tw-prose-bullets": "hsl(var(--foreground) / 0.7)",
            "--tw-prose-hr": "hsl(var(--foreground) / 0.25)",
            "--tw-prose-quotes": "hsl(var(--foreground))",
            "--tw-prose-quote-borders": "hsl(var(--foreground) / 0.4)",
            "--tw-prose-captions": "hsl(var(--foreground) / 0.7)",
            "--tw-prose-kbd": "hsl(var(--foreground))",
            "--tw-prose-kbd-shadows": "hsl(var(--foreground) / 0.1)",
            "--tw-prose-code": "hsl(var(--foreground))",
            "--tw-prose-th-borders": "hsl(var(--foreground) / 0.25)",
            "--tw-prose-td-borders": "hsl(var(--foreground) / 0.2)",
            color: "hsl(var(--foreground))",
            a: {
              color: "hsl(var(--primary))",
              "&:hover": {
                color: "hsl(var(--primary) / 0.8)",
              },
            },
            h1: {
              color: "hsl(var(--foreground))",
            },
            h2: {
              color: "hsl(var(--foreground))",
            },
          },
        },
      },
    },
  },
  plugins: [require("@tailwindcss/aspect-ratio"), require("@tailwindcss/typography"), require("tailwindcss-animate")],
};
