<template>
  <BaseModal :dialog-id="DialogID.PrintLabelTemplate" :title="$t('components.label_template.print_dialog.title')">
    <div class="space-y-4">
      <div v-if="pending" class="flex items-center justify-center py-8">
        <div class="text-muted-foreground">{{ $t("global.loading") }}</div>
      </div>

      <template v-else>
        <div class="space-y-2">
          <Label for="template">{{ $t("components.label_template.print_dialog.select_template") }}</Label>
          <Select v-model="selectedTemplateId">
            <SelectTrigger>
              <SelectValue :placeholder="$t('components.label_template.print_dialog.select_placeholder')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="tpl in templates" :key="tpl.id" :value="tpl.id">
                {{ tpl.name }} ({{ tpl.width }}mm x {{ tpl.height }}mm)
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div v-if="selectedTemplate" class="rounded border p-3 text-sm">
          <div class="font-medium">{{ selectedTemplate.name }}</div>
          <div class="text-muted-foreground">
            {{ selectedTemplate.width }}mm x {{ selectedTemplate.height }}mm
            <span v-if="selectedTemplate.preset" class="ml-2">({{ selectedTemplate.preset }})</span>
          </div>
        </div>

        <!-- Preview with real data -->
        <div v-if="selectedTemplateId" class="rounded border bg-gray-50 p-4">
          <div v-if="isLoadingPreview" class="flex items-center justify-center py-8">
            <div class="text-sm text-muted-foreground">
              {{ $t("components.label_template.print_dialog.loading_preview") }}
            </div>
          </div>
          <div v-else-if="previewUrl" class="flex flex-col items-center gap-2">
            <img :src="previewUrl" alt="Label preview" class="max-h-48" />
            <div v-if="currentIds.length > 1" class="text-xs text-muted-foreground">
              <template v-if="mode === 'location'">
                {{ $t("components.label_template.print_dialog.preview_of_locations", { count: currentIds.length }) }}
              </template>
              <template v-else>
                {{ $t("components.label_template.print_dialog.preview_of_items", { count: currentIds.length }) }}
              </template>
            </div>
          </div>
        </div>

        <!-- Per-Item/Location Quantities (shown when multiple) -->
        <div v-if="currentIds.length > 1" class="space-y-2 rounded border p-3">
          <div class="flex items-center justify-between">
            <Label>
              <template v-if="mode === 'location'">
                {{ $t("components.label_template.print_dialog.location_quantities") }}
              </template>
              <template v-else>
                {{ $t("components.label_template.print_dialog.item_quantities") }}
              </template>
            </Label>
            <span class="text-xs text-muted-foreground">
              {{ $t("components.label_template.print_dialog.total_labels", { count: totalLabels }) }}
            </span>
          </div>
          <div class="max-h-48 space-y-2 overflow-y-auto">
            <div
              v-for="id in currentIds"
              :key="id"
              class="flex items-center justify-between gap-2 rounded bg-gray-50 px-2 py-1"
            >
              <span class="truncate text-sm">{{ id }}</span>
              <Input
                :model-value="getItemQuantity(id)"
                type="number"
                min="1"
                max="100"
                class="h-7 w-16 text-center"
                @update:model-value="setItemQuantity(id, Number($event))"
              />
            </div>
          </div>
        </div>

        <!-- Printer Selection Section -->
        <div class="space-y-2 rounded border p-3">
          <div class="flex items-center justify-between">
            <Label>{{ $t("components.label_template.print_dialog.printer") }}</Label>
            <RouterLink to="/printers" class="text-xs text-muted-foreground hover:underline">
              {{ $t("components.label_template.print_dialog.manage_printers") }}
            </RouterLink>
          </div>

          <Select v-model="selectedPrinterId">
            <SelectTrigger>
              <SelectValue :placeholder="$t('components.label_template.print_dialog.select_printer')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="_browser">
                <span class="flex items-center gap-2">
                  <MdiDesktopClassic class="size-4" />
                  {{ $t("components.label_template.print_dialog.browser_print") }}
                </span>
              </SelectItem>
              <template v-if="printers && printers.length > 0">
                <SelectItem v-for="printer in printers" :key="printer.id" :value="printer.id">
                  <span class="flex items-center gap-2">
                    <MdiPrinter class="size-4" />
                    {{ printer.name }}
                    <span v-if="printer.isDefault" class="text-xs text-muted-foreground">(Default)</span>
                  </span>
                </SelectItem>
              </template>
              <template v-if="localPrinters.length > 0">
                <SelectItem v-for="printer in localPrinters" :key="printer.id" :value="printer.id">
                  <span class="flex items-center gap-2">
                    <MdiUsb v-if="printer.connectionType === 'usb'" class="size-4" />
                    <MdiBluetooth v-else class="size-4" />
                    {{ printer.name }}
                    <span class="text-xs text-muted-foreground">(Local)</span>
                  </span>
                </SelectItem>
              </template>
            </SelectContent>
          </Select>

          <!-- Print settings for direct print (only show copies for single item/location) -->
          <div
            v-if="selectedPrinterId && selectedPrinterId !== '_browser' && currentIds.length === 1"
            class="space-y-3 border-t pt-3"
          >
            <!-- Copies -->
            <div class="flex items-center gap-4">
              <Label for="copies" class="whitespace-nowrap">{{
                $t("components.label_template.print_dialog.copies")
              }}</Label>
              <Input id="copies" v-model.number="copies" type="number" min="1" max="100" class="w-20" />
            </div>
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <Button type="button" variant="outline" @click="closeDialog(DialogID.PrintLabelTemplate)">
            {{ $t("global.cancel") }}
          </Button>
          <Button :disabled="!selectedTemplateId || isRendering" variant="outline" @click="handleDownload">
            <MdiDownload class="mr-2 size-4" />
            {{ $t("components.label_template.print_dialog.download") }}
          </Button>
          <Button :disabled="!selectedTemplateId || isRendering || isPrinting" @click="handlePrint">
            <MdiPrinter class="mr-2 size-4" />
            <template v-if="selectedPrinterId && selectedPrinterId !== '_browser'">
              {{ $t("components.label_template.print_dialog.print_direct") }}
            </template>
            <template v-else>
              {{ $t("components.label_template.print_dialog.print") }}
            </template>
          </Button>
        </div>
      </template>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiDownload from "~icons/mdi/download";
  import MdiPrinter from "~icons/mdi/printer";
  import MdiDesktopClassic from "~icons/mdi/desktop-classic";
  import MdiUsb from "~icons/mdi/usb";
  import MdiBluetooth from "~icons/mdi/bluetooth";
  import { Button } from "@/components/ui/button";
  import { Label } from "@/components/ui/label";
  import { Input } from "@/components/ui/input";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { useDialog, DialogID } from "@/components/ui/dialog-provider/utils";
  import BaseModal from "@/components/App/CreateModal.vue";
  import type { LocalPrinter } from "~~/lib/printing/types";

  type PrintMode = "item" | "location";

  const { t } = useI18n();
  const { closeDialog, registerOpenDialogCallback } = useDialog();
  const { templates, pending } = useLabelTemplates();
  const { render, renderLocations } = useLabelTemplateActions();
  const { printers } = usePrinters();
  const { printLabels, printLocationLabels } = useDirectPrint();
  const { localPrinters } = useLocalPrinters();
  const { print: printToLocal, isPrinting: isLocalPrinting } = useLocalPrint();

  const mode = ref<PrintMode>("item");
  const itemIds = ref<string[]>([]);
  const locationIds = ref<string[]>([]);
  const selectedTemplateId = ref<string>("");
  const selectedPrinterId = ref<string>("_browser");
  const copies = ref<number>(1);
  const isRendering = ref(false);
  const isPrinting = computed(() => isRendering.value || isLocalPrinting.value);

  // Get the current IDs based on mode
  const currentIds = computed(() => (mode.value === "item" ? itemIds.value : locationIds.value));

  // Per-item quantities (itemId -> quantity)
  const itemQuantities = ref<Record<string, number>>({});

  // Get quantity for an item (defaults to 1)
  function getItemQuantity(id: string): number {
    return itemQuantities.value[id] ?? 1;
  }

  // Set quantity for an item
  function setItemQuantity(id: string, qty: number) {
    itemQuantities.value[id] = Math.max(1, Math.min(100, qty));
  }

  // Total labels to print
  const totalLabels = computed(() => {
    return itemIds.value.reduce((sum, id) => sum + getItemQuantity(id), 0);
  });

  const selectedTemplate = computed(() => {
    if (!selectedTemplateId.value || !templates.value) return null;
    return templates.value.find(t => t.id === selectedTemplateId.value);
  });

  // Real data preview - render with actual item data
  const previewUrl = ref<string | null>(null);
  const isLoadingPreview = ref(false);

  // Render preview with real data when template or IDs change
  async function updatePreview() {
    // Clean up previous preview URL
    if (previewUrl.value) {
      URL.revokeObjectURL(previewUrl.value);
      previewUrl.value = null;
    }

    const templateId = selectedTemplateId.value;
    const firstId = currentIds.value[0];

    if (!templateId || !firstId) {
      return;
    }

    isLoadingPreview.value = true;
    try {
      // Render with the first item/location to show preview with real data
      let blob: Blob;
      if (mode.value === "location") {
        blob = await renderLocations(templateId, [firstId]);
      } else {
        blob = await render(templateId, [firstId]);
      }
      previewUrl.value = URL.createObjectURL(blob);
    } catch {
      // Fall back to static preview if render fails
      const api = useUserApi();
      previewUrl.value = api.labelTemplates.getPreviewUrl(templateId);
    } finally {
      isLoadingPreview.value = false;
    }
  }

  // Watch for changes and update preview with debounce
  watch(
    [selectedTemplateId, currentIds],
    () => {
      updatePreview();
    },
    { immediate: false }
  );

  // Find selected local printer
  const selectedLocalPrinter = computed(() => {
    if (!selectedPrinterId.value || selectedPrinterId.value === "_browser") return null;
    return localPrinters.value.find(p => p.id === selectedPrinterId.value);
  });

  // Check if selected printer is a server printer
  const isServerPrinter = computed(() => {
    if (!selectedPrinterId.value || selectedPrinterId.value === "_browser") return false;
    if (selectedLocalPrinter.value) return false;
    return printers.value?.some(p => p.id === selectedPrinterId.value) ?? false;
  });

  async function handleDownload() {
    if (!selectedTemplateId.value || currentIds.value.length === 0) return;

    isRendering.value = true;
    try {
      let blob: Blob;
      if (mode.value === "location") {
        blob = await renderLocations(selectedTemplateId.value, locationIds.value);
      } else {
        blob = await render(selectedTemplateId.value, itemIds.value);
      }

      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `label-${Date.now()}.png`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      toast.success(t("components.label_template.toast.downloaded"));
      closeDialog(DialogID.PrintLabelTemplate);
    } catch {
      toast.error(t("components.label_template.toast.render_failed"));
    } finally {
      isRendering.value = false;
    }
  }

  async function handlePrint() {
    if (!selectedTemplateId.value || currentIds.value.length === 0) return;

    // Browser print
    if (selectedPrinterId.value === "_browser") {
      await handleBrowserPrint();
      return;
    }

    // Server printer (IPP/CUPS)
    if (isServerPrinter.value) {
      await handleServerPrint();
      return;
    }

    // Local printer (WebUSB/Bluetooth)
    if (selectedLocalPrinter.value) {
      await handleLocalPrint(selectedLocalPrinter.value);
      return;
    }

    // Fallback to browser print
    await handleBrowserPrint();
  }

  async function handleBrowserPrint() {
    isRendering.value = true;
    try {
      let blob: Blob;
      if (mode.value === "location") {
        blob = await renderLocations(selectedTemplateId.value, locationIds.value);
      } else {
        blob = await render(selectedTemplateId.value, itemIds.value);
      }

      const url = URL.createObjectURL(blob);
      const printWindow = window.open(url, "_blank");
      if (printWindow) {
        printWindow.onload = () => {
          printWindow.print();
        };
      }

      toast.success(t("components.label_template.toast.printed"));
      closeDialog(DialogID.PrintLabelTemplate);
    } catch {
      toast.error(t("components.label_template.toast.render_failed"));
    } finally {
      isRendering.value = false;
    }
  }

  async function handleServerPrint() {
    isRendering.value = true;
    try {
      let result;

      if (mode.value === "location") {
        // Location printing
        if (locationIds.value.length > 1) {
          const locations = locationIds.value.map(id => ({
            id,
            quantity: getItemQuantity(id),
          }));
          result = await printLocationLabels(selectedTemplateId.value, locations, selectedPrinterId.value);
        } else {
          result = await printLocationLabels(
            selectedTemplateId.value,
            locationIds.value,
            selectedPrinterId.value,
            copies.value
          );
        }
      } else {
        // Item printing
        if (itemIds.value.length > 1) {
          const items = itemIds.value.map(id => ({
            id,
            quantity: getItemQuantity(id),
          }));
          result = await printLabels(selectedTemplateId.value, items, selectedPrinterId.value);
        } else {
          result = await printLabels(selectedTemplateId.value, itemIds.value, selectedPrinterId.value, copies.value);
        }
      }

      if (result?.success) {
        toast.success(t("components.label_template.toast.print_sent", { printer: result.printerName }));
        closeDialog(DialogID.PrintLabelTemplate);
      } else {
        toast.error(result?.message || t("components.label_template.toast.print_failed"));
      }
    } catch {
      toast.error(t("components.label_template.toast.print_failed"));
    } finally {
      isRendering.value = false;
    }
  }

  async function handleLocalPrint(printer: LocalPrinter) {
    if (!selectedTemplate.value) return;

    isRendering.value = true;
    try {
      // Render the label as PNG
      let blob: Blob;
      if (mode.value === "location") {
        blob = await renderLocations(selectedTemplateId.value, locationIds.value);
      } else {
        blob = await render(selectedTemplateId.value, itemIds.value);
      }
      const arrayBuffer = await blob.arrayBuffer();
      const imageData = new Uint8Array(arrayBuffer);

      // Send to local printer using template dimensions
      const result = await printToLocal(printer, imageData, {
        copies: copies.value,
        labelWidth: selectedTemplate.value.width,
        labelHeight: selectedTemplate.value.height,
      });

      if (result.success) {
        toast.success(t("components.label_template.toast.print_sent", { printer: printer.name }));
        closeDialog(DialogID.PrintLabelTemplate);
      } else {
        toast.error(result.message || t("components.label_template.toast.print_failed"));
      }
    } catch (error) {
      toast.error((error as Error).message || t("components.label_template.toast.print_failed"));
    } finally {
      isRendering.value = false;
    }
  }

  onMounted(() => {
    registerOpenDialogCallback(DialogID.PrintLabelTemplate, params => {
      // Determine mode based on which IDs are provided
      if (params.locationIds && params.locationIds.length > 0) {
        mode.value = "location";
        locationIds.value = params.locationIds;
        itemIds.value = [];
      } else {
        mode.value = "item";
        itemIds.value = params.itemIds || [];
        locationIds.value = [];
      }

      selectedTemplateId.value = "";
      copies.value = 1;

      // Reset per-item/location quantities (all default to 1)
      itemQuantities.value = {};

      // Clean up previous preview
      if (previewUrl.value) {
        URL.revokeObjectURL(previewUrl.value);
        previewUrl.value = null;
      }

      // Default to default printer or browser print
      const defaultPrinter = printers.value?.find(p => p.isDefault);
      selectedPrinterId.value = defaultPrinter?.id || "_browser";
    });
  });

  // Clean up preview URL on unmount
  onBeforeUnmount(() => {
    if (previewUrl.value) {
      URL.revokeObjectURL(previewUrl.value);
    }
  });
</script>
