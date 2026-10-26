<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import {
    FIXED_VALUE_SOURCE,
    compatibleItemSources,
    compatibleLocationSources,
    isFixedValueSource,
    isNumericFormat,
  } from "./barcode-sources";

  const props = defineProps<{
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    selectedObject: any;
  }>();

  const emit = defineEmits<{
    updateObject: [updates: Record<string, unknown>];
  }>();

  const { t } = useI18n();
  const { formats } = useBarcodeFormats();

  const data = computed<Record<string, unknown>>(() => props.selectedObject?.data ?? {});
  const formatKey = computed(() => String(data.value.format ?? ""));
  const format = computed(() => formats.value?.find(f => f.format === formatKey.value) ?? null);
  const numeric = computed(() => isNumericFormat(format.value));

  const contentSource = computed(() => String(data.value.contentSource ?? ""));
  const isFixed = computed(() => isFixedValueSource(contentSource.value));
  const selectValue = computed(() => (isFixed.value ? FIXED_VALUE_SOURCE : contentSource.value));

  const itemSources = computed(() => (format.value ? compatibleItemSources(format.value) : []));
  const locationSources = computed(() => (format.value ? compatibleLocationSources(format.value) : []));

  function setSource(source: string) {
    emit("updateObject", { data: { ...data.value, contentSource: source } });
  }

  function handleSourceChange(value: unknown) {
    if (typeof value !== "string") return;
    if (value === FIXED_VALUE_SOURCE) {
      // Keep a literal the user already typed; otherwise start empty so the preview shows a placeholder.
      setSource(isFixed.value ? contentSource.value : "");
      return;
    }
    setSource(value);
  }

  function handleValueChange(value: string) {
    const cleaned = numeric.value ? value.replace(/\D+/g, "") : value.trim();
    setSource(cleaned);
  }
</script>

<template>
  <div class="space-y-3">
    <div class="space-y-1.5">
      <Label class="text-xs">{{ $t("components.label_template.editor.properties.barcode_format") }}</Label>
      <div class="text-sm">
        {{ format?.name ?? formatKey.toUpperCase() }}
        <span v-if="format?.description" class="text-xs text-muted-foreground"> · {{ format.description }}</span>
      </div>
    </div>

    <div class="space-y-1.5">
      <Label class="text-xs">{{ $t("components.label_template.editor.properties.barcode_source") }}</Label>
      <Select :model-value="selectValue" @update:model-value="handleSourceChange">
        <SelectTrigger class="h-8">
          <SelectValue />
        </SelectTrigger>
        <SelectContent class="max-h-80">
          <template v-if="itemSources.length > 0">
            <div class="px-2 py-1.5 text-xs font-semibold text-muted-foreground">
              {{ $t("components.label_template.editor.properties.item_fields") }}
            </div>
            <SelectItem v-for="source in itemSources" :key="source.value" :value="source.value">
              {{ t(`components.label_template.editor.barcode_sources.${source.labelKey}`) }}
            </SelectItem>
          </template>
          <template v-if="locationSources.length > 0">
            <div class="px-2 py-1.5 text-xs font-semibold text-muted-foreground">
              {{ $t("components.label_template.editor.properties.location_fields") }}
            </div>
            <SelectItem v-for="source in locationSources" :key="source.value" :value="source.value">
              {{ t(`components.label_template.editor.barcode_sources.${source.labelKey}`) }}
            </SelectItem>
          </template>
          <SelectItem :value="FIXED_VALUE_SOURCE">
            {{ $t("components.label_template.editor.barcode_sources.fixed_value_option") }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div v-if="isFixed" class="space-y-1.5">
      <Label class="text-xs" for="barcode-fixed-value">
        {{ $t("components.label_template.editor.properties.barcode_value") }}
      </Label>
      <Input
        id="barcode-fixed-value"
        :model-value="contentSource"
        :inputmode="numeric ? 'numeric' : 'text'"
        :pattern="numeric ? '[0-9]*' : undefined"
        class="h-8"
        @change="handleValueChange(($event.target as HTMLInputElement).value)"
      />
      <p class="text-xs text-muted-foreground">
        {{
          numeric
            ? $t("components.label_template.editor.barcode_sources.digits_only")
            : $t("components.label_template.editor.barcode_sources.fixed_value_hint")
        }}
      </p>
    </div>
  </div>
</template>
