import type { CompileError, MessageContext } from "vue-i18n";
import { createI18n } from "vue-i18n";
import { IntlMessageFormat } from "intl-messageformat";

export default defineNuxtPlugin(async ({ vueApp }) => {
  async function checkDefaultLanguage() {
    let matched = null;
    const languages = Object.getOwnPropertyNames(await messages());
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
    locale: preferences.value.language || await checkDefaultLanguage() || "en",
    messageCompiler,
    messages: await messages(),
  });
  vueApp.use(i18n);
});

export const messages = async () => {
  const messages: Record<string, any> = {};
  // const modules = import.meta.glob("~//locales/**.json", { eager: true });
  // for (const path in modules) {
  //   const key = path.slice(9, -5);
  //   messages[key] = modules[path];
  // }
  console.log('Fetching translations...');
  const en = await (await fetch('https://raw.githubusercontent.com/sysadminsmedia/homebox/refs/heads/main/frontend/locales/en.json')).json();
  console.log('Fetched translations.');
  messages['en'] = en;
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
