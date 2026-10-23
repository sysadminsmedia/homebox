<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { watchDebounced } from "@vueuse/core";
  import { toast } from "@/components/ui/sonner";
  import type { EntitySummary, TagSummary } from "~~/lib/api/types/data-contracts";
  import type { FlatTreeItem } from "~~/composables/use-location-helpers";
  import { useTagStore } from "~~/stores/tags";
  import { useLocationStore } from "~~/stores/locations";
  import MdiLoading from "~icons/mdi/loading";
  import MdiPrinter from "~icons/mdi/printer";
  import MdiDownload from "~icons/mdi/download";
  import MdiDesktopClassic from "~icons/mdi/desktop-classic";
  import MdiDrag from "~icons/mdi/drag";
  import MdiArrowLeft from "~icons/mdi/arrow-left";
  import { Input } from "~/components/ui/input";
  import { Button } from "@/components/ui/button";
  import { Label } from "@/components/ui/label";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Badge } from "@/components/ui/badge";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
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
  const preferences = useViewPreferences();

  // ---------------------------------------------------------------------------
  // What to print: items or locations
  // ---------------------------------------------------------------------------
  type Kind = "items" | "locations";
  const kind = ref<Kind>("items");

  const loading = ref(false);
  const results = ref<EntitySummary[]>([]);
  const query = ref("");

  const locationsStore = useLocationStore();
  const tagStore = useTagStore();
  const locationFlatTree = useFlatLocations();
  const tags = computed(() => tagStore.tags);
  const selectedLocations = ref<FlatTreeItem[]>([]);
  const selectedTags = ref<TagSummary[]>([]);

  onMounted(async () => {
    await Promise.all([locationsStore.ensureLocationsFetched(), tagStore.ensureAllTagsFetched()]);
    await search();
  });

  async function search() {
    loading.value = true;
    try {
      if (kind.value === "locations") {
        const { data, error } = await api.items.getLocations({ filterChildren: false });
        if (error) throw new Error("locations");
        const q = query.value.trim().toLowerCase();
        results.value = (data ?? []).filter(l => !q || l.name.toLowerCase().includes(q));
      } else {
        const { data, error } = await api.items.getAll({
          q: query.value || "",
          parentIds: selectedLocations.value.map(l => l.id),
          tags: selectedTags.value.map(l => l.id),
          pageSize: 100,
        });
        if (error) throw new Error("items");
        results.value = data.items || [];
      }
    } catch {
      toast.error(t("items.toast.failed_search_items"));
    } finally {
      loading.value = false;
    }
  }

  watchDebounced([query, selectedTags, selectedLocations], search, { debounce: 250, maxWait: 1000 });
  watch(kind, () => {
    selectedIds.value = [];
    quantities.value = {};
    search();
  });

  // ---------------------------------------------------------------------------
  // Selection (ordered) and copies
  // ---------------------------------------------------------------------------
  const selectedIds = ref<string[]>([]);
  const quantities = ref<Record<string, number>>({});
  const getQuantity = (id: string) => quantities.value[id] ?? 1;
  const setQuantity = (id: string, qty: number) => {
    quantities.value[id] = Math.max(1, Math.min(100, Number.isFinite(qty) ? qty : 1));
  };

  const isSelected = (id: string) => selectedIds.value.includes(id);
  function toggle(id: string, on?: boolean) {
    const has = isSelected(id);
    const want = on ?? !has;
    if (want && !has) selectedIds.value = [...selectedIds.value, id];
    if (!want && has) selectedIds.value = selectedIds.value.filter(x => x !== id);
  }
  const allSelected = computed(() => results.value.length > 0 && results.value.every(r => isSelected(r.id)));
  function toggleAll() {
    if (allSelected.value) {
      selectedIds.value = [];
    } else {
      const current = new Set(selectedIds.value);
      selectedIds.value = [...selectedIds.value, ...results.value.filter(r => !current.has(r.id)).map(r => r.id)];
    }
  }

  const orderOf = (id: string) => selectedIds.value.indexOf(id) + 1;
  const totalLabels = computed(() => selectedIds.value.reduce((sum, id) => sum + getQuantity(id), 0));
  const expandedIds = computed(() => selectedIds.value.flatMap(id => Array(getQuantity(id)).fill(id) as string[]));

  // Drag to reorder among selected rows
  const dragId = ref<string | null>(null);
  function onDragStart(id: string) {
    if (isSelected(id)) dragId.value = id;
  }
  function onDrop(targetId: string) {
    const from = dragId.value;
    dragId.value = null;
    if (!from || from === targetId || !isSelected(targetId)) return;
    const order = selectedIds.value.filter(x => x !== from);
    order.splice(order.indexOf(targetId), 0, from);
    selectedIds.value = order;
  }

  // ---------------------------------------------------------------------------
  // Template, sheet options, destination
  // ---------------------------------------------------------------------------
  const { templates, pending: templatesPending } = useLabelTemplates();
  const { find: findPreset, title: presetTitle } = useLabelPresetLookup();
  const selectedTemplateId = ref<string>("");
  watch(
    templates,
    list => {
      if (!list || selectedTemplateId.value) return;
      selectedTemplateId.value =
        list.find(tpl => tpl.id === preferences.value.defaultTemplateId)?.id ?? list[0]?.id ?? "";
    },
    { immediate: true }
  );
  const selectedTemplate = computed(() => templates.value?.find(tpl => tpl.id === selectedTemplateId.value) ?? null);
  const preset = computed(() => findPreset(selectedTemplate.value?.preset));
  const isSheet = computed(() => preset.value?.kind === "sheet");
  const perSheet = computed(() => preset.value?.perSheet ?? 0);

  const skipLabels = ref(0);
  const showCutGuides = ref(false);
  const pageSize = ref<"Letter" | "A4">("Letter");
  const sheetCount = computed(() =>
    perSheet.value ? Math.ceil((totalLabels.value + skipLabels.value) / perSheet.value) : 0
  );

  const { printLabels, printLocationLabels } = useDirectPrint();
  const { labelPrinting } = useServerLabelPrinting();

  const DOWNLOAD = "_download";
  const BROWSER = "_browser";
  const SERVER = "_server";
  const destination = ref<string>(DOWNLOAD);

  interface Destination {
    id: string;
    label: string;
    icon: Component;
    hint?: string;
    disabled?: boolean;
  }
  const destinations = computed<Destination[]>(() => {
    const list: Destination[] = [
      {
        id: DOWNLOAD,
        label:
          isSheet.value || totalLabels.value > 1
            ? t("pages.label_templates.batch.download_pdf")
            : t("pages.label_templates.batch.download_png"),
        icon: MdiDownload,
        hint: isSheet.value ? t("pages.label_templates.batch.sheets_need_pdf") : undefined,
      },
      { id: BROWSER, label: t("pages.label_templates.batch.print_browser"), icon: MdiDesktopClassic },
    ];
    if (labelPrinting.value) {
      // Server-side printing renders one label per page; sheets go out as PDF.
      list.push({
        id: SERVER,
        label: t("components.label_template.print_dialog.label_printer"),
        icon: MdiPrinter,
        disabled: isSheet.value,
        hint: isSheet.value
          ? t("pages.label_templates.batch.printer_not_for_sheets")
          : t("components.label_template.print_dialog.label_printer_hint"),
      });
    }
    return list;
  });
  watch(destinations, list => {
    if (list.find(d => d.id === destination.value)?.disabled) destination.value = DOWNLOAD;
  });
  // Prefer the server's label printer once we know it exists, unless the user already chose
  let destinationTouched = false;
  function pickDestination(id: string) {
    destinationTouched = true;
    destination.value = id;
  }
  watch(
    [labelPrinting, isSheet],
    ([enabled, sheet]) => {
      if (enabled && !sheet && !destinationTouched) {
        destination.value = SERVER;
      }
    },
    { immediate: true }
  );

  // Sheet preview (first sheet) or a single label
  const previewUrl = ref<string | null>(null);
  const isLoadingPreview = ref(false);

  async function renderBlob(format: "png" | "pdf", ids: string[]): Promise<Blob> {
    const body = {
      format,
      pageSize: pageSize.value,
      showCutGuides: showCutGuides.value,
      skipLabels: isSheet.value ? skipLabels.value : 0,
      canvasData: "",
    };
    return kind.value === "locations"
      ? api.labelTemplates.renderLocations(selectedTemplateId.value, { ...body, locationIds: ids })
      : api.labelTemplates.render(selectedTemplateId.value, { ...body, itemIds: ids });
  }

  async function updatePreview() {
    if (previewUrl.value) {
      URL.revokeObjectURL(previewUrl.value);
      previewUrl.value = null;
    }
    if (!selectedTemplateId.value) return;
    const ids = isSheet.value ? expandedIds.value.slice(0, perSheet.value || 1) : selectedIds.value.slice(0, 1);
    if (ids.length === 0) return;
    isLoadingPreview.value = true;
    try {
      const blob = await renderBlob("png", ids);
      previewUrl.value = URL.createObjectURL(blob);
    } catch {
      previewUrl.value = null;
    } finally {
      isLoadingPreview.value = false;
    }
  }
  watchDebounced([selectedTemplateId, expandedIds, isSheet], updatePreview, { debounce: 400, deep: true });

  // ---------------------------------------------------------------------------
  // Go
  // ---------------------------------------------------------------------------
  const isWorking = ref(false);
  const isBusy = computed(() => isWorking.value);
  const canGo = computed(() => !!selectedTemplateId.value && selectedIds.value.length > 0 && !isBusy.value);

  const summary = computed(() => {
    const base = t("pages.label_templates.batch.summary", {
      labels: totalLabels.value,
      count: selectedIds.value.length,
      template: selectedTemplate.value?.name ?? "",
    });
    return isSheet.value && sheetCount.value
      ? `${base} · ${t(`pages.label_templates.batch.sheet_count_${sheetCount.value === 1 ? "one" : "many"}`, { count: sheetCount.value })}`
      : base;
  });

  const actionLabel = computed(() => {
    if (destination.value === DOWNLOAD) {
      return isSheet.value || totalLabels.value > 1
        ? t(`pages.label_templates.batch.action_download_pdf_${(sheetCount.value || 1) === 1 ? "one" : "many"}`, {
            count: sheetCount.value || 1,
          })
        : t("pages.label_templates.batch.action_download_png");
    }
    if (destination.value === BROWSER)
      return t("pages.label_templates.batch.action_print_browser", { count: totalLabels.value });
    return t("pages.label_templates.batch.send_to_printer", { count: totalLabels.value });
  });

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

  async function go() {
    if (!canGo.value) return;
    isWorking.value = true;
    try {
      const pdf = isSheet.value || totalLabels.value > 1;
      if (destination.value === DOWNLOAD) {
        const blob = await renderBlob(pdf ? "pdf" : "png", pdf ? expandedIds.value : selectedIds.value);
        triggerDownload(blob, `labels-batch-${Date.now()}.${pdf ? "pdf" : "png"}`);
        toast.success(t("components.label_template.toast.downloaded"));
      } else if (destination.value === BROWSER) {
        const blob = await renderBlob(pdf ? "pdf" : "png", pdf ? expandedIds.value : selectedIds.value);
        const url = URL.createObjectURL(blob);
        const w = window.open(url, "_blank");
        if (w) w.onload = () => w.print();
        toast.success(t("components.label_template.toast.printed"));
      } else if (destination.value === SERVER) {
        const withQty = selectedIds.value.map(id => ({ id, quantity: getQuantity(id) }));
        const result =
          kind.value === "locations"
            ? await printLocationLabels(selectedTemplateId.value, withQty)
            : await printLabels(selectedTemplateId.value, withQty);
        if (!result?.success) {
          toast.error(result?.message || t("pages.label_templates.batch.direct_print_failed"));
          return;
        }
        toast.success(
          result.message || t("pages.label_templates.batch.direct_print_success", { count: result.labelCount })
        );
      }
    } catch {
      toast.error(
        t(
          destination.value === SERVER
            ? "components.label_template.toast.print_failed"
            : "components.label_template.toast.render_failed"
        )
      );
    } finally {
      isWorking.value = false;
    }
  }

  onBeforeUnmount(() => {
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value);
  });
</script>

<template>
  <BaseContainer class="max-w-none">
    <div class="mb-3 flex flex-wrap items-center gap-3">
      <Button variant="ghost" size="icon" as-child>
        <NuxtLink to="/label-templates" :aria-label="$t('components.label_template.editor.header.back')">
          <MdiArrowLeft class="size-5" />
        </NuxtLink>
      </Button>
      <div>
        <h1 class="text-xl font-semibold">{{ $t("pages.label_templates.batch.title") }}</h1>
        <p class="text-sm text-muted-foreground">{{ $t("pages.label_templates.batch.description") }}</p>
      </div>
    </div>

    <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_400px]">
      <!-- Left: pick what to print -->
      <section class="flex min-h-[420px] flex-col rounded-lg border bg-card">
        <div class="flex flex-col gap-3 border-b p-4">
          <div class="flex flex-wrap items-center gap-2">
            <Input v-model="query" :placeholder="$t('global.search')" class="min-w-48 flex-1" />
            <template v-if="kind === 'items'">
              <SearchFilter v-model="selectedLocations" :label="$t('global.locations')" :options="locationFlatTree" />
              <SearchFilter v-model="selectedTags" :label="$t('global.tags')" :options="tags" />
            </template>
            <div role="group" class="inline-flex h-9 overflow-hidden rounded-md border text-sm">
              <button
                type="button"
                class="px-3"
                :class="kind === 'items' ? 'bg-accent font-semibold' : 'text-muted-foreground'"
                :aria-pressed="kind === 'items'"
                @click="kind = 'items'"
              >
                {{ $t("pages.label_templates.batch.kind_items") }}
              </button>
              <button
                type="button"
                class="border-l px-3"
                :class="kind === 'locations' ? 'bg-accent font-semibold' : 'text-muted-foreground'"
                :aria-pressed="kind === 'locations'"
                @click="kind = 'locations'"
              >
                {{ $t("pages.label_templates.batch.kind_locations") }}
              </button>
            </div>
          </div>
          <div class="flex items-center justify-between text-sm text-muted-foreground">
            <label class="flex items-center gap-2">
              <Checkbox
                :model-value="allSelected"
                :aria-label="$t('pages.label_templates.batch.select_all')"
                @update:model-value="toggleAll"
              />
              {{
                $t("pages.label_templates.batch.results_selected", {
                  results: results.length,
                  selected: selectedIds.length,
                })
              }}
            </label>
            <span class="hidden sm:inline">{{ $t("pages.label_templates.batch.drag_hint") }}</span>
          </div>
        </div>

        <div v-if="loading" class="flex flex-1 items-center justify-center py-12">
          <MdiLoading class="size-6 animate-spin text-muted-foreground" />
        </div>
        <div
          v-else-if="results.length === 0"
          class="flex flex-1 items-center justify-center p-8 text-center text-sm text-muted-foreground"
        >
          {{ $t("pages.label_templates.batch.no_items") }}
        </div>
        <ul v-else class="max-h-[60vh] divide-y overflow-y-auto">
          <li
            v-for="row in results"
            :key="row.id"
            class="flex items-center gap-3 px-4 py-2"
            :class="{ 'bg-primary/5': isSelected(row.id) }"
            :draggable="isSelected(row.id)"
            @dragstart="onDragStart(row.id)"
            @dragover.prevent
            @drop="onDrop(row.id)"
          >
            <Checkbox
              :model-value="isSelected(row.id)"
              :aria-label="row.name"
              @update:model-value="toggle(row.id, $event === true)"
            />
            <span class="w-7 shrink-0 text-center text-xs text-muted-foreground">
              <Badge v-if="isSelected(row.id)" variant="secondary" class="px-1.5">{{ orderOf(row.id) }}</Badge>
            </span>
            <MdiDrag
              class="size-4 shrink-0"
              :class="isSelected(row.id) ? 'cursor-grab text-muted-foreground' : 'text-transparent'"
            />
            <div class="min-w-0 flex-1">
              <div class="truncate font-medium">{{ row.name }}</div>
              <div class="truncate text-xs text-muted-foreground">
                <template v-if="kind === 'items'">
                  {{ row.parent?.name || $t("pages.label_templates.batch.no_location") }}
                  <span v-if="row.assetId && row.assetId !== '0'" class="ml-2">#{{ row.assetId }}</span>
                </template>
                <template v-else>
                  {{ $t("pages.label_templates.batch.item_count", { count: row.itemCount ?? 0 }) }}
                </template>
              </div>
            </div>
            <Input
              :model-value="getQuantity(row.id)"
              type="number"
              min="1"
              max="100"
              class="h-8 w-16 text-center"
              :disabled="!isSelected(row.id)"
              :aria-label="$t('components.label_template.print_dialog.copies')"
              @update:model-value="setQuantity(row.id, Number($event))"
            />
          </li>
        </ul>
      </section>

      <!-- Right: template, output, destination -->
      <aside class="flex flex-col gap-4">
        <section class="space-y-3 rounded-lg border bg-card p-4">
          <div class="flex items-center justify-between">
            <Label class="text-sm font-semibold">{{ $t("components.label_template.print_dialog.template") }}</Label>
            <NuxtLink to="/label-templates" class="text-xs text-muted-foreground hover:underline">
              {{ $t("components.label_template.print_dialog.manage_templates") }}
            </NuxtLink>
          </div>
          <div v-if="templatesPending" class="text-sm text-muted-foreground">{{ $t("global.loading") }}</div>
          <div v-else-if="!templates || templates.length === 0" class="text-sm text-muted-foreground">
            {{ $t("pages.label_templates.batch.no_templates") }}
            <NuxtLink to="/label-templates" class="text-primary underline">
              {{ $t("pages.label_templates.batch.create_template") }}
            </NuxtLink>
          </div>
          <Select v-else v-model="selectedTemplateId">
            <SelectTrigger>
              <SelectValue :placeholder="$t('components.label_template.print_dialog.select_placeholder')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="tpl in templates" :key="tpl.id" :value="tpl.id">
                {{ tpl.name }}
                <span class="ml-1 text-muted-foreground">· {{ formatLabelSize(tpl.width, tpl.height) }}</span>
              </SelectItem>
            </SelectContent>
          </Select>
          <p v-if="selectedTemplate?.preset" class="text-xs text-muted-foreground">
            {{ presetTitle(selectedTemplate.preset) }}
          </p>

          <div class="flex flex-col items-center gap-2 rounded-lg bg-muted/60 p-3">
            <div v-if="isLoadingPreview" class="flex h-32 items-center justify-center text-sm text-muted-foreground">
              {{ $t("components.label_template.print_dialog.loading_preview") }}
            </div>
            <img
              v-else-if="previewUrl"
              :src="previewUrl"
              alt=""
              class="max-h-64 max-w-full rounded-sm bg-white shadow"
            />
            <div v-else class="flex h-32 items-center justify-center text-center text-sm text-muted-foreground">
              {{ $t("pages.label_templates.batch.preview_hint") }}
            </div>
            <span v-if="isSheet && selectedIds.length" class="text-xs text-muted-foreground">
              {{ $t("pages.label_templates.batch.sheet_preview", { per: perSheet, sheets: sheetCount }) }}
            </span>
          </div>

          <div v-if="isSheet" class="flex flex-wrap items-center gap-3 text-sm">
            <label class="flex items-center gap-2">
              {{ $t("pages.label_templates.batch.skip_first") }}
              <Input
                v-model.number="skipLabels"
                type="number"
                min="0"
                :max="Math.max(perSheet - 1, 0)"
                class="h-8 w-16 text-center"
              />
            </label>
            <label class="flex items-center gap-2">
              <Checkbox v-model="showCutGuides" />
              {{ $t("pages.label_templates.batch.show_cut_guides") }}
            </label>
            <label class="ml-auto flex items-center gap-2">
              {{ $t("pages.label_templates.batch.page_size") }}
              <Select v-model="pageSize">
                <SelectTrigger class="h-8 w-24">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="Letter">Letter</SelectItem>
                  <SelectItem value="A4">A4</SelectItem>
                </SelectContent>
              </Select>
            </label>
          </div>
        </section>

        <section class="space-y-2 rounded-lg border bg-card p-4">
          <Label class="text-sm font-semibold">{{ $t("components.label_template.print_dialog.send_to") }}</Label>
          <div role="radiogroup" class="flex flex-col gap-1.5">
            <button
              v-for="dest in destinations"
              :key="dest.id"
              type="button"
              role="radio"
              :aria-checked="destination === dest.id"
              :disabled="dest.disabled"
              class="flex items-center gap-3 rounded-md border px-3 py-2 text-left text-sm disabled:cursor-not-allowed disabled:opacity-50"
              :class="destination === dest.id ? 'border-primary bg-primary/5 font-medium ring-1 ring-primary' : ''"
              @click="pickDestination(dest.id)"
            >
              <component :is="dest.icon" class="size-4 shrink-0 text-muted-foreground" />
              <span class="flex-1 whitespace-nowrap">{{ dest.label }}</span>
              <span v-if="dest.hint" class="min-w-0 truncate text-right text-xs font-normal text-muted-foreground">{{
                dest.hint
              }}</span>
            </button>
          </div>
          <p v-if="!labelPrinting" class="text-xs text-muted-foreground">
            {{ $t("components.label_template.print_dialog.server_print_disabled_hint") }}
          </p>
        </section>
      </aside>
    </div>

    <!-- Footer -->
    <div
      class="sticky bottom-0 mt-4 flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-card px-4 py-3 shadow-sm"
    >
      <span class="text-sm text-muted-foreground">
        <template v-if="selectedIds.length">{{ summary }}</template>
        <template v-else>{{ $t("pages.label_templates.batch.nothing_selected") }}</template>
      </span>
      <div class="flex gap-2">
        <Button variant="outline" :disabled="selectedIds.length === 0" @click="selectedIds = []">
          {{ $t("pages.label_templates.batch.clear_selection") }}
        </Button>
        <Button :disabled="!canGo" @click="go">
          <MdiLoading v-if="isBusy" class="mr-2 size-4 animate-spin" />
          <MdiPrinter v-else class="mr-2 size-4" />
          {{ actionLabel }}
        </Button>
      </div>
    </div>
  </BaseContainer>
</template>
