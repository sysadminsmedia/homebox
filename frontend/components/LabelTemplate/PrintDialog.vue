<template>
  <BaseModal :dialog-id="DialogID.PrintLabelTemplate" hide-footer>
    <template #title>
      <div class="flex flex-col gap-0.5">
        <span>{{ $t("components.label_template.print_dialog.title") }}</span>
        <span class="text-sm font-normal text-muted-foreground">
          {{ subtitle }}
        </span>
      </div>
    </template>

    <div v-if="pending" class="flex items-center justify-center py-8">
      <div class="text-muted-foreground">{{ $t("global.loading") }}</div>
    </div>

    <div v-else-if="!templates || templates.length === 0" class="space-y-3 py-4 text-center">
      <p class="text-sm text-muted-foreground">{{ $t("components.label_template.print_dialog.no_templates") }}</p>
      <Button variant="outline" as-child>
        <NuxtLink to="/label-templates" @click="closeDialog(DialogID.PrintLabelTemplate)">
          {{ $t("components.label_template.print_dialog.create_template") }}
        </NuxtLink>
      </Button>
    </div>

    <div v-else class="space-y-4">
      <div class="grid gap-4 md:grid-cols-[260px_1fr]">
        <!-- Template picker -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <Label class="text-sm font-semibold">{{ $t("components.label_template.print_dialog.template") }}</Label>
            <NuxtLink
              to="/label-templates"
              class="text-xs text-muted-foreground hover:underline"
              @click="closeDialog(DialogID.PrintLabelTemplate)"
            >
              {{ $t("components.label_template.print_dialog.manage_templates") }}
            </NuxtLink>
          </div>
          <div role="radiogroup" class="max-h-72 space-y-2 overflow-y-auto pr-1">
            <button
              v-for="tpl in templates"
              :key="tpl.id"
              type="button"
              role="radio"
              :aria-checked="selectedTemplateId === tpl.id"
              class="flex w-full items-center gap-3 rounded-lg border p-2 text-left transition-colors hover:bg-muted/60"
              :class="selectedTemplateId === tpl.id ? 'border-primary bg-primary/5 ring-1 ring-primary' : ''"
              @click="selectedTemplateId = tpl.id"
            >
              <img
                :src="api.labelTemplates.getPreviewUrl(tpl.id)"
                alt=""
                class="h-10 w-20 shrink-0 rounded-sm bg-white object-cover object-left shadow"
                loading="lazy"
              />
              <span class="flex min-w-0 flex-col">
                <span class="truncate text-sm font-medium">{{ tpl.name }}</span>
                <span class="truncate text-xs text-muted-foreground">
                  {{ presetTitle(tpl.preset) || formatLabelSize(tpl.width, tpl.height) }}
                  <template v-if="tpl.id === preferences.defaultTemplateId">
                    · {{ $t("components.label_template.default").toLowerCase() }}</template
                  >
                </span>
              </span>
            </button>
          </div>
        </div>

        <!-- Preview + quantities -->
        <div class="min-w-0 space-y-3">
          <div class="flex flex-col items-center gap-2 rounded-lg bg-muted/60 p-3">
            <div v-if="isLoadingPreview" class="flex h-36 items-center justify-center text-sm text-muted-foreground">
              {{ $t("components.label_template.print_dialog.loading_preview") }}
            </div>
            <img
              v-else-if="previewUrl"
              :src="previewUrl"
              alt=""
              class="max-h-48 max-w-full rounded-sm bg-white shadow"
            />
            <div v-else class="flex h-36 items-center justify-center text-sm text-muted-foreground">
              {{ $t("components.label_template.print_dialog.select_placeholder") }}
            </div>
            <span v-if="selectedTemplate" class="text-xs text-muted-foreground">
              {{
                $t("components.label_template.print_dialog.preview_of", {
                  size: formatLabelSize(selectedTemplate.width, selectedTemplate.height),
                })
              }}
              <template v-if="isSheet"> · {{ $t("components.label_template.print_dialog.sheet_pdf_hint") }}</template>
            </span>
          </div>

          <table class="w-full text-sm">
            <caption class="pb-1 text-left text-sm font-semibold">
              {{
                mode === "location"
                  ? $t("components.label_template.print_dialog.copies_per_location")
                  : $t("components.label_template.print_dialog.copies_per_item")
              }}
            </caption>
            <tbody>
              <tr v-for="id in currentIds" :key="id" class="border-t">
                <td class="truncate py-1.5 pr-2">{{ nameFor(id) }}</td>
                <td class="py-1.5 text-right">
                  <Input
                    :model-value="getQuantity(id)"
                    type="number"
                    min="1"
                    max="100"
                    class="h-8 w-16 text-center"
                    :aria-label="$t('components.label_template.print_dialog.copies')"
                    @update:model-value="setQuantity(id, Number($event))"
                  />
                </td>
              </tr>
              <tr class="border-t">
                <td class="py-1.5 text-muted-foreground">{{ $t("components.label_template.print_dialog.total") }}</td>
                <td class="py-1.5 text-right font-semibold">
                  {{ totalLabels }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Destination -->
      <div class="space-y-2 border-t pt-3">
        <div class="flex items-center justify-between">
          <Label class="text-sm font-semibold">{{ $t("components.label_template.print_dialog.send_to") }}</Label>
          <NuxtLink
            to="/collection/printers"
            class="text-xs text-muted-foreground hover:underline"
            @click="closeDialog(DialogID.PrintLabelTemplate)"
          >
            {{ $t("components.label_template.print_dialog.manage_printers") }}
          </NuxtLink>
        </div>
        <div role="radiogroup" class="flex flex-wrap gap-2">
          <button
            v-for="dest in destinations"
            :key="dest.id"
            type="button"
            role="radio"
            :aria-checked="selectedDestination === dest.id"
            :disabled="dest.disabled"
            :title="dest.disabledReason"
            class="flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm disabled:cursor-not-allowed disabled:opacity-50"
            :class="
              selectedDestination === dest.id ? 'border-primary bg-primary/5 font-medium ring-1 ring-primary' : ''
            "
            @click="selectedDestination = dest.id"
          >
            <component :is="dest.icon" class="size-4 text-muted-foreground" />
            <span>{{ dest.label }}</span>
            <span
              v-if="dest.status"
              class="size-2 rounded-full"
              :class="{
                'bg-green-500': dest.status === 'online',
                'bg-red-500': dest.status === 'offline',
                'bg-gray-400': dest.status === 'unknown',
              }"
              :title="statusLabel(dest.status)"
            />
          </button>
        </div>
        <p v-if="selectedDestinationInfo?.disabledReason" class="text-xs text-muted-foreground">
          {{ selectedDestinationInfo.disabledReason }}
        </p>
      </div>

      <!-- Actions -->
      <div class="flex flex-wrap items-center justify-between gap-2 border-t pt-3">
        <Button variant="outline" :disabled="!selectedTemplateId || isBusy" @click="handleDownload">
          <MdiDownload class="mr-2 size-4" />
          {{
            isSheet || totalLabels > 1
              ? $t("components.label_template.print_dialog.download_pdf")
              : $t("components.label_template.print_dialog.download_png")
          }}
        </Button>
        <div class="flex gap-2">
          <Button type="button" variant="outline" @click="closeDialog(DialogID.PrintLabelTemplate)">
            {{ $t("global.cancel") }}
          </Button>
          <Button :disabled="!selectedTemplateId || isBusy" @click="handlePrint">
            <MdiLoading v-if="isBusy" class="mr-2 size-4 animate-spin" />
            <MdiPrinter v-else class="mr-2 size-4" />
            {{ printLabel }}
          </Button>
        </div>
      </div>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiDownload from "~icons/mdi/download";
  import MdiPrinter from "~icons/mdi/printer";
  import MdiLoading from "~icons/mdi/loading";
  import MdiDesktopClassic from "~icons/mdi/desktop-classic";
  import MdiUsb from "~icons/mdi/usb";
  import MdiBluetooth from "~icons/mdi/bluetooth";
  import { Button } from "@/components/ui/button";
  import { Label } from "@/components/ui/label";
  import { Input } from "@/components/ui/input";
  import { useDialog, DialogID } from "@/components/ui/dialog-provider/utils";
  import BaseModal from "@/components/App/CreateModal.vue";
  import type { LocalPrinter } from "~~/lib/printing/types";

  type PrintMode = "item" | "location";
  const BROWSER = "_browser";

  const { t } = useI18n();
  const api = useUserApi();
  const preferences = useViewPreferences();
  const { closeDialog, registerOpenDialogCallback } = useDialog();
  const { templates, pending } = useLabelTemplates();
  const { find: findPreset, title: presetTitle } = useLabelPresetLookup();
  const { printers } = usePrinters();
  const { printLabels, printLocationLabels } = useDirectPrint();
  const { localPrinters } = useLocalPrinters();
  const { print: printToLocal, isPrinting: isLocalPrinting } = useLocalPrint();

  const mode = ref<PrintMode>("item");
  const itemIds = ref<string[]>([]);
  const locationIds = ref<string[]>([]);
  const names = ref<Record<string, string>>({});
  const selectedTemplateId = ref<string>("");
  const selectedDestination = ref<string>(BROWSER);
  const isRendering = ref(false);
  const isBusy = computed(() => isRendering.value || isLocalPrinting.value);

  const currentIds = computed(() => (mode.value === "item" ? itemIds.value : locationIds.value));

  const quantities = ref<Record<string, number>>({});
  const getQuantity = (id: string) => quantities.value[id] ?? 1;
  const setQuantity = (id: string, qty: number) => {
    quantities.value[id] = Math.max(1, Math.min(100, Number.isFinite(qty) ? qty : 1));
  };
  const totalLabels = computed(() => currentIds.value.reduce((sum, id) => sum + getQuantity(id), 0));

  /** ids repeated per quantity, so a rendered PDF carries the right number of copies */
  const expandedIds = computed(() => currentIds.value.flatMap(id => Array(getQuantity(id)).fill(id) as string[]));

  const selectedTemplate = computed(() => templates.value?.find(tpl => tpl.id === selectedTemplateId.value) ?? null);
  const isSheet = computed(() => findPreset(selectedTemplate.value?.preset)?.kind === "sheet");

  const subtitle = computed(() => {
    const n = currentIds.value.length;
    const key = mode.value === "location" ? "locations" : "items";
    return t(`components.label_template.print_dialog.subtitle_${key}_${n === 1 ? "one" : "many"}`, { count: n });
  });
  const printLabel = computed(() =>
    t(`components.label_template.print_dialog.print_${totalLabels.value === 1 ? "one" : "many"}`, {
      count: totalLabels.value,
    })
  );

  function nameFor(id: string) {
    return names.value[id] ?? id;
  }

  // Resolve display names for the entities being printed
  async function loadNames(ids: string[]) {
    const missing = ids.filter(id => !names.value[id]);
    await Promise.all(
      missing.map(async id => {
        const { data } = await api.items.get(id);
        if (data?.name) names.value[id] = data.name;
      })
    );
  }

  // --- destinations ---------------------------------------------------------
  interface Destination {
    id: string;
    label: string;
    icon: Component;
    status?: string;
    disabled?: boolean;
    disabledReason?: string;
  }

  const destinations = computed<Destination[]>(() => {
    const list: Destination[] = [];
    for (const p of printers.value ?? []) {
      const isRoll = p.printerType === "brother_raster";
      const disabled = isSheet.value && isRoll;
      list.push({
        id: p.id,
        label: p.isDefault ? `${p.name} (${t("components.label_template.default").toLowerCase()})` : p.name,
        icon: MdiPrinter,
        status: p.status || "unknown",
        disabled,
        disabledReason: disabled ? t("components.label_template.print_dialog.roll_not_for_sheets") : undefined,
      });
    }
    for (const p of localPrinters.value) {
      const disabled = isSheet.value;
      list.push({
        id: p.id,
        label: p.name,
        icon: p.connectionType === "usb" ? MdiUsb : MdiBluetooth,
        disabled,
        disabledReason: disabled ? t("components.label_template.print_dialog.roll_not_for_sheets") : undefined,
      });
    }
    list.push({
      id: BROWSER,
      label: t("components.label_template.print_dialog.this_browser"),
      icon: MdiDesktopClassic,
    });
    return list;
  });

  const selectedDestinationInfo = computed(() => destinations.value.find(d => d.id === selectedDestination.value));
  const selectedServerPrinter = computed(() => printers.value?.find(p => p.id === selectedDestination.value) ?? null);
  const selectedLocalPrinter = computed(
    () => localPrinters.value.find(p => p.id === selectedDestination.value) as LocalPrinter | undefined
  );

  // Fall back to the browser when the chosen printer cannot take this template
  watch([isSheet, destinations], () => {
    if (selectedDestinationInfo.value?.disabled) selectedDestination.value = BROWSER;
  });

  function statusLabel(status: string) {
    return t(`components.label_template.print_dialog.status_${status}`);
  }

  // --- preview ----------------------------------------------------------------
  const previewUrl = ref<string | null>(null);
  const isLoadingPreview = ref(false);

  async function renderBlob(format: "png" | "pdf", ids: string[]): Promise<Blob> {
    const body = { format, pageSize: "Letter", showCutGuides: false, canvasData: "", skipLabels: 0 };
    return mode.value === "location"
      ? api.labelTemplates.renderLocations(selectedTemplateId.value, { ...body, locationIds: ids })
      : api.labelTemplates.render(selectedTemplateId.value, { ...body, itemIds: ids });
  }

  async function updatePreview() {
    if (previewUrl.value) {
      URL.revokeObjectURL(previewUrl.value);
      previewUrl.value = null;
    }
    const firstId = currentIds.value[0];
    if (!selectedTemplateId.value || !firstId) return;

    isLoadingPreview.value = true;
    try {
      const blob = await renderBlob("png", [firstId]);
      previewUrl.value = URL.createObjectURL(blob);
    } catch {
      previewUrl.value = api.labelTemplates.getPreviewUrl(selectedTemplateId.value);
    } finally {
      isLoadingPreview.value = false;
    }
  }

  watch([selectedTemplateId, currentIds], updatePreview);

  // --- actions ----------------------------------------------------------------
  function triggerDownload(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  async function handleDownload() {
    if (!selectedTemplateId.value || currentIds.value.length === 0) return;
    isRendering.value = true;
    try {
      const pdf = isSheet.value || totalLabels.value > 1;
      const blob = await renderBlob(pdf ? "pdf" : "png", pdf ? expandedIds.value : currentIds.value);
      triggerDownload(blob, `labels-${Date.now()}.${pdf ? "pdf" : "png"}`);
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
    if (selectedServerPrinter.value) return handleServerPrint();
    if (selectedLocalPrinter.value) return handleLocalPrint(selectedLocalPrinter.value);
    return handleBrowserPrint();
  }

  async function handleBrowserPrint() {
    isRendering.value = true;
    try {
      const pdf = isSheet.value || totalLabels.value > 1;
      const blob = await renderBlob(pdf ? "pdf" : "png", pdf ? expandedIds.value : currentIds.value);
      const url = URL.createObjectURL(blob);
      const printWindow = window.open(url, "_blank");
      if (printWindow) {
        printWindow.onload = () => printWindow.print();
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
      const withQty = currentIds.value.map(id => ({ id, quantity: getQuantity(id) }));
      const result =
        mode.value === "location"
          ? await printLocationLabels(selectedTemplateId.value, withQty, selectedDestination.value)
          : await printLabels(selectedTemplateId.value, withQty, selectedDestination.value);

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
      for (const id of currentIds.value) {
        const blob = await renderBlob("png", [id]);
        const imageData = new Uint8Array(await blob.arrayBuffer());
        const result = await printToLocal(printer, imageData, {
          copies: getQuantity(id),
          labelWidth: selectedTemplate.value.width,
          labelHeight: selectedTemplate.value.height,
        });
        if (!result.success) {
          toast.error(result.message || t("components.label_template.toast.print_failed"));
          return;
        }
      }
      toast.success(t("components.label_template.toast.print_sent", { printer: printer.name }));
      closeDialog(DialogID.PrintLabelTemplate);
    } catch (error) {
      toast.error((error as Error).message || t("components.label_template.toast.print_failed"));
    } finally {
      isRendering.value = false;
    }
  }

  onMounted(() => {
    registerOpenDialogCallback(DialogID.PrintLabelTemplate, params => {
      if (params.locationIds && params.locationIds.length > 0) {
        mode.value = "location";
        locationIds.value = params.locationIds;
        itemIds.value = [];
      } else {
        mode.value = "item";
        itemIds.value = params.itemIds || [];
        locationIds.value = [];
      }

      quantities.value = {};
      if (previewUrl.value) {
        URL.revokeObjectURL(previewUrl.value);
        previewUrl.value = null;
      }

      // Pre-select the user's default template, else the first one
      const available = templates.value ?? [];
      const preferred = available.find(tpl => tpl.id === preferences.value.defaultTemplateId);
      selectedTemplateId.value = preferred?.id ?? available[0]?.id ?? "";

      const defaultPrinter = printers.value?.find(p => p.isDefault);
      selectedDestination.value = defaultPrinter?.id || BROWSER;

      void loadNames(currentIds.value);
    });
  });

  onBeforeUnmount(() => {
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value);
  });
</script>
