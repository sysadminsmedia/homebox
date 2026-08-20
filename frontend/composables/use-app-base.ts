/**
 * Returns the app's base URL, injected by the Go backend at runtime
 * via window.__HOMEBOX_BASE__ in the served index.html.
 *
 * Named with "use" prefix to follow Nuxt composable conventions (auto-imported
 * from composables/), though it is not reactive — the base path never changes
 * during the lifetime of the app.
 *
 * Note: This composable requires a browser environment (window). It returns "/"
 * if window is unavailable. SSR is not supported.
 */
export function useAppBase(): string {
  if (typeof window === "undefined") return "/";
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return (window as any).__HOMEBOX_BASE__ || "/";
}
