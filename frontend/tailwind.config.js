import { config } from "dotenv";
config();

// check if DISABLE_DAISYUI is set to true in the environment
const isDisabled = process.env.DISABLE_DAISYUI === "true";

if (isDisabled) {
  console.log("DAISYUI DISABLED");
}

/** @type {import('tailwindcss').Config} */
module.exports = {
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
    "theme-dark",
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
    "theme-dim",
    "theme-nord",
    "theme-sunset",
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
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
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
          foreground: "hsl(var(--muted-foreground))",
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
    },
  },
  daisyui: {
    themes: [
      {
        homebox: {
          primary: "#5C7F67",
          secondary: "#ECF4E7",
          accent: "#FFDA56",
          neutral: "#2C2E27",
          "base-100": "#FFFFFF",
          info: "#3ABFF8",
          success: "#36D399",
          warning: "#FBBD23",
          error: "#F87272",
        },
      },
      "light",
      "dark",
      "cupcake",
      "bumblebee",
      "emerald",
      "corporate",
      "synthwave",
      "retro",
      "cyberpunk",
      "valentine",
      "halloween",
      "garden",
      "forest",
      "aqua",
      "lofi",
      "pastel",
      "fantasy",
      "wireframe",
      "black",
      "luxury",
      "dracula",
      "cmyk",
      "autumn",
      "business",
      "acid",
      "lemonade",
      "night",
      "coffee",
      "winter",
    ],
  },
  plugins: isDisabled
    ? [require("@tailwindcss/aspect-ratio"), require("@tailwindcss/typography"), require("tailwindcss-animate")]
    : [
        require("@tailwindcss/aspect-ratio"),
        require("@tailwindcss/typography"),
        require("daisyui"),
        require("tailwindcss-animate"),
      ],
};
