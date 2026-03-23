/// <reference types="unplugin-icons/types/vue" />

declare module "#app" {
  interface NuxtApp {
    $otelEnabled: boolean;
  }
}

declare module "vue" {
  interface ComponentCustomProperties {
    $otelEnabled: boolean;
  }
}
export {};
