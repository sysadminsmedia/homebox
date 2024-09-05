import type { CompileError, MessageCompiler, MessageContext } from "vue-i18n";
import { createI18n } from "vue-i18n";
import { IntlMessageFormat } from "intl-messageformat";

export default defineNuxtPlugin(({ vueApp }) => {
  function checkDefaultLanguage() {
    let matched = null;
    const languages = Object.getOwnPropertyNames(messages())
    languages.forEach(lang => {
      if (lang === navigator.language.replace('-', '_')) {
        matched = lang;
      }
    });
    if (!matched) {
      languages.forEach(lang => {
        const languagePartials = navigator.language.split('-')[0]
        if (lang === languagePartials) {
          matched = lang;
        }
      });
    }
    return matched;
  }
  const i18n = createI18n({
    legacy: false,
    globalInjection: true,
    locale: checkDefaultLanguage() || "en",
    fallbackLocale: "en",
    messageCompiler,
    messages: messages(),
  });
  vueApp.use(i18n);
});

export const messageCompiler: MessageCompiler = (message, { locale, key, onError }) => {
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

export const messages: Object = () => {
  let messages = {};
  const modules = import.meta.glob('~//locales/**.json', { eager: true });
  for (const path in modules) {
    const key = path.slice(9, -5);
    messages[key] = modules[path];
  }
  return messages;
};
