import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Returns either '#000' or '#fff' depending on which has better contrast with the given background color.
 * Accepts hex (#RRGGBB or #RGB) or rgb(a) strings.
 */
export function getContrastTextColor(bgColor: string): string {
  let r = 0;
  let g = 0;
  let b = 0;
  if (bgColor.startsWith("#")) {
    let hex = bgColor.slice(1);
    if (hex.length === 3) {
      hex = hex
        .split("")
        .map(x => x + x)
        .join("");
    }
    r = parseInt(hex.slice(0, 2), 16);
    g = parseInt(hex.slice(2, 4), 16);
    b = parseInt(hex.slice(4, 6), 16);
  } else if (bgColor.startsWith("rgb")) {
    const match = bgColor.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
    if (match) {
      r = parseInt(match[1]);
      g = parseInt(match[2]);
      b = parseInt(match[3]);
    }
  }
  // Calculate luminance
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
  return luminance > 0.5 ? "#000" : "#fff";
}
