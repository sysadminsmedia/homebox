<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import type { ItemSummary, LabelSummary, LocationOutCount } from "~~/lib/api/types/data-contracts";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiLoading from "~icons/mdi/loading";
  import MdiMagnify from "~icons/mdi/magnify";
  import MdiPrinter from "~icons/mdi/printer";
  import MdiDownload from "~icons/mdi/download";
  import MdiCheckboxMarked from "~icons/mdi/checkbox-marked";
  import MdiCheckboxBlankOutline from "~icons/mdi/checkbox-blank-outline";
  import MdiArrowUp from "~icons/mdi/arrow-up";
  import MdiArrowDown from "~icons/mdi/arrow-down";
  import MdiClose from "~icons/mdi/close";
  import MdiDrag from "~icons/mdi/drag";
  import MdiPrinterPos from "~icons/mdi/printer-pos";
  import { Input } from "~/components/ui/input";
  import { Button } from "@/components/ui/button";
  import { Label } from "@/components/ui/label";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Badge } from "@/components/ui/badge";
  import BaseContainer from "@/components/Base/Container.vue";
  import SearchFilter from "~/components/Search/Filter.vue";

  const { t } = useI18n();

  definePageMeta({
    middleware: ["auth"],
  });

  useHead({
    title: computed(() => `HomeBox | ${t("pages.label_templates.batch.title")}`),
  });

  const api = useUserApi();
  const loading = ref(false);
  const isRendering = ref(false);
  const items = ref<ItemSummary[]>([]);
  // Ordered array of selected item IDs (preserves print order)
  const selectedItemIds = ref<string[]>([]);

  // Template selection
  const { templates, pending: templatesPending } = useLabelTemplates();
  const selectedTemplateId = ref<string>("");

  const selectedTemplate = computed(() => {
    if (!selectedTemplateId.value || !templates.value) return null;
    return templates.value.find(t => t.id === selectedTemplateId.value);
  });

  // Printer selection for direct printing
  const { printers } = usePrinters();
  const { printLabels } = useDirectPrint();
  const selectedPrinterId = ref<string>("");
  const isPrinting = ref(false);

  // Find default printer on load
  watch(
    printers,
    newPrinters => {
      if (newPrinters && !selectedPrinterId.value) {
        const defaultPrinter = newPrinters.find(p => p.isDefault);
        if (defaultPrinter) {
          selectedPrinterId.value = defaultPrinter.id;
        }
      }
    },
    { immediate: true }
  );

  const hasPrinters = computed(() => printers.value && printers.value.length > 0);

  // Search state
  const query = ref("");
  const locationsStore = useLocationStore();
  const labelStore = useLabelStore();
  const locationFlatTree = useFlatLocations();
  const labels = computed(() => labelStore.labels);
  const selectedLocations = ref<LocationOutCount[]>([]);
  const selectedLabels = ref<LabelSummary[]>([]);

  const locIDs = computed(() => selectedLocations.value.map(l => l.id));
  const labIDs = computed(() => selectedLabels.value.map(l => l.id));

  onMounted(async () => {
    await Promise.all([locationsStore.ensureLocationsFetched(), labelStore.ensureAllLabelsFetched()]);
  });

  async function search() {
    loading.value = true;

    const { data, error } = await api.items.getAll({
      q: query.value || "",
      locations: locIDs.value,
      labels: labIDs.value,
      pageSize: 100,
    });

    if (error) {
      toast.error(t("items.toast.failed_search_items"));
      loading.value = false;
      return;
    }

    items.value = data.items || [];
    loading.value = false;
  }

  watchDebounced([query, selectedLabels, selectedLocations], search, { debounce: 250, maxWait: 1000 });

  // Check if an item is selected
  function isSelected(itemId: string): boolean {
    return selectedItemIds.value.includes(itemId);
  }

  function toggleItem(itemId: string) {
    const index = selectedItemIds.value.indexOf(itemId);
    if (index >= 0) {
      // Remove from selection
      selectedItemIds.value = selectedItemIds.value.filter(id => id !== itemId);
    } else {
      // Add to end of selection (preserves order)
      selectedItemIds.value = [...selectedItemIds.value, itemId];
    }
  }

  function selectAll() {
    // Add items not already selected (preserves existing order)
    const currentIds = new Set(selectedItemIds.value);
    const newIds = items.value.filter(item => !currentIds.has(item.id)).map(item => item.id);
    selectedItemIds.value = [...selectedItemIds.value, ...newIds];
  }

  function deselectAll() {
    selectedItemIds.value = [];
  }

  function removeFromQueue(itemId: string) {
    selectedItemIds.value = selectedItemIds.value.filter(id => id !== itemId);
  }

  function moveUp(index: number) {
    if (index <= 0) return;
    const newOrder = [...selectedItemIds.value];
    const temp = newOrder[index]!;
    newOrder[index] = newOrder[index - 1]!;
    newOrder[index - 1] = temp;
    selectedItemIds.value = newOrder;
  }

  function moveDown(index: number) {
    if (index >= selectedItemIds.value.length - 1) return;
    const newOrder = [...selectedItemIds.value];
    const temp = newOrder[index]!;
    newOrder[index] = newOrder[index + 1]!;
    newOrder[index + 1] = temp;
    selectedItemIds.value = newOrder;
  }

  const selectedCount = computed(() => selectedItemIds.value.length);
  const allSelected = computed(() => items.value.length > 0 && items.value.every(i => isSelected(i.id)));

  // Get full item data for selected items in order
  const selectedItems = computed(() => {
    const itemMap = new Map(items.value.map(item => [item.id, item]));
    return selectedItemIds.value.map(id => itemMap.get(id)).filter((item): item is ItemSummary => item !== undefined);
  });

  // PDF options
  const outputFormat = ref<"png" | "pdf">("png");
  const pageSize = ref<"Letter" | "A4">("Letter");
  const showCutGuides = ref(false);

  async function handleDownload() {
    if (!selectedTemplateId.value || selectedCount.value === 0) return;

    isRendering.value = true;
    try {
      const blob = await api.labelTemplates.render(selectedTemplateId.value, {
        itemIds: selectedItemIds.value,
        format: outputFormat.value,
        pageSize: pageSize.value,
        showCutGuides: showCutGuides.value,
        canvasData: "", // Use saved template data
      });

      const ext = outputFormat.value === "pdf" ? "pdf" : "png";
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `labels-batch-${Date.now()}.${ext}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      toast.success(t("components.label_template.toast.downloaded"));
    } catch {
      toast.error(t("components.label_template.toast.render_failed"));
    } finally {
      isRendering.value = false;
    }
  }

  async function handlePrint() {
    if (!selectedTemplateId.value || selectedCount.value === 0) return;

    isRendering.value = true;
    try {
      // For printing, always use PDF format for multiple items
      const format = selectedCount.value > 1 ? "pdf" : outputFormat.value;
      const blob = await api.labelTemplates.render(selectedTemplateId.value, {
        itemIds: selectedItemIds.value,
        format,
        pageSize: pageSize.value,
        showCutGuides: showCutGuides.value,
        canvasData: "", // Use saved template data
      });

      const url = URL.createObjectURL(blob);
      const printWindow = window.open(url, "_blank");
      if (printWindow) {
        printWindow.onload = () => {
          printWindow.print();
        };
      }

      toast.success(t("components.label_template.toast.printed"));
    } catch {
      toast.error(t("components.label_template.toast.render_failed"));
    } finally {
      isRendering.value = false;
    }
  }

  async function handleDirectPrint() {
    if (!selectedTemplateId.value || !selectedPrinterId.value || selectedCount.value === 0) return;

    isPrinting.value = true;
    try {
      await printLabels(selectedTemplateId.value, selectedItemIds.value, selectedPrinterId.value);
      toast.success(t("pages.label_templates.batch.direct_print_success", { count: selectedCount.value }));
    } catch {
      toast.error(t("pages.label_templates.batch.direct_print_failed"));
    } finally {
      isPrinting.value = false;
    }
  }
</script>

<template>
  <BaseContainer>
    <div class="mb-6">
      <h1 class="text-2xl font-semibold">{{ $t("pages.label_templates.batch.title") }}</h1>
      <p class="text-muted-foreground">{{ $t("pages.label_templates.batch.description") }}</p>
    </div>

    <!-- Template Selection -->
    <Card class="mb-6">
      <CardHeader>
        <CardTitle>{{ $t("pages.label_templates.batch.select_template") }}</CardTitle>
        <CardDescription>{{ $t("pages.label_templates.batch.select_template_description") }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="templatesPending" class="text-muted-foreground">{{ $t("global.loading") }}</div>
        <div v-else-if="!templates || templates.length === 0" class="text-muted-foreground">
          {{ $t("pages.label_templates.batch.no_templates") }}
          <NuxtLink to="/label-templates" class="text-primary underline">
            {{ $t("pages.label_templates.batch.create_template") }}
          </NuxtLink>
        </div>
        <div v-else class="flex flex-col gap-4 md:flex-row md:items-end">
          <div class="flex-1">
            <Label for="template">{{ $t("components.label_template.print_dialog.select_template") }}</Label>
            <Select v-model="selectedTemplateId">
              <SelectTrigger class="mt-1">
                <SelectValue :placeholder="$t('components.label_template.print_dialog.select_placeholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="tpl in templates" :key="tpl.id" :value="tpl.id">
                  {{ tpl.name }} ({{ tpl.width }}mm x {{ tpl.height }}mm)
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div v-if="selectedTemplate" class="text-sm text-muted-foreground">
            {{ selectedTemplate.width }}mm x {{ selectedTemplate.height }}mm
            <span v-if="selectedTemplate.preset" class="ml-2">({{ selectedTemplate.preset }})</span>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Item Search and Selection -->
    <Card class="mb-6">
      <CardHeader>
        <CardTitle>{{ $t("pages.label_templates.batch.select_items") }}</CardTitle>
        <CardDescription>{{ $t("pages.label_templates.batch.select_items_description") }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="flex flex-col gap-4">
          <!-- Search Controls -->
          <div class="flex flex-wrap items-end gap-4 md:flex-nowrap">
            <div class="w-full md:flex-1">
              <Input v-model="query" :placeholder="$t('global.search')" />
            </div>
            <Button @click="search">
              <MdiLoading v-if="loading" class="mr-2 animate-spin" />
              <MdiMagnify v-else class="mr-2" />
              {{ $t("global.search") }}
            </Button>
          </div>

          <!-- Filters -->
          <div class="flex flex-wrap gap-2">
            <SearchFilter v-model="selectedLocations" :label="$t('global.locations')" :options="locationFlatTree" />
            <SearchFilter v-model="selectedLabels" :label="$t('global.labels')" :options="labels" />
          </div>

          <!-- Selection Controls -->
          <div class="flex items-center justify-between border-b pb-2">
            <div class="flex items-center gap-4">
              <Button v-if="!allSelected" variant="outline" size="sm" @click="selectAll">
                <MdiCheckboxMarked class="mr-2 size-4" />
                {{ $t("pages.label_templates.batch.select_all") }}
              </Button>
              <Button v-else variant="outline" size="sm" @click="deselectAll">
                <MdiCheckboxBlankOutline class="mr-2 size-4" />
                {{ $t("pages.label_templates.batch.deselect_all") }}
              </Button>
            </div>
            <Badge variant="secondary">
              {{ $t("pages.label_templates.batch.selected_count", { count: selectedCount }) }}
            </Badge>
          </div>

          <!-- Item List -->
          <div v-if="loading" class="flex items-center justify-center py-8">
            <MdiLoading class="size-6 animate-spin text-muted-foreground" />
          </div>
          <div v-else-if="items.length === 0" class="py-8 text-center text-muted-foreground">
            {{ $t("pages.label_templates.batch.no_items") }}
          </div>
          <div v-else class="max-h-96 space-y-1 overflow-y-auto">
            <div
              v-for="item in items"
              :key="item.id"
              class="flex cursor-pointer items-center gap-3 rounded-md p-2 hover:bg-muted"
              @click="toggleItem(item.id)"
            >
              <Checkbox :checked="isSelected(item.id)" />
              <div class="flex-1">
                <div class="font-medium">{{ item.name }}</div>
                <div class="text-sm text-muted-foreground">
                  {{ item.location?.name || $t("pages.label_templates.batch.no_location") }}
                  <span v-if="item.assetId" class="ml-2">#{{ item.assetId }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Print Queue (Reorder) -->
    <Card v-if="selectedCount > 0" class="mb-6">
      <CardHeader>
        <CardTitle>{{ $t("pages.label_templates.batch.print_queue") }}</CardTitle>
        <CardDescription>{{ $t("pages.label_templates.batch.print_queue_description") }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="space-y-1">
          <div
            v-for="(item, index) in selectedItems"
            :key="item.id"
            class="flex items-center gap-2 rounded-md border bg-card p-2"
          >
            <MdiDrag class="size-4 text-muted-foreground" />
            <span class="w-6 text-center text-sm text-muted-foreground">{{ index + 1 }}</span>
            <div class="flex-1">
              <div class="font-medium">{{ item.name }}</div>
              <div class="text-xs text-muted-foreground">
                {{ item.location?.name || $t("pages.label_templates.batch.no_location") }}
                <span v-if="item.assetId" class="ml-1">#{{ item.assetId }}</span>
              </div>
            </div>
            <div class="flex items-center gap-1">
              <Button
                variant="ghost"
                size="icon"
                class="size-7"
                :disabled="index === 0"
                :title="$t('pages.label_templates.batch.move_up')"
                @click="moveUp(index)"
              >
                <MdiArrowUp class="size-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                class="size-7"
                :disabled="index === selectedItems.length - 1"
                :title="$t('pages.label_templates.batch.move_down')"
                @click="moveDown(index)"
              >
                <MdiArrowDown class="size-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                class="size-7 text-destructive hover:text-destructive"
                :title="$t('global.remove')"
                @click="removeFromQueue(item.id)"
              >
                <MdiClose class="size-4" />
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Output Options -->
    <Card class="mb-6">
      <CardHeader>
        <CardTitle>{{ $t("pages.label_templates.batch.output_options") }}</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="flex flex-col gap-4 md:flex-row">
          <div class="flex-1">
            <Label>{{ $t("pages.label_templates.batch.output_format") }}</Label>
            <Select v-model="outputFormat">
              <SelectTrigger class="mt-1">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="png">PNG ({{ $t("pages.label_templates.batch.single_image") }})</SelectItem>
                <SelectItem value="pdf">PDF ({{ $t("pages.label_templates.batch.multi_page") }})</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div v-if="outputFormat === 'pdf'" class="flex-1">
            <Label>{{ $t("pages.label_templates.batch.page_size") }}</Label>
            <Select v-model="pageSize">
              <SelectTrigger class="mt-1">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="Letter">Letter (8.5" x 11")</SelectItem>
                <SelectItem value="A4">A4 (210mm x 297mm)</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <div v-if="outputFormat === 'pdf'" class="flex items-center gap-2 pt-2">
          <Checkbox id="cutGuides" v-model:checked="showCutGuides" />
          <Label for="cutGuides" class="cursor-pointer text-sm font-normal">
            {{ $t("pages.label_templates.batch.show_cut_guides") }}
          </Label>
        </div>
      </CardContent>
    </Card>

    <!-- Direct Print to Label Maker -->
    <Card v-if="hasPrinters" class="mb-6">
      <CardHeader>
        <CardTitle>{{ $t("pages.label_templates.batch.direct_print") }}</CardTitle>
        <CardDescription>{{ $t("pages.label_templates.batch.direct_print_description") }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="flex flex-col gap-4 md:flex-row md:items-end">
          <div class="flex-1">
            <Label>{{ $t("pages.label_templates.batch.select_printer") }}</Label>
            <Select v-model="selectedPrinterId">
              <SelectTrigger class="mt-1">
                <SelectValue :placeholder="$t('pages.label_templates.batch.select_printer_placeholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="printer in printers" :key="printer.id" :value="printer.id">
                  {{ printer.name }}
                  <span v-if="printer.isDefault" class="ml-1 text-xs text-muted-foreground"
                    >({{ $t("global.default") }})</span
                  >
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button
            :disabled="!selectedTemplateId || !selectedPrinterId || selectedCount === 0 || isPrinting"
            @click="handleDirectPrint"
          >
            <MdiLoading v-if="isPrinting" class="mr-2 size-4 animate-spin" />
            <MdiPrinterPos v-else class="mr-2 size-4" />
            {{ $t("pages.label_templates.batch.send_to_printer", { count: selectedCount }) }}
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Action Buttons -->
    <div class="flex justify-end gap-2">
      <Button
        variant="outline"
        :disabled="!selectedTemplateId || selectedCount === 0 || isRendering"
        @click="handleDownload"
      >
        <MdiDownload class="mr-2 size-4" />
        {{ $t("components.label_template.print_dialog.download") }}
        {{ outputFormat === "pdf" ? "PDF" : "PNG" }}
      </Button>
      <Button :disabled="!selectedTemplateId || selectedCount === 0 || isRendering" @click="handlePrint">
        <MdiPrinter class="mr-2 size-4" />
        {{ $t("pages.label_templates.batch.print_labels", { count: selectedCount }) }}
      </Button>
    </div>
  </BaseContainer>
</template>
