import type { ComputedRef } from "vue";
import type { DaisyTheme } from "~~/lib/data/themes";

export interface UseTheme {
  theme: ComputedRef<DaisyTheme>;
  setTheme: (theme: DaisyTheme) => void;
}

const themeRef = ref<DaisyTheme>("garden");

export function useTheme(): UseTheme {
  const preferences = useViewPreferences();
  themeRef.value = preferences.value.theme;

  const setTheme = (newTheme: DaisyTheme) => {
    preferences.value.theme = newTheme;

    if (htmlEl) {
      htmlEl.value?.setAttribute("data-theme", newTheme);
      // FIXME: this is a hack to remove the theme class from the html element
      htmlEl.value?.classList.remove(...themes);
      htmlEl.value?.classList.add("theme-" + newTheme);
    }

    themeRef.value = newTheme;
  };

  const htmlEl = ref<HTMLElement | null>();

  onMounted(() => {
    if (htmlEl.value) {
      return;
    }

    htmlEl.value = document.querySelector("html");
  });

  const theme = computed(() => {
    return themeRef.value;
  });

  return { theme, setTheme };
}

export function useIsDark() {
  const theme = useTheme();

  const darkthemes = [
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

  return computed(() => {
    return darkthemes.includes(theme.theme.value);
  });
}

export const themes = [
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
];
