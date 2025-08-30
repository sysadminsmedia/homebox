type ColorType = "hsla";

export type VarOptions = {
  type: ColorType;
  transparency?: number;
  apply?: (value: string) => string;
};

export type Breakpoints = {
  sm: boolean;
  md: boolean;
  lg: boolean;
  xl: boolean;
  xxl: boolean;
};

export function useBreakpoints(): Breakpoints {
  const breakpoints: Breakpoints = reactive({
    sm: false,
    md: false,
    lg: false,
    xl: false,
    xxl: false,
  });

  const updateBreakpoints = () => {
    breakpoints.sm = window.innerWidth < 640;
    breakpoints.md = window.innerWidth >= 640;
    breakpoints.lg = window.innerWidth >= 768;
    breakpoints.xl = window.innerWidth >= 1024;
    breakpoints.xxl = window.innerWidth >= 1280;
  };

  onMounted(() => {
    updateBreakpoints();
    window.addEventListener("resize", updateBreakpoints);
  });

  onUnmounted(() => {
    window.removeEventListener("resize", updateBreakpoints);
  });

  return breakpoints;
}

class ThemeObserver {
  private static instance?: ThemeObserver;
  private readonly observer: MutationObserver;

  private fns: (() => void)[] = [];

  private constructor() {
    this.observer = new MutationObserver(mutations => {
      mutations.forEach(mutation => {
        if (mutation.attributeName === "data-theme") {
          this.fire();
        }
      });
    });

    const html = document.querySelector("html");
    if (!html) {
      throw new Error("No html element found");
    }

    this.observer.observe(html, { attributes: true });
  }

  public static getInstance() {
    if (!ThemeObserver.instance) {
      ThemeObserver.instance = new ThemeObserver();
    }

    return ThemeObserver.instance;
  }

  private fire() {
    this.fns.forEach(fn => fn());
  }

  public add(fn: () => void) {
    this.fns.push(fn);
  }

  public remove(fn: () => void) {
    this.fns = this.fns.filter(f => f !== fn);
  }
}

export function useCssVar(name: string, options?: VarOptions) {
  if (!options) {
    options = {
      type: "hsla",
      transparency: 1,
      apply: undefined,
    };
  }

  const cssVal = ref(getComputedStyle(document.documentElement).getPropertyValue(name).trim());
  const update = () => {
    cssVal.value = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  };

  ThemeObserver.getInstance().add(update);
  onUnmounted(() => {
    ThemeObserver.getInstance().remove(update);
  });

  switch (options.type) {
    case "hsla": {
      return computed(() => {
        if (!document) {
          return "";
        }

        let val = cssVal.value.trim().split(" ").join(", ");
        if (options?.transparency) {
          val += `, ${options.transparency}`;
        }

        return `hsla(${val})`;
      });
    }
  }
}
