<script setup lang="ts">
  import { computed } from "vue";
  import * as Locales from "date-fns/locale";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Label } from "@/components/ui/label";
  import { fmtDate } from "~~/composables/use-formatters";
  import { localeMeasurementSystem, measurementLocale } from "~~/composables/use-measurement";
  import { useViewPreferences } from "~~/composables/use-preferences";

  const preferences = useViewPreferences();

  const locales = Object.values(Locales).map(l => {
    return {
      code: l.code,
      name: new Intl.DisplayNames([preferences.value.language ?? "en-US"], { type: "language" }).of(l.code) ?? l.code,
      localName: new Intl.DisplayNames([l.code], { type: "language" }).of(l.code) ?? l.code,
    };
  });

  defineProps({
    expanded: {
      type: Boolean,
      default: true,
    },
  });

  function setLanguage(lang: string) {
    preferences.value.language = lang;
  }

  function setOverrideLocale(locale: string | undefined) {
    if (locale === undefined || locale === preferences.value.overrideFormatLocale) {
      preferences.value.overrideFormatLocale = undefined;
    } else {
      preferences.value.overrideFormatLocale = locale;
    }
  }

  // "auto" stands in for a null preference, which follows the locale
  const measurementSystem = computed({
    get: () => preferences.value.measurementSystem ?? "auto",
    set: (value: string) => {
      preferences.value.measurementSystem = value === "metric" || value === "imperial" ? value : null;
    },
  });

  const localeSystem = computed(() => {
    // read the override so the label updates when it changes
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const locale = preferences.value.overrideFormatLocale;
    return localeMeasurementSystem(measurementLocale());
  });

  const dateExample = computed(() => {
    // hack to force vue to update
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const locale = preferences.value.overrideFormatLocale;
    return fmtDate(new Date(Date.now() - 15 * 60000), "relative");
  });
</script>

<template>
  <div class="w-full" :class="{ 'p-5 pt-0': expanded }">
    <Label v-if="expanded" for="language"> {{ $t("profile.language") }} </Label>
    <Select
      id="language"
      v-model="$i18n.locale"
      @update:model-value="
        event => {
          setLanguage(event as string);
        }
      "
    >
      <SelectTrigger>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem v-for="lang in $i18n.availableLocales" :key="lang" :value="lang">
          {{ $t(`languages.${lang}`) }} ({{ $t(`languages.${lang}`, 1, { locale: lang }) }})
        </SelectItem>
      </SelectContent>
    </Select>
    <template v-if="expanded">
      <Label for="overrideLocale"> {{ $t("profile.override_locale") }} </Label>
      <Select
        id="overrideLocale"
        :model-value="preferences.overrideFormatLocale"
        @update:model-value="
          val => {
            setOverrideLocale(val?.toString());
          }
        "
      >
        <SelectTrigger>
          <SelectValue :placeholder="$t('profile.no_override')" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="locale in locales" :key="locale.code" :value="locale.code">
            {{ locale.name }} - ({{ locale.localName }})
          </SelectItem>
        </SelectContent>
      </Select>
      <p class="m-2 text-sm">{{ $t("profile.example") }}: {{ $t("global.created") }} {{ dateExample }}</p>
      <Label for="measurementSystem"> {{ $t("profile.measurement_system") }} </Label>
      <Select id="measurementSystem" v-model="measurementSystem">
        <SelectTrigger>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="auto">
            {{ $t("profile.measurement_auto", { system: $t(`profile.measurement_${localeSystem}`) }) }}
          </SelectItem>
          <SelectItem value="metric">{{ $t("profile.measurement_metric") }}</SelectItem>
          <SelectItem value="imperial">{{ $t("profile.measurement_imperial") }}</SelectItem>
        </SelectContent>
      </Select>
    </template>
  </div>
</template>
