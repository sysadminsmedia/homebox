<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import MdiCursorDefault from "~icons/mdi/cursor-default";
  import MdiFormatText from "~icons/mdi/format-text";
  import MdiQrcode from "~icons/mdi/qrcode";
  import MdiShapeRectanglePlus from "~icons/mdi/shape-rectangle-plus";
  import MdiMinus from "~icons/mdi/minus";
  import MdiImagePlus from "~icons/mdi/image-plus";
  import MdiAlignHorizontalLeft from "~icons/mdi/align-horizontal-left";
  import MdiAlignHorizontalCenter from "~icons/mdi/align-horizontal-center";
  import MdiAlignHorizontalRight from "~icons/mdi/align-horizontal-right";
  import MdiAlignVerticalTop from "~icons/mdi/align-vertical-top";
  import MdiAlignVerticalCenter from "~icons/mdi/align-vertical-center";
  import MdiAlignVerticalBottom from "~icons/mdi/align-vertical-bottom";
  import MdiUndo from "~icons/mdi/undo";
  import MdiRedo from "~icons/mdi/redo";
  import MdiDelete from "~icons/mdi/delete";
  import { toast } from "@/components/ui/sonner";
  import { Button } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
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

  export type AlignTarget = "left" | "center" | "right" | "top" | "middle" | "bottom";

  const props = defineProps<{
    hasSelection: boolean;
    canUndo: boolean;
    canRedo: boolean;
  }>();

  const emit = defineEmits<{
    addText: [];
    addBarcode: [format: string, contentSource: string];
    addShape: [type: "rect" | "line"];
    addImage: [file: File];
    align: [target: AlignTarget];
    undo: [];
    redo: [];
    deleteSelected: [];
  }>();

  const { t } = useI18n();
  const { formats } = useBarcodeFormats();

  // Content type compatibility between a barcode symbology and the data it encodes:
  // "any" = URLs and arbitrary text, "alphanumeric" = letters and digits, "numeric" = digits only.
  type SourceContentType = "any" | "alphanumeric" | "numeric";

  interface BarcodeContentSource {
    value: string;
    labelKey: string;
    contentType: SourceContentType;
  }

  const itemBarcodeContentSources: BarcodeContentSource[] = [
    { value: "item_url", labelKey: "item_url", contentType: "any" },
    { value: "asset_id", labelKey: "asset_id", contentType: "numeric" },
    { value: "serial_number", labelKey: "serial_number", contentType: "alphanumeric" },
    { value: "model_number", labelKey: "model_number", contentType: "alphanumeric" },
    { value: "id", labelKey: "item_id", contentType: "any" },
  ];

  const locationBarcodeContentSources: BarcodeContentSource[] = [
    { value: "location_url", labelKey: "location_url", contentType: "any" },
    { value: "location_name", labelKey: "location_name", contentType: "alphanumeric" },
    { value: "location_path", labelKey: "location_full_path", contentType: "any" },
  ];

  function isSourceCompatible(source: BarcodeContentSource, format: LabelmakerBarcodeFormatInfo): boolean {
    if (format.contentType === LabelmakerContentType.ContentTypeAny) return true;
    if (format.contentType === LabelmakerContentType.ContentTypeAlphanumeric) {
      return source.contentType === "alphanumeric" || source.contentType === "numeric";
    }
    if (format.contentType === LabelmakerContentType.ContentTypeNumeric) {
      return source.contentType === "numeric";
    }
    return true;
  }

  const compatibleItemSources = (format: LabelmakerBarcodeFormatInfo) =>
    itemBarcodeContentSources.filter(s => isSourceCompatible(s, format));
  const compatibleLocationSources = (format: LabelmakerBarcodeFormatInfo) =>
    locationBarcodeContentSources.filter(s => isSourceCompatible(s, format));

  // Image picking: small raster logos only; the data URL is stored inside the template JSON.
  const MAX_IMAGE_BYTES = 512 * 1024;
  const fileInput = ref<HTMLInputElement | null>(null);

  function pickImage() {
    fileInput.value?.click();
  }

  function onImagePicked(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = "";
    if (!file) return;
    if (file.size > MAX_IMAGE_BYTES) {
      toast.error(t("components.label_template.editor.tools.image_too_large", { max: "512 KB" }));
      return;
    }
    emit("addImage", file);
  }

  const alignTargets: { target: AlignTarget; icon: Component; key: string }[] = [
    { target: "left", icon: MdiAlignHorizontalLeft, key: "left" },
    { target: "center", icon: MdiAlignHorizontalCenter, key: "center" },
    { target: "right", icon: MdiAlignHorizontalRight, key: "right" },
    { target: "top", icon: MdiAlignVerticalTop, key: "top" },
    { target: "middle", icon: MdiAlignVerticalCenter, key: "middle" },
    { target: "bottom", icon: MdiAlignVerticalBottom, key: "bottom" },
  ];
</script>

<template>
  <TooltipProvider :delay-duration="300">
    <div
      role="toolbar"
      :aria-label="$t('components.label_template.editor.tools.title')"
      class="flex flex-row flex-wrap items-center gap-1 rounded-lg border bg-card p-1.5 lg:w-14 lg:flex-col lg:flex-nowrap lg:self-start"
    >
      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            class="size-10 bg-accent"
            :aria-label="$t('components.label_template.editor.tools.select')"
          >
            <MdiCursorDefault class="size-5" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">{{ $t("components.label_template.editor.tools.select") }}</TooltipContent>
      </Tooltip>

      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            class="size-10"
            :aria-label="$t('components.label_template.editor.toolbar.add_text')"
            @click="emit('addText')"
          >
            <MdiFormatText class="size-5" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">{{ $t("components.label_template.editor.toolbar.add_text") }}</TooltipContent>
      </Tooltip>

      <DropdownMenu>
        <Tooltip>
          <TooltipTrigger as-child>
            <DropdownMenuTrigger as-child>
              <Button
                variant="ghost"
                size="icon"
                class="size-10"
                :aria-label="$t('components.label_template.editor.toolbar.add_barcode')"
              >
                <MdiQrcode class="size-5" />
              </Button>
            </DropdownMenuTrigger>
          </TooltipTrigger>
          <TooltipContent side="right">{{ $t("components.label_template.editor.toolbar.add_barcode") }}</TooltipContent>
        </Tooltip>
        <DropdownMenuContent side="right" align="start" class="w-56">
          <template v-for="format in formats ?? []" :key="format.format">
            <DropdownMenuSub>
              <DropdownMenuSubTrigger>
                {{ format.name }}
                <span v-if="format.is2D" class="ml-1 text-xs text-muted-foreground">(2D)</span>
              </DropdownMenuSubTrigger>
              <DropdownMenuSubContent class="max-h-80">
                <template v-if="compatibleItemSources(format).length > 0">
                  <DropdownMenuLabel class="text-xs text-muted-foreground">
                    {{ $t("components.label_template.editor.properties.item_fields") }}
                  </DropdownMenuLabel>
                  <DropdownMenuItem
                    v-for="source in compatibleItemSources(format)"
                    :key="source.value"
                    @click="emit('addBarcode', format.format, source.value)"
                  >
                    {{ t(`components.label_template.editor.barcode_sources.${source.labelKey}`) }}
                  </DropdownMenuItem>
                </template>
                <template v-if="compatibleLocationSources(format).length > 0">
                  <DropdownMenuSeparator v-if="compatibleItemSources(format).length > 0" />
                  <DropdownMenuLabel class="text-xs text-muted-foreground">
                    {{ $t("components.label_template.editor.properties.location_fields") }}
                  </DropdownMenuLabel>
                  <DropdownMenuItem
                    v-for="source in compatibleLocationSources(format)"
                    :key="source.value"
                    @click="emit('addBarcode', format.format, source.value)"
                  >
                    {{ t(`components.label_template.editor.barcode_sources.${source.labelKey}`) }}
                  </DropdownMenuItem>
                </template>
                <DropdownMenuItem
                  v-if="compatibleItemSources(format).length === 0 && compatibleLocationSources(format).length === 0"
                  disabled
                >
                  {{ $t("components.label_template.editor.barcode_sources.no_compatible") }}
                </DropdownMenuItem>
              </DropdownMenuSubContent>
            </DropdownMenuSub>
          </template>
        </DropdownMenuContent>
      </DropdownMenu>

      <DropdownMenu>
        <Tooltip>
          <TooltipTrigger as-child>
            <DropdownMenuTrigger as-child>
              <Button
                variant="ghost"
                size="icon"
                class="size-10"
                :aria-label="$t('components.label_template.editor.toolbar.add_shape')"
              >
                <MdiShapeRectanglePlus class="size-5" />
              </Button>
            </DropdownMenuTrigger>
          </TooltipTrigger>
          <TooltipContent side="right">{{ $t("components.label_template.editor.toolbar.add_shape") }}</TooltipContent>
        </Tooltip>
        <DropdownMenuContent side="right" align="start">
          <DropdownMenuItem @click="emit('addShape', 'rect')">
            <MdiShapeRectanglePlus class="mr-2 size-4" />
            {{ $t("components.label_template.editor.tools.rectangle") }}
          </DropdownMenuItem>
          <DropdownMenuItem @click="emit('addShape', 'line')">
            <MdiMinus class="mr-2 size-4" />
            {{ $t("components.label_template.editor.tools.line") }}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            class="size-10"
            :aria-label="$t('components.label_template.editor.tools.image')"
            @click="pickImage"
          >
            <MdiImagePlus class="size-5" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">{{ $t("components.label_template.editor.tools.image") }}</TooltipContent>
      </Tooltip>
      <input
        ref="fileInput"
        type="file"
        accept="image/png,image/jpeg,image/gif"
        class="hidden"
        @change="onImagePicked"
      />

      <div class="mx-1 h-6 w-px bg-border lg:mx-0 lg:my-1 lg:h-px lg:w-8" />

      <DropdownMenu>
        <Tooltip>
          <TooltipTrigger as-child>
            <DropdownMenuTrigger as-child>
              <Button
                variant="ghost"
                size="icon"
                class="size-10"
                :disabled="!props.hasSelection"
                :aria-label="$t('components.label_template.editor.tools.align')"
              >
                <MdiAlignHorizontalCenter class="size-5" />
              </Button>
            </DropdownMenuTrigger>
          </TooltipTrigger>
          <TooltipContent side="right">{{ $t("components.label_template.editor.tools.align") }}</TooltipContent>
        </Tooltip>
        <DropdownMenuContent side="right" align="start">
          <DropdownMenuLabel class="text-xs text-muted-foreground">
            {{ $t("components.label_template.editor.tools.align_to_label") }}
          </DropdownMenuLabel>
          <DropdownMenuItem v-for="a in alignTargets" :key="a.target" @click="emit('align', a.target)">
            <component :is="a.icon" class="mr-2 size-4" />
            {{ t(`components.label_template.editor.align.${a.key}`) }}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            class="size-10"
            :disabled="!props.canUndo"
            :aria-label="$t('components.label_template.editor.tools.undo')"
            @click="emit('undo')"
          >
            <MdiUndo class="size-5" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">{{ $t("components.label_template.editor.tools.undo") }} (⌘Z)</TooltipContent>
      </Tooltip>
      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            class="size-10"
            :disabled="!props.canRedo"
            :aria-label="$t('components.label_template.editor.tools.redo')"
            @click="emit('redo')"
          >
            <MdiRedo class="size-5" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">{{ $t("components.label_template.editor.tools.redo") }} (⇧⌘Z)</TooltipContent>
      </Tooltip>

      <div class="mx-1 h-6 w-px bg-border lg:mx-0 lg:my-1 lg:h-px lg:w-8" />

      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            class="size-10 text-destructive hover:text-destructive"
            :disabled="!props.hasSelection"
            :aria-label="$t('components.label_template.editor.tools.delete')"
            @click="emit('deleteSelected')"
          >
            <MdiDelete class="size-5" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="right">{{ $t("components.label_template.editor.tools.delete") }} (⌫)</TooltipContent>
      </Tooltip>
    </div>
  </TooltipProvider>
</template>
