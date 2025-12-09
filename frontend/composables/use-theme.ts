import type { ComputedRef } from "vue";
import type { DaisyTheme } from "~~/lib/data/themes";
import { useMediaQuery } from "@vueuse/core";
import type { DarkModePreference } from "./use-preferences";

export interface UseTheme {
  theme: ComputedRef<DaisyTheme>;
  setTheme: (theme: DaisyTheme) => void;
  darkMode: ComputedRef<DarkModePreference>;
  setDarkMode: (mode: DarkModePreference) => void;
}

const themeRef = ref<DaisyTheme>("garden");

export function useTheme(): UseTheme {
  const preferences = useViewPreferences();
  themeRef.value = preferences.value.theme;

  const isSystemDark = useMediaQuery("(prefers-color-scheme: dark)");

  const currentThemeClass = ref<string | null>(null);

  const applyThemeToDOM = (newTheme: DaisyTheme) => {
    if (!htmlEl.value) return;

    htmlEl.value.setAttribute("data-theme", newTheme);

    // Remove previous theme class if it exists
    if (currentThemeClass.value) {
      htmlEl.value.classList.remove(currentThemeClass.value);
    }

    // Add new theme class and track it
    const newThemeClass = "theme-" + newTheme;
    htmlEl.value.classList.add(newThemeClass);
    currentThemeClass.value = newThemeClass;

    // Ensure homebox class is present for homebox theme dark mode CSS
    if (newTheme === "homebox") {
      htmlEl.value.classList.add("homebox");
    } else {
      htmlEl.value.classList.remove("homebox");
    }

    // Apply dark mode (only works for homebox theme)
    applyDarkMode();
  };

  const setTheme = (newTheme: DaisyTheme) => {
    preferences.value.theme = newTheme;
    themeRef.value = newTheme;
    applyThemeToDOM(newTheme);
  };

  const setDarkMode = (mode: DarkModePreference) => {
    preferences.value.darkMode = mode;
    applyDarkMode();
  };

  const applyDarkMode = () => {
    if (!htmlEl.value) return;

    const currentTheme = preferences.value.theme;
    const darkMode = preferences.value.darkMode;

    // Only apply dark mode to homebox theme
    if (currentTheme !== "homebox") {
      htmlEl.value.classList.remove("dark");
      htmlEl.value.removeAttribute("data-dark-mode");
      return;
    }

    let shouldBeDark = false;

    if (darkMode === "auto") {
      shouldBeDark = isSystemDark.value;
    } else if (darkMode === "dark") {
      shouldBeDark = true;
    } else {
      shouldBeDark = false;
    }

    if (shouldBeDark) {
      htmlEl.value.classList.add("dark");
      htmlEl.value.setAttribute("data-dark-mode", "dark");
    } else {
      htmlEl.value.classList.remove("dark");
      htmlEl.value.removeAttribute("data-dark-mode");
    }
  };

  const htmlEl = ref<HTMLElement | null>();

  onMounted(() => {
    if (htmlEl.value) {
      return;
    }

    htmlEl.value = document.querySelector("html");

    // Initialize current theme class from existing class on HTML element
    if (htmlEl.value) {
      const existingThemeClass = Array.from(htmlEl.value.classList).find(cls => cls.startsWith("theme-"));
      if (existingThemeClass) {
        currentThemeClass.value = existingThemeClass;
      }
    }

    // Apply initial theme and dark mode
    applyThemeToDOM(preferences.value.theme);
  });

  watch(
    () => preferences.value.darkMode,
    () => {
      applyDarkMode();
    }
  );

  watch(isSystemDark, () => {
    if (preferences.value.darkMode === "auto") {
      applyDarkMode();
    }
  });

  watch(
    () => preferences.value.theme,
    newTheme => {
      if (themeRef.value !== newTheme) {
        themeRef.value = newTheme;
        applyThemeToDOM(newTheme);
      }
    }
  );

  const theme = computed(() => {
    return themeRef.value;
  });

  const darkMode = computed(() => {
    return preferences.value.darkMode;
  });

  return { theme, setTheme, darkMode, setDarkMode };
}

export function useIsThemeInList(list: DaisyTheme[]) {
  const theme = useTheme();

  return computed(() => {
    return list.includes(theme.theme.value);
  });
}
