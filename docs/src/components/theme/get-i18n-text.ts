import type { StarlightRouteData } from '@astrojs/starlight/route-data'
import config from 'virtual:starlight/user-config'

export function getI18nText(
  value: string | Record<string, string>,
  route: StarlightRouteData,
): string {
  if (typeof value === 'string') {
    return value
  }

  const { lang, locale } = route

  // Try `lang` first, which is a BCP-47 language tag
  if (value[lang]) {
    return value[lang]
  }

  // Use the same "root" convention as the Starlight docs to get the default value.
  // https://github.com/withastro/starlight/blob/9d3ba179c5d524c1c61d771ceb1a7b4e754bee16/docs/src/content/docs/guides/i18n.mdx?plain=1#L72-L75
  if (
    config.defaultLocale.lang === lang ||
    config.defaultLocale.locale === locale
  ) {
    if (value['root']) {
      return value['root']
    }
  }

  const message =
    `[starlight-theme-nova] Unable to find the translation for language "${lang}".\n` +
    `Are you using the correct BCP-47 language tag ` +
    `(e.g. "en", "ar", or "zh-CN") in the following configuration?\n` +
    JSON.stringify(value, null, 2) +
    '\n'

  // Fallback to the `locale`, which is a base path at which a language is
  // served. This is not of the same format as the `lang` BCP-47 language tag.
  if (locale && value[locale]) {
    console.warn(message)
    return value[locale]
  }

  throw new Error(message)
}
