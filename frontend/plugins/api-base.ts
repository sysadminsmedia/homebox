import { overrideParts } from "~~/lib/api/base/urls";

// This plugin runs early (order: -20) to configure the API URL builder
// before any other plugin or composable makes API calls.
// It depends on window.__HOMEBOX_BASE__ which is set via an inline <script>
// in <head> (injected by the Go backend), so it's available before any
// module-type scripts execute.
export default defineNuxtPlugin({
  name: "api-base",
  order: -20,
  setup() {
    const base = useAppBase();
    // "http://localhost.com" is a dummy host required by the URL() constructor
    // in route(). It gets stripped from the output — only the path is used.
    overrideParts("http://localhost.com", "/api/v1", base);
  },
});
