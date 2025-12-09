<script setup lang="ts">
  import type { Component } from "vue";
  import { useI18n } from "vue-i18n";
  import { useTheme } from "~/composables/use-theme";
  import type { DarkModePreference } from "~/composables/use-preferences";
  import MdiWeatherSunny from "~icons/mdi/weather-sunny";
  import MdiWeatherNight from "~icons/mdi/weather-night";
  import MdiBrightnessAuto from "~icons/mdi/brightness-auto";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "@/components/ui/dropdown-menu";
  import { SidebarMenuButton } from "@/components/ui/sidebar";

  const { t } = useI18n();
  const { darkMode, setDarkMode } = useTheme();

  const darkModeOptions: {
    value: DarkModePreference;
    label: () => string;
    icon: Component;
  }[] = [
    {
      value: "auto",
      label: () => t("menu.dark_mode_system"),
      icon: MdiBrightnessAuto,
    },
    {
      value: "light",
      label: () => t("menu.dark_mode_light"),
      icon: MdiWeatherSunny,
    },
    {
      value: "dark",
      label: () => t("menu.dark_mode_dark"),
      icon: MdiWeatherNight,
    },
  ];

  const currentOption = computed(() => {
    return darkModeOptions.find(opt => opt.value === darkMode.value) ?? darkModeOptions[0];
  });

  const handleSelect = (value: DarkModePreference) => {
    setDarkMode(value);
  };
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <SidebarMenuButton :tooltip="$t('menu.dark_mode')" class="text-sidebar-foreground">
        <component :is="currentOption?.icon" class="size-6 shrink-0 text-sidebar-foreground" />
        <span class="font-medium">{{ currentOption?.label() }}</span>
      </SidebarMenuButton>
    </DropdownMenuTrigger>
    <DropdownMenuContent class="z-40 min-w-[var(--reka-dropdown-menu-trigger-width)]">
      <DropdownMenuItem
        v-for="option in darkModeOptions"
        :key="option.value"
        class="group cursor-pointer text-lg"
        :class="{ 'bg-accent text-accent-foreground': darkMode === option.value }"
        @click="handleSelect(option.value)"
      >
        <component :is="option.icon" class="mr-2 size-5 shrink-0" />
        {{ option.label() }}
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
