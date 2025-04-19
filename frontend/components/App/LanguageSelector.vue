<script setup lang="ts">
  import { computed } from "vue";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Label } from "@/components/ui/label";
  import { fmtDate } from "~~/composables/use-formatters";
  import { useViewPreferences } from "~~/composables/use-preferences";

  defineProps({
    includeText: {
      type: Boolean,
      default: true,
    },
  });

  const preferences = useViewPreferences();

  function setLanguage(lang: string) {
    preferences.value.language = lang;
  }

  const dateExample = computed(() => {
    return fmtDate(new Date(Date.now() - 15 * 60000), "relative");
  });
</script>

<template>
  <div class="w-full" :class="{ 'p-5 pt-0': includeText }">
    <Label v-if="includeText" for="language"> {{ $t("profile.language") }} </Label>
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
    <p v-if="includeText" class="m-2 text-sm">
      {{ $t("profile.example") }}: {{ $t("global.created") }} {{ dateExample }}
    </p>
  </div>
</template>
