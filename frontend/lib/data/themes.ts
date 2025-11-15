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
  label: string;
  value: DaisyTheme;
};

export const themes: ThemeOption[] = [
  {
    label: "Homebox",
    value: "homebox",
  },
  {
    label: "Garden",
    value: "garden",
  },
  {
    label: "Light",
    value: "light",
  },
  {
    label: "Cupcake",
    value: "cupcake",
  },
  {
    label: "Bumblebee",
    value: "bumblebee",
  },
  {
    label: "Emerald",
    value: "emerald",
  },
  {
    label: "Corporate",
    value: "corporate",
  },
  {
    label: "Synthwave",
    value: "synthwave",
  },
  {
    label: "Retro",
    value: "retro",
  },
  {
    label: "Cyberpunk",
    value: "cyberpunk",
  },
  {
    label: "Valentine",
    value: "valentine",
  },
  {
    label: "Halloween",
    value: "halloween",
  },
  {
    label: "Forest",
    value: "forest",
  },
  {
    label: "Aqua",
    value: "aqua",
  },
  {
    label: "Lofi",
    value: "lofi",
  },
  {
    label: "Pastel",
    value: "pastel",
  },
  {
    label: "Fantasy",
    value: "fantasy",
  },
  {
    label: "Wireframe",
    value: "wireframe",
  },
  {
    label: "Black",
    value: "black",
  },
  {
    label: "Luxury",
    value: "luxury",
  },
  {
    label: "Dracula",
    value: "dracula",
  },
  {
    label: "Cmyk",
    value: "cmyk",
  },
  {
    label: "Autumn",
    value: "autumn",
  },
  {
    label: "Business",
    value: "business",
  },
  {
    label: "Acid",
    value: "acid",
  },
  {
    label: "Lemonade",
    value: "lemonade",
  },
  {
    label: "Night",
    value: "night",
  },
  {
    label: "Coffee",
    value: "coffee",
  },
  {
    label: "Winter",
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
