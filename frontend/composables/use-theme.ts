import type { ComputedRef } from "vue";
import type { DaisyTheme } from "~~/lib/data/themes";

export interface UseTheme {
  theme: ComputedRef<DaisyTheme>;
  setTheme: (theme: DaisyTheme) => void;
}

export function useTheme(): UseTheme {
  const preferences = useViewPreferences();
  const theme = computed(() => preferences.value.theme);
  const htmlEl = ref<HTMLElement | null>(null);

  const applyThemeToDom = (newTheme: DaisyTheme) => {
    if (!htmlEl.value) {
      return;
    }

    htmlEl.value.setAttribute("data-theme", newTheme);

    const prefixedThemeClasses = Array.from(htmlEl.value.classList).filter(className => className.startsWith("theme-"));
    if (prefixedThemeClasses.length > 0) {
      htmlEl.value.classList.remove(...prefixedThemeClasses);
    }

    htmlEl.value.classList.remove(...themes);
    htmlEl.value.classList.add("theme-" + newTheme);
  };

  const setTheme = (newTheme: DaisyTheme) => {
    preferences.value.theme = newTheme;
  };

  onMounted(() => {
    htmlEl.value = document.querySelector("html");
    applyThemeToDom(theme.value);
  });

  watch(theme, newTheme => {
    applyThemeToDom(newTheme);
  });

  return { theme, setTheme };
}

export function useIsThemeInList(list: DaisyTheme[]) {
  const theme = useTheme();

  return computed(() => {
    return list.includes(theme.theme.value);
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
