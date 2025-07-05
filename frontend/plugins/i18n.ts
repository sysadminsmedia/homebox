import type { CompileError, MessageContext } from "vue-i18n";
import { createI18n } from "vue-i18n";
import { IntlMessageFormat } from "intl-messageformat";

export default defineNuxtPlugin(({ vueApp }) => {
  function checkDefaultLanguage() {
    let matched = null;
    const languages = Object.getOwnPropertyNames(messages());
    const matching = navigator.languages.filter(lang => languages.some(l => l.toLowerCase() === lang.toLowerCase()));
    if (matching.length > 0) {
      matched = matching[0];
    }
    if (!matched) {
      languages.forEach(lang => {
        const languagePartials = navigator.language.split("-")[0];
        if (lang.toLowerCase() === languagePartials) {
          matched = lang;
        }
      });
    }
    return matched;
  }
  const preferences = useViewPreferences();
  const i18n = createI18n({
    fallbackLocale: "en",
    globalInjection: true,
    legacy: false,
    locale: "en", // Force English only
    messageCompiler,
    messages: messages(),
    fallbackWarn: false,
    missingWarn: false,

    missing: (locale, key) => {  
      // Always return English translation if available, even if the key exists but is an empty string in the current locale
      const fallbackMessages = i18n.global.getLocaleMessage("en") || {};
      const fallbackMsg = getNested(fallbackMessages, key);
      if (fallbackMsg !== undefined && fallbackMsg !== "") {
        return fallbackMsg;
      }
      // Fallback to showing the raw key if nothing is found
      return key;
    },
    // Custom fallback for empty string values
    postTranslation: (result, key) => {
      if (result === "") {
        const fallbackMessages = i18n.global.getLocaleMessage("en") || {};
        const fallbackMsg = getNested(fallbackMessages, key);
        if (fallbackMsg !== undefined && fallbackMsg !== "") {
          return fallbackMsg;
        }
        return key;
      }
      return result;
    },
  });
  vueApp.use(i18n);

  return {
    provide: {
      i18nGlobal: i18n.global,
    },
  };
});

// Utility to resolve nested keys like 'components.app.create_modal.enter'
function getNested(obj: any, path: string) {
  return path.split('.').reduce((o, k) => (o && o[k] !== undefined ? o[k] : undefined), obj);
}

export const messages = () => {
  const messages: Record<string, any> = {};
  const modules = import.meta.glob("~//locales/**.json", { eager: true });
  for (const path in modules) {
    const key = path.slice(9, -5);
    messages[key] = modules[path];
  }
  return messages;
};

export const messageCompiler: (
  message: String | any,
  {
    locale,
    key,
    onError,
  }: {
    locale: any;
    key: any;
    onError: any;
  }
) => (ctx: MessageContext) => unknown = (message, { locale, key, onError }) => {
  if (typeof message === "string") {
    /**
     * You can tune your message compiler performance more with your cache strategy or also memoization at here
     */
    const formatter = new IntlMessageFormat(message, locale);
    return (ctx: MessageContext) => {
      return formatter.format(ctx.values);
    };
  } else {
    /**
     * for AST.
     * If you would like to support it,
     * You need to transform locale messages such as `json`, `yaml`, etc. with the bundle plugin.
     */
    onError && onError(new Error("not support for AST") as CompileError);
    return () => key;
  }
};