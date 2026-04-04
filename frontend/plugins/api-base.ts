/**
 * Injects the Nuxt runtime base URL into the API URL builder so that
 * `lib/api/` remains a pure TS layer with no composable dependencies.
 */

import { overrideParts } from "~~/lib/api/base/urls";

export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig();
  overrideParts("http://localhost.com", "/api/v1", config.app.baseURL);
});
