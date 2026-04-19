/**
 * Use the base URL of the app.
 * @returns `config.app.baseURL` of Nuxt config
 */
export function useAppBase(): string {
  const config = useRuntimeConfig();
  return config.app.baseURL;
}
