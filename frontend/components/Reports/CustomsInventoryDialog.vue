<template>
  <BaseModal :dialog-id="DialogID.CustomsInventoryExport" hide-footer content-class="max-w-4xl">
    <template #title>
      <div class="flex flex-col gap-0.5">
        <span>{{ $t("components.reports.customs_inventory.title") }}</span>
        <span v-if="preview" class="text-sm font-normal text-muted-foreground">
          {{ $t("components.reports.customs_inventory.currency_hint", { currency: preview.currency }) }}
        </span>
      </div>
    </template>

    <div class="grid gap-6 md:grid-cols-2">
      <!-- Options -->
      <div class="space-y-5">
        <fieldset class="space-y-2">
          <legend class="mb-1 text-sm font-semibold">{{ $t("components.reports.customs_inventory.scope") }}</legend>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="scope" type="radio" value="selected" :disabled="itemIds.length === 0" />
            {{ $t("components.reports.customs_inventory.scope_selected", { count: itemIds.length }) }}
          </label>
          <label v-if="search" class="flex items-center gap-2 text-sm">
            <input v-model="scope" type="radio" value="query" />
            {{ $t("components.reports.customs_inventory.scope_query", { count: search.total }) }}
          </label>
        </fieldset>

        <fieldset class="space-y-2">
          <legend class="mb-1 text-sm font-semibold">{{ $t("components.reports.customs_inventory.group_by") }}</legend>
          <div class="flex flex-wrap gap-1">
            <Button
              v-for="g in groupings"
              :key="g.value"
              type="button"
              size="sm"
              :variant="prefs.groupBy === g.value ? 'default' : 'outline'"
              @click="prefs.groupBy = g.value"
            >
              {{ g.label }}
            </Button>
          </div>
          <template v-if="prefs.groupBy === 'tag'">
            <TagSelector v-model="prefs.groupTagIds" :tags="tags" />
            <p class="text-xs text-muted-foreground">
              {{ $t("components.reports.customs_inventory.group_tags_hint") }}
            </p>
          </template>
          <div class="flex items-start gap-2">
            <Checkbox id="customs-merge" v-model="prefs.mergeIdentical" class="mt-0.5" />
            <div>
              <Label for="customs-merge" class="cursor-pointer">
                {{ $t("components.reports.customs_inventory.merge") }}
              </Label>
              <p class="text-xs text-muted-foreground">{{ $t("components.reports.customs_inventory.merge_hint") }}</p>
            </div>
          </div>
        </fieldset>

        <fieldset class="space-y-3">
          <legend class="mb-1 text-sm font-semibold">{{ $t("components.reports.customs_inventory.rules") }}</legend>
          <div class="grid gap-3 sm:grid-cols-2">
            <div>
              <FormDatePicker
                v-model="prefs.referenceDate"
                :label="$t('components.reports.customs_inventory.reference_date')"
              />
              <p class="mt-1 text-xs text-muted-foreground">
                {{ $t("components.reports.customs_inventory.reference_date_hint") }}
              </p>
            </div>
            <div class="space-y-1">
              <Label for="customs-min-months">{{ $t("components.reports.customs_inventory.min_months") }}</Label>
              <Input id="customs-min-months" v-model.number="minMonths" type="number" min="1" max="60" step="1" />
            </div>
          </div>
          <div class="space-y-1">
            <Label for="customs-cutoff">
              {{ $t("components.reports.customs_inventory.cutoff", { months: prefs.minMonths }) }}
            </Label>
            <Input id="customs-cutoff" :model-value="cutoffLabel" readonly disabled />
          </div>
          <TagSelector
            v-model="prefs.dutyFreeTagIds"
            :tags="tags"
            :name="$t('components.reports.customs_inventory.duty_free_tags')"
          />
          <TagSelector
            v-model="prefs.dutiableTagIds"
            :tags="tags"
            :name="$t('components.reports.customs_inventory.dutiable_tags')"
          />
          <p class="text-xs text-muted-foreground">{{ $t("components.reports.customs_inventory.rules_hint") }}</p>
        </fieldset>

        <fieldset class="space-y-3">
          <legend class="mb-1 text-sm font-semibold">{{ $t("components.reports.customs_inventory.header") }}</legend>
          <div class="space-y-1">
            <Label for="customs-title">{{ $t("components.reports.customs_inventory.header_title") }}</Label>
            <Input id="customs-title" v-model="prefs.header.title" maxlength="200" />
          </div>
          <div class="grid gap-3 sm:grid-cols-2">
            <div class="space-y-1">
              <Label for="customs-name">{{ $t("global.name") }}</Label>
              <Input id="customs-name" v-model="prefs.header.name" maxlength="200" />
            </div>
            <div class="space-y-1">
              <Label for="customs-phone">{{ $t("components.reports.customs_inventory.header_phone") }}</Label>
              <Input id="customs-phone" v-model="prefs.header.phone" maxlength="100" />
            </div>
            <div class="space-y-1">
              <Label for="customs-address">{{ $t("components.reports.customs_inventory.header_address") }}</Label>
              <Textarea id="customs-address" v-model="prefs.header.address" rows="2" maxlength="1000" />
            </div>
            <div class="space-y-1">
              <Label for="customs-email">{{ $t("global.email") }}</Label>
              <Input id="customs-email" v-model="prefs.header.email" maxlength="200" />
            </div>
          </div>
          <p class="text-xs text-muted-foreground">
            {{ $t("components.reports.customs_inventory.header_saved_hint") }}
          </p>
        </fieldset>

        <fieldset class="space-y-2">
          <legend class="mb-1 text-sm font-semibold">
            {{ $t("components.reports.customs_inventory.receipts_legend") }}
          </legend>
          <div class="flex items-start gap-2">
            <Checkbox id="customs-receipts" v-model="prefs.includeReceipts" class="mt-0.5" />
            <div>
              <Label for="customs-receipts" class="cursor-pointer">
                {{ $t("components.reports.customs_inventory.receipts") }}
              </Label>
              <p class="text-xs text-muted-foreground">
                {{ $t("components.reports.customs_inventory.receipts_hint") }}
              </p>
            </div>
          </div>
          <p class="text-sm">{{ $t("components.reports.customs_inventory.receipt_types") }}</p>
          <div class="flex flex-wrap gap-x-4 gap-y-1">
            <div v-for="rt in receiptTypes" :key="rt.value" class="flex items-center gap-2">
              <Checkbox
                :id="`customs-receipt-type-${rt.value}`"
                :model-value="prefs.receiptTypes.includes(rt.value)"
                @update:model-value="checked => toggleReceiptType(rt.value, checked === true)"
              />
              <Label :for="`customs-receipt-type-${rt.value}`" class="cursor-pointer font-normal">{{ rt.label }}</Label>
            </div>
          </div>
        </fieldset>
      </div>

      <!-- Pre-export check -->
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-semibold">{{ $t("components.reports.customs_inventory.check") }}</h3>
          <Button
            type="button"
            size="sm"
            variant="ghost"
            :disabled="!prefs.referenceDate || loading"
            @click="runPreview"
          >
            <MdiRefresh class="mr-1 size-4" />
            {{ $t("components.reports.customs_inventory.recheck") }}
          </Button>
        </div>

        <p v-if="!prefs.referenceDate" class="text-sm text-muted-foreground">
          {{ $t("components.reports.customs_inventory.check_needs_date") }}
        </p>
        <div v-else-if="loading && !preview" class="flex items-center justify-center py-8">
          <MdiLoading class="size-6 animate-spin text-muted-foreground" />
        </div>

        <template v-if="preview && prefs.referenceDate">
          <div class="grid grid-cols-2 gap-2" :class="{ 'opacity-60': loading }">
            <div class="rounded-lg border p-3">
              <div class="text-xs text-muted-foreground">
                {{ $t("components.reports.customs_inventory.kpi_positions") }}
              </div>
              <div class="text-lg font-semibold">{{ preview.lineCount }} / {{ preview.itemCount }}</div>
            </div>
            <div class="rounded-lg border p-3">
              <div class="text-xs text-muted-foreground">
                {{ $t("components.reports.customs_inventory.kpi_duty_free") }}
              </div>
              <div class="text-lg font-semibold">{{ money(preview.dutyFreeTotalCents) }}</div>
            </div>
            <div class="rounded-lg border p-3">
              <div class="text-xs text-muted-foreground">
                {{ $t("components.reports.customs_inventory.kpi_dutiable", { count: preview.dutiableItemCount }) }}
              </div>
              <div class="text-lg font-semibold">{{ money(preview.dutiableTotalCents) }}</div>
            </div>
            <div class="rounded-lg border p-3">
              <div class="text-xs text-muted-foreground">
                {{ $t("components.reports.customs_inventory.kpi_total") }}
              </div>
              <div class="text-lg font-semibold">{{ money(preview.grandTotalCents) }}</div>
            </div>
          </div>

          <p v-if="preview.issues.length === 0" class="text-sm text-muted-foreground">
            {{ $t("components.reports.customs_inventory.no_issues") }}
          </p>
          <div class="max-h-96 space-y-2 overflow-y-auto pr-1">
            <div
              v-for="issue in preview.issues"
              :key="issue.code"
              class="rounded-lg border p-2 text-sm"
              :class="issue.severity === 'warning' ? 'border-amber-500/60 bg-amber-500/5' : 'bg-muted/40'"
            >
              <div class="flex items-center justify-between gap-2 font-medium">
                <span>{{ $t(`components.reports.customs_inventory.issue_${issue.code}`) }}</span>
                <span class="text-xs text-muted-foreground">{{ issue.items.length }}</span>
              </div>
              <ul class="mt-1 space-y-0.5">
                <li v-for="it in issue.items" :key="it.id" class="truncate">
                  <NuxtLink
                    v-if="it.name"
                    :to="`/${issue.code === 'excluded_location' ? 'location' : 'item'}/${it.id}`"
                    target="_blank"
                    class="hover:underline"
                  >
                    {{ [it.assetId, it.name].filter(Boolean).join(" ") }}
                  </NuxtLink>
                  <span v-else class="text-muted-foreground">{{ it.id }}</span>
                  <span v-if="it.detail" class="text-muted-foreground">: {{ it.detail }}</span>
                </li>
              </ul>
            </div>
          </div>
          <p class="text-xs text-muted-foreground">{{ $t("components.reports.customs_inventory.check_hint") }}</p>
        </template>
      </div>
    </div>

    <div class="mt-4 flex flex-wrap items-center justify-end gap-2 border-t pt-4">
      <span v-if="warningCount > 0" class="mr-auto text-xs text-muted-foreground">
        {{ $t("components.reports.customs_inventory.warnings_footer", { count: warningCount }, warningCount) }}
      </span>
      <Button type="button" variant="outline" @click="closeDialog(DialogID.CustomsInventoryExport)">
        {{ $t("global.cancel") }}
      </Button>
      <Button type="button" :disabled="!canExport" @click="doExport">
        <MdiLoading v-if="exporting || loading" class="mr-2 size-4 animate-spin" />
        <MdiDownload v-else class="mr-2 size-4" />
        {{
          warningCount > 0
            ? $t("components.reports.customs_inventory.export_anyway")
            : $t("components.reports.customs_inventory.export")
        }}
      </Button>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiDownload from "~icons/mdi/download";
  import MdiLoading from "~icons/mdi/loading";
  import MdiRefresh from "~icons/mdi/refresh";
  import { Button } from "@/components/ui/button";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { Textarea } from "@/components/ui/textarea";
  import { useDialog, DialogID } from "@/components/ui/dialog-provider/utils";
  import BaseModal from "@/components/App/CreateModal.vue";
  import TagSelector from "~/components/Tag/Selector.vue";
  import FormDatePicker from "~/components/Form/DatePicker.vue";
  import type { CustomsInventoryPreview } from "~~/lib/api/types/data-contracts";
  import { useTagStore } from "~/stores/tags";
  import {
    buildRequest,
    cutoffDate,
    DEFAULT_CUSTOMS_INVENTORY_PREFS,
    formatCents,
    sanitizePrefs,
    type CustomsInventoryGrouping,
    type CustomsInventoryPrefs,
    type CustomsInventorySearchContext,
  } from "~~/lib/customs-inventory";

  const { t, locale } = useI18n();
  const api = useUserApi();
  const preferences = useViewPreferences();
  const tagStore = useTagStore();
  const { closeDialog, registerOpenDialogCallback } = useDialog();

  const tags = computed(() => tagStore.tags);
  const itemIds = ref<string[]>([]);
  const search = ref<CustomsInventorySearchContext | null>(null);
  const scope = ref<"selected" | "query">("selected");
  const prefs = ref<CustomsInventoryPrefs>(sanitizePrefs(DEFAULT_CUSTOMS_INVENTORY_PREFS, new Set()));
  const preview = ref<CustomsInventoryPreview | null>(null);
  const loading = ref(false);
  const previewFailed = ref(false);
  const exporting = ref(false);
  let previewSeq = 0;

  const groupings = computed<{ value: CustomsInventoryGrouping; label: string }[]>(() => [
    { value: "location", label: t("components.reports.customs_inventory.group_location") },
    { value: "entityType", label: t("components.reports.customs_inventory.group_entity_type") },
    { value: "tag", label: t("components.reports.customs_inventory.group_tag") },
  ]);

  const receiptTypes = computed(() => [
    { value: "receipt", label: t("items.receipt") },
    { value: "attachment", label: t("items.attachment") },
    { value: "warranty", label: t("items.warranty") },
    { value: "manual", label: t("items.manual") },
  ]);

  // The number input can be emptied or overshoot; keep the saved value valid.
  const minMonths = computed({
    get: () => prefs.value.minMonths,
    set: (v: number) => {
      prefs.value.minMonths = Number.isInteger(v) && v >= 1 && v <= 60 ? v : prefs.value.minMonths;
    },
  });

  const cutoffLabel = computed(() => {
    const c = cutoffDate(prefs.value.referenceDate, prefs.value.minMonths);
    return c ? fmtDate(c, "human", "date") : "";
  });

  const warningCount = computed(() => preview.value?.issues.filter(i => i.severity === "warning").length ?? 0);
  const canExport = computed(
    () => !!preview.value && !!prefs.value.referenceDate && !loading.value && !previewFailed.value && !exporting.value
  );

  function money(cents: number) {
    return formatCents(cents, preview.value?.currency ?? "EUR", locale.value);
  }

  function toggleReceiptType(type: string, on: boolean) {
    const current = prefs.value.receiptTypes.filter(t => t !== type);
    prefs.value.receiptTypes = on ? [...current, type] : current;
  }

  function request() {
    return buildRequest({ scope: scope.value, itemIds: itemIds.value, search: search.value, prefs: prefs.value });
  }

  async function runPreview() {
    if (!prefs.value.referenceDate) {
      preview.value = null;
      return;
    }

    const seq = ++previewSeq;
    loading.value = true;
    const { data, error } = await api.reports.customsInventoryPreview(request());
    if (seq !== previewSeq) {
      return; // a newer preview is on its way
    }
    loading.value = false;

    if (error) {
      previewFailed.value = true;
      toast.error(t("components.reports.customs_inventory.toast_preview_failed"));
      return;
    }
    previewFailed.value = false;
    preview.value = data;
  }

  watchDebounced([prefs, scope], runPreview, { debounce: 400, deep: true });

  async function doExport() {
    exporting.value = true;
    try {
      const { blob, filename } = await api.reports.customsInventoryExport(request());
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      a.click();
      a.remove();
      URL.revokeObjectURL(url);

      preferences.value.customsInventory = JSON.parse(JSON.stringify(prefs.value));
      toast.success(t("components.reports.customs_inventory.toast_exported"));
      closeDialog(DialogID.CustomsInventoryExport);
    } catch {
      toast.error(t("components.reports.customs_inventory.toast_failed"));
    } finally {
      exporting.value = false;
    }
  }

  onMounted(() => {
    registerOpenDialogCallback(DialogID.CustomsInventoryExport, async params => {
      itemIds.value = params.itemIds;
      search.value = params.search ?? null;
      scope.value = itemIds.value.length > 0 || !search.value ? "selected" : "query";
      preview.value = null;
      previewFailed.value = false;

      await tagStore.ensureAllTagsFetched();
      prefs.value = sanitizePrefs(preferences.value.customsInventory, new Set(tags.value.map(tag => tag.id)));
    });
  });
</script>
