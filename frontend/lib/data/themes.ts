export type DaisyTheme =
  | "homebox"
  | "light"
  | "dark"
  | "cupcake"
  | "bumblebee"
  | "emerald"
  | "corporate"
  | "synthwave"
  | "retro"
  | "cyberpunk"
  | "valentine"
  | "halloween"
  | "garden"
  | "forest"
  | "aqua"
  | "lofi"
  | "pastel"
  | "fantasy"
  | "wireframe"
  | "black"
  | "luxury"
  | "dracula"
  | "cmyk"
  | "autumn"
  | "business"
  | "acid"
  | "lemonade"
  | "night"
  | "coffee"
  | "winter";

export type ThemeOption = {
  tag: string;
  value: DaisyTheme;
};

export const themes: ThemeOption[] = [
  {
    tag: "Homebox",
    value: "homebox",
  },
  {
    tag: "Garden",
    value: "garden",
  },
  {
    tag: "Light",
    value: "light",
  },
  {
    tag: "Cupcake",
    value: "cupcake",
  },
  {
    tag: "Bumblebee",
    value: "bumblebee",
  },
  {
    tag: "Emerald",
    value: "emerald",
  },
  {
    tag: "Corporate",
    value: "corporate",
  },
  {
    tag: "Synthwave",
    value: "synthwave",
  },
  {
    tag: "Retro",
    value: "retro",
  },
  {
    tag: "Cyberpunk",
    value: "cyberpunk",
  },
  {
    tag: "Valentine",
    value: "valentine",
  },
  {
    tag: "Halloween",
    value: "halloween",
  },
  {
    tag: "Forest",
    value: "forest",
  },
  {
    tag: "Aqua",
    value: "aqua",
  },
  {
    tag: "Lofi",
    value: "lofi",
  },
  {
    tag: "Pastel",
    value: "pastel",
  },
  {
    tag: "Fantasy",
    value: "fantasy",
  },
  {
    tag: "Wireframe",
    value: "wireframe",
  },
  {
    tag: "Black",
    value: "black",
  },
  {
    tag: "Luxury",
    value: "luxury",
  },
  {
    tag: "Dracula",
    value: "dracula",
  },
  {
    tag: "Cmyk",
    value: "cmyk",
  },
  {
    tag: "Autumn",
    value: "autumn",
  },
  {
    tag: "Business",
    value: "business",
  },
  {
    tag: "Acid",
    value: "acid",
  },
  {
    tag: "Lemonade",
    value: "lemonade",
  },
  {
    tag: "Night",
    value: "night",
  },
  {
    tag: "Coffee",
    value: "coffee",
  },
  {
    tag: "Winter",
    value: "winter",
  },
];

export const darkThemes: DaisyTheme[] = [
  "synthwave",
  "retro",
  "cyberpunk",
  "valentine",
  "halloween",
  "forest",
  "aqua",
  "black",
  "luxury",
  "dracula",
  "business",
  "night",
  "coffee",
];
