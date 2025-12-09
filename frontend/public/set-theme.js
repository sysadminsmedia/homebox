try {
  console.log("Setting theme");
  const preferences = JSON.parse(localStorage.getItem("homebox/preferences/location"));
  if (preferences) {
    const theme = preferences.theme;
    const darkMode = preferences.darkMode || "auto";
    
    if (theme) {
      document.documentElement.setAttribute("data-theme", theme);
      document.documentElement.classList.add("theme-" + theme);
      
      // Ensure homebox class is present for homebox theme dark mode CSS
      if (theme === "homebox") {
        document.documentElement.classList.add("homebox");
      } else {
        document.documentElement.classList.remove("homebox");
      }
      
      // Apply dark mode only for homebox theme
      if (theme === "homebox") {
        let shouldBeDark = false;
        
        if (darkMode === "auto") {
          // Check system preference
          shouldBeDark = window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
        } else if (darkMode === "dark") {
          shouldBeDark = true;
        }
        
        if (shouldBeDark) {
          document.documentElement.classList.add("dark");
          document.documentElement.setAttribute("data-dark-mode", "dark");
        } else {
          document.documentElement.classList.remove("dark");
          document.documentElement.removeAttribute("data-dark-mode");
        }
      } else {
        document.documentElement.classList.remove("dark");
        document.documentElement.removeAttribute("data-dark-mode");
      }
    }
  }
} catch (e) {
  console.error("Failed to set theme", e);
}
