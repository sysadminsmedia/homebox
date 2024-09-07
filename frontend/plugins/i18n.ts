import type { CompileError, MessageCompiler, MessageContext } from "vue-i18n";
import { createI18n } from "vue-i18n";
import { IntlMessageFormat } from "intl-messageformat";
import { messages } from "./messages";

export default defineNuxtPlugin(({ vueApp }) => {
  function checkDefaultLanguage() {
    let matched = null;
    const languages = Object.getOwnPropertyNames(messages());
    languages.forEach(lang => {
      navigator.languages.forEach(language => {
        if (lang === language) {
          matched = lang;
        }
      });
    });
    if (!matched) {
      languages.forEach(lang => {
        const languagePartials = navigator.language.split("-")[0];
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
