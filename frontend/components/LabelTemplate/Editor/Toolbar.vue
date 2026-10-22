<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import MdiFormatText from "~icons/mdi/format-text";
  import MdiQrcode from "~icons/mdi/qrcode";
  import MdiShapeRectanglePlus from "~icons/mdi/shape-rectangle-plus";
  import MdiMinus from "~icons/mdi/minus";
  import MdiDelete from "~icons/mdi/delete";
  import { Button } from "@/components/ui/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
    DropdownMenuSub,
    DropdownMenuSubTrigger,
    DropdownMenuSubContent,
    DropdownMenuSeparator,
    DropdownMenuLabel,
  } from "@/components/ui/dropdown-menu";
  import { LabelmakerContentType, type LabelmakerBarcodeFormatInfo } from "~~/lib/api/types/data-contracts";

  const { t } = useI18n();

  const emit = defineEmits<{
    addText: [];
    addBarcode: [format: string, contentSource: string];
    addShape: [type: "rect" | "line"];
    deleteSelected: [];
  }>();

  const { formats } = useBarcodeFormats();

  // Content type compatibility:
  // - "any": can encode URLs, special chars, anything
  // - "alphanumeric": letters, numbers, limited symbols (no URLs)
  // - "numeric": digits only (only asset_id if configured as numeric)
  type SourceContentType = "any" | "alphanumeric" | "numeric";

  interface BarcodeContentSource {
    value: string;
    labelKey: string;
    contentType: SourceContentType;
  }

  // Item barcode content sources with their content type requirements
  const itemBarcodeContentSources: BarcodeContentSource[] = [
    { value: "item_url", labelKey: "item_url", contentType: "any" },
    { value: "asset_id", labelKey: "asset_id", contentType: "numeric" }, // Usually numeric
    { value: "serial_number", labelKey: "serial_number", contentType: "alphanumeric" },
    { value: "model_number", labelKey: "model_number", contentType: "alphanumeric" },
    { value: "id", labelKey: "item_id", contentType: "any" }, // UUID has dashes
  ];

  // Location barcode content sources
  const locationBarcodeContentSources: BarcodeContentSource[] = [
    { value: "location_url", labelKey: "location_url", contentType: "any" },
    { value: "location_name", labelKey: "location_name", contentType: "alphanumeric" },
    { value: "location_path", labelKey: "location_full_path", contentType: "any" }, // Has > separators
  ];

  // Check if a source is compatible with a barcode format's content type
  function isSourceCompatible(source: BarcodeContentSource, format: LabelmakerBarcodeFormatInfo): boolean {
    // "any" format accepts all source types
    if (format.contentType === LabelmakerContentType.ContentTypeAny) {
      return true;
    }
    // "alphanumeric" format accepts alphanumeric and numeric sources
    if (format.contentType === LabelmakerContentType.ContentTypeAlphanumeric) {
      return source.contentType === "alphanumeric" || source.contentType === "numeric";
    }
    // "numeric" format only accepts numeric sources
    if (format.contentType === LabelmakerContentType.ContentTypeNumeric) {
      return source.contentType === "numeric";
    }
    return true;
  }

  // Get compatible item sources for a format
  function getCompatibleItemSources(format: LabelmakerBarcodeFormatInfo): BarcodeContentSource[] {
    return itemBarcodeContentSources.filter(source => isSourceCompatible(source, format));
  }

  // Get compatible location sources for a format
  function getCompatibleLocationSources(format: LabelmakerBarcodeFormatInfo): BarcodeContentSource[] {
    return locationBarcodeContentSources.filter(source => isSourceCompatible(source, format));
  }
</script>

<template>
  <div class="flex flex-wrap items-center gap-2 rounded-lg border bg-card p-2">
    <Button variant="outline" size="sm" @click="emit('addText')">
      <MdiFormatText class="mr-1 size-4" />
      {{ $t("components.label_template.editor.toolbar.add_text") }}
    </Button>

    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <Button variant="outline" size="sm">
          <MdiQrcode class="mr-1 size-4" />
          {{ $t("components.label_template.editor.toolbar.add_barcode") }}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent class="w-56">
        <template v-for="format in formats" :key="format.format">
          <DropdownMenuSub>
            <DropdownMenuSubTrigger>
              {{ format.name }}
              <span v-if="format.is2D" class="ml-1 text-xs text-muted-foreground">(2D)</span>
            </DropdownMenuSubTrigger>
            <DropdownMenuSubContent class="max-h-80">
              <!-- Item Fields (filtered by compatibility) -->
              <template v-if="getCompatibleItemSources(format).length > 0">
                <DropdownMenuLabel class="text-xs text-muted-foreground">
                  {{ $t("components.label_template.editor.properties.item_fields") }}
                </DropdownMenuLabel>
                <DropdownMenuItem
                  v-for="source in getCompatibleItemSources(format)"
                  :key="source.value"
                  @click="emit('addBarcode', format.format, source.value)"
                >
                  {{ t(`components.label_template.editor.barcode_sources.${source.labelKey}`) }}
                </DropdownMenuItem>
              </template>
              <!-- Location Fields (filtered by compatibility) -->
              <template v-if="getCompatibleLocationSources(format).length > 0">
                <DropdownMenuSeparator v-if="getCompatibleItemSources(format).length > 0" />
                <DropdownMenuLabel class="text-xs text-muted-foreground">
                  {{ $t("components.label_template.editor.properties.location_fields") }}
                </DropdownMenuLabel>
                <DropdownMenuItem
                  v-for="source in getCompatibleLocationSources(format)"
                  :key="source.value"
                  @click="emit('addBarcode', format.format, source.value)"
                >
                  {{ t(`components.label_template.editor.barcode_sources.${source.labelKey}`) }}
                </DropdownMenuItem>
              </template>
              <!-- No compatible sources message -->
              <template
                v-if="
                  getCompatibleItemSources(format).length === 0 && getCompatibleLocationSources(format).length === 0
                "
              >
                <div class="px-2 py-1.5 text-xs text-muted-foreground">
                  {{ $t("components.label_template.editor.barcode_sources.no_compatible") }}
                </div>
              </template>
            </DropdownMenuSubContent>
          </DropdownMenuSub>
        </template>
      </DropdownMenuContent>
    </DropdownMenu>

    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <Button variant="outline" size="sm">
          <MdiShapeRectanglePlus class="mr-1 size-4" />
          {{ $t("components.label_template.editor.toolbar.add_shape") }}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuItem @click="emit('addShape', 'rect')">
          <MdiShapeRectanglePlus class="mr-2 size-4" />
          Rectangle
        </DropdownMenuItem>
        <DropdownMenuItem @click="emit('addShape', 'line')">
          <MdiMinus class="mr-2 size-4" />
          Line
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>

    <DropdownMenuSeparator class="h-6 w-px bg-border" />

    <Button variant="destructive" size="sm" @click="emit('deleteSelected')">
      <MdiDelete class="mr-1 size-4" />
      Delete
    </Button>
  </div>
</template>
