<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import MdiFormatBold from "~icons/mdi/format-bold";
  import MdiFormatAlignLeft from "~icons/mdi/format-align-left";
  import MdiFormatAlignCenter from "~icons/mdi/format-align-center";
  import MdiFormatAlignRight from "~icons/mdi/format-align-right";
  import MdiFormatSize from "~icons/mdi/format-size";
  import MdiTextBoxPlus from "~icons/mdi/text-box-plus";
  import { Input } from "@/components/ui/input";
  import { Button } from "@/components/ui/button";
  import { Label } from "@/components/ui/label";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

  const { t } = useI18n();

  const props = defineProps<{
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    selectedObject: any;
  }>();

  const emit = defineEmits<{
    updateObject: [updates: Record<string, unknown>];
    insertField: [fieldName: string];
  }>();

  // Item data fields available for insertion
  const itemDataFields = [
    { field: "item_name", key: "item_name" },
    { field: "description", key: "description" },
    { field: "asset_id", key: "asset_id" },
    { field: "serial_number", key: "serial_number" },
    { field: "model_number", key: "model_number" },
    { field: "manufacturer", key: "manufacturer" },
    { field: "location", key: "location" },
    { field: "location_path", key: "location_path" },
    { field: "labels", key: "labels" },
    { field: "quantity", key: "quantity" },
    { field: "notes", key: "notes" },
    { field: "item_url", key: "item_url" },
  ];

  // Location data fields available for insertion
  const locationDataFields = [
    { field: "location_name", key: "location_name" },
    { field: "description", key: "location_description" },
    { field: "location_path", key: "location_full_path" },
    { field: "item_count", key: "item_count" },
    { field: "location_url", key: "location_url" },
  ];

  // Common fields (work for both item and location labels)
  const commonDataFields = [
    { field: "current_date", key: "current_date" },
    { field: "current_time", key: "current_time" },
  ];

  // Reactive properties
  const objectProps = computed(() => {
    if (!props.selectedObject) return null;

    const obj = props.selectedObject;
    const data = obj.data || {};

    return {
      fontSize: obj.fontSize || 16,
      fontWeight: String(obj.fontWeight || "normal"),
      fill: obj.fill?.toString() || "#000000",
      textAlign: obj.textAlign || "left",
      // Textbox-specific
      textboxWidth: Math.round(obj.width || 150),
      autofit: data.autofit || false,
      fixedWidth: data.fixedWidth || Math.round(obj.width || 150),
      fixedHeight: data.fixedHeight || null,
      maxFontSize: data.maxFontSize || obj.fontSize || 16,
    };
  });

  function handleFontSizeChange(value: string) {
    const numValue = parseInt(value, 10);
    if (!isNaN(numValue) && numValue > 0) {
      emit("updateObject", { fontSize: numValue });
    }
  }

  function handleFontWeightChange(pressed: boolean) {
    emit("updateObject", { fontWeight: pressed ? "bold" : "normal" });
  }

  function handleColorChange(event: Event) {
    const target = event.target as HTMLInputElement;
    emit("updateObject", { fill: target.value });
  }

  function handleTextAlignChange(align: string) {
    emit("updateObject", { textAlign: align });
  }

  function handleTextboxWidthChange(value: string) {
    const numValue = parseInt(value, 10);
    if (!isNaN(numValue) && numValue > 0) {
      emit("updateObject", { width: numValue });
    }
  }

  function handleAutofitChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const currentData = props.selectedObject?.data || {};
    emit("updateObject", {
      data: {
        ...currentData,
        autofit: target.checked,
        maxFontSize: currentData.maxFontSize || props.selectedObject?.fontSize || 16,
      },
    });
  }

  function handleFixedHeightChange(value: string) {
    const numValue = parseInt(value, 10);
    const currentData = props.selectedObject?.data || {};
    emit("updateObject", {
      data: {
        ...currentData,
        fixedHeight: isNaN(numValue) || numValue <= 0 ? null : numValue,
      },
    });
  }

  function handleFixedWidthChange(value: string) {
    const numValue = parseInt(value, 10);
    const currentData = props.selectedObject?.data || {};
    const newWidth = isNaN(numValue) || numValue <= 0 ? 150 : numValue;
    emit("updateObject", {
      width: newWidth,
      data: {
        ...currentData,
        fixedWidth: newWidth,
      },
    });
  }

  function handleInsertField(value: unknown) {
    if (value && typeof value === "string") {
      emit("insertField", value);
    }
  }
</script>

<template>
  <div v-if="objectProps" class="flex flex-wrap items-center gap-3 rounded-lg border bg-card p-2">
    <!-- Insert Data Field - Prominently first -->
    <div class="flex items-center gap-2">
      <MdiTextBoxPlus class="size-4 text-muted-foreground" />
      <Select model-value="" @update:model-value="handleInsertField">
        <SelectTrigger class="h-8 w-40">
          <SelectValue :placeholder="$t('components.label_template.editor.properties.insert_field')" />
        </SelectTrigger>
        <SelectContent class="max-h-80">
          <!-- Item Fields -->
          <div class="px-2 py-1.5 text-xs font-semibold text-muted-foreground">
            {{ $t("components.label_template.editor.properties.item_fields") }}
          </div>
          <SelectItem v-for="field in itemDataFields" :key="field.field" :value="field.field">
            {{ t(`components.label_template.editor.data_fields.${field.key}`) }}
          </SelectItem>
          <!-- Location Fields -->
          <div class="px-2 py-1.5 text-xs font-semibold text-muted-foreground">
            {{ $t("components.label_template.editor.properties.location_fields") }}
          </div>
          <SelectItem v-for="field in locationDataFields" :key="field.field" :value="field.field">
            {{ t(`components.label_template.editor.data_fields.${field.key}`) }}
          </SelectItem>
          <!-- Common Fields -->
          <div class="px-2 py-1.5 text-xs font-semibold text-muted-foreground">
            {{ $t("components.label_template.editor.properties.common_fields") }}
          </div>
          <SelectItem v-for="field in commonDataFields" :key="field.field" :value="field.field">
            {{ t(`components.label_template.editor.data_fields.${field.key}`) }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="h-6 w-px bg-border" />

    <!-- Font Size -->
    <div class="flex items-center gap-1">
      <MdiFormatSize class="size-4 text-muted-foreground" />
      <Input
        type="number"
        :model-value="objectProps.fontSize"
        class="h-8 w-16"
        min="6"
        max="72"
        @update:model-value="handleFontSizeChange(String($event))"
      />
    </div>

    <!-- Bold Toggle -->
    <Button
      :variant="objectProps.fontWeight === 'bold' ? 'default' : 'outline'"
      size="icon"
      class="size-8"
      aria-label="Toggle bold"
      @click="handleFontWeightChange(objectProps.fontWeight !== 'bold')"
    >
      <MdiFormatBold class="size-4" />
    </Button>

    <div class="h-6 w-px bg-border" />

    <!-- Text Alignment -->
    <div class="flex items-center">
      <Button
        :variant="objectProps.textAlign === 'left' ? 'default' : 'outline'"
        size="icon"
        class="size-8 rounded-r-none"
        aria-label="Align left"
        @click="handleTextAlignChange('left')"
      >
        <MdiFormatAlignLeft class="size-4" />
      </Button>
      <Button
        :variant="objectProps.textAlign === 'center' ? 'default' : 'outline'"
        size="icon"
        class="size-8 rounded-none border-x-0"
        aria-label="Align center"
        @click="handleTextAlignChange('center')"
      >
        <MdiFormatAlignCenter class="size-4" />
      </Button>
      <Button
        :variant="objectProps.textAlign === 'right' ? 'default' : 'outline'"
        size="icon"
        class="size-8 rounded-l-none"
        aria-label="Align right"
        @click="handleTextAlignChange('right')"
      >
        <MdiFormatAlignRight class="size-4" />
      </Button>
    </div>

    <div class="h-6 w-px bg-border" />

    <!-- Text Color -->
    <div class="flex items-center gap-1">
      <input
        type="color"
        :value="objectProps.fill"
        class="size-8 cursor-pointer rounded border"
        @input="handleColorChange"
      />
    </div>

    <div class="h-6 w-px bg-border" />

    <!-- Text Width -->
    <div class="flex items-center gap-1">
      <Label class="text-xs text-muted-foreground">W</Label>
      <Input
        type="number"
        :model-value="objectProps.textboxWidth"
        class="h-8 w-16"
        min="20"
        @update:model-value="handleTextboxWidthChange(String($event))"
      />
    </div>

    <!-- Autofit Popover -->
    <Popover>
      <PopoverTrigger as-child>
        <Button variant="outline" size="sm" :class="{ 'border-primary': objectProps.autofit }">
          {{ $t("components.label_template.editor.properties.autofit") }}
          <span v-if="objectProps.autofit" class="ml-1 size-2 rounded-full bg-primary" />
        </Button>
      </PopoverTrigger>
      <PopoverContent class="w-64" align="start">
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <Label for="autofit" class="text-sm">
              {{ $t("components.label_template.editor.properties.autofit") }}
            </Label>
            <input
              id="autofit"
              type="checkbox"
              :checked="objectProps.autofit"
              class="size-4"
              @change="handleAutofitChange"
            />
          </div>

          <template v-if="objectProps.autofit">
            <div class="grid grid-cols-2 gap-2">
              <div class="space-y-1">
                <Label for="fixedWidth" class="text-xs text-muted-foreground">
                  {{ $t("components.label_template.editor.properties.max_width") }}
                </Label>
                <div class="flex items-center gap-1">
                  <Input
                    id="fixedWidth"
                    type="number"
                    :model-value="objectProps.fixedWidth || ''"
                    :placeholder="$t('components.label_template.editor.properties.auto')"
                    class="h-7 text-xs"
                    min="20"
                    @update:model-value="handleFixedWidthChange(String($event))"
                  />
                  <span class="text-xs text-muted-foreground">px</span>
                </div>
              </div>
              <div class="space-y-1">
                <Label for="fixedHeight" class="text-xs text-muted-foreground">
                  {{ $t("components.label_template.editor.properties.max_height") }}
                </Label>
                <div class="flex items-center gap-1">
                  <Input
                    id="fixedHeight"
                    type="number"
                    :model-value="objectProps.fixedHeight || ''"
                    :placeholder="$t('components.label_template.editor.properties.auto')"
                    class="h-7 text-xs"
                    min="10"
                    @update:model-value="handleFixedHeightChange(String($event))"
                  />
                  <span class="text-xs text-muted-foreground">px</span>
                </div>
              </div>
            </div>
            <p class="text-xs text-muted-foreground">
              {{ $t("components.label_template.editor.properties.autofit_hint") }}
            </p>
          </template>
        </div>
      </PopoverContent>
    </Popover>
  </div>
</template>
