<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiDownload from "~icons/mdi/download";
  import MdiLoading from "~icons/mdi/loading";
  import MdiPrinter from "~icons/mdi/printer";
  import { Button } from "@/components/ui/button";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
  import { Checkbox } from "@/components/ui/checkbox";
  import BaseContainer from "@/components/Base/Container.vue";

  // Asset ID labels: print-ahead tags for a range of asset IDs, rendered from a
  // label template. Asset IDs no item has yet get a label with just the ID and
  // its /a/<asset id> link, so tags can be applied when items arrive.

  const MAX_RANGE = 1000;

  const { t } = useI18n();
  const measurement = useMeasurement();

  definePageMeta({
    middleware: ["auth"],
  });

  useHead({
    title: computed(() => `HomeBox | ${t("reports.label_generator.title")}`),
  });

  const api = useUserApi();
  const { templates, pending: templatesPending } = useLabelTemplates();
  const { presets } = useLabelPresets();

  const selectedTemplateId = ref<string>("");
  const selectedTemplate = computed(() => templates.value?.find(tpl => tpl.id === selectedTemplateId.value) ?? null);

  // Labels per sheet for templates printed on sheets: their own layout, else their preset's (e.g. Avery 5160)
  const sheetCapacity = computed(() => {
    const tpl = selectedTemplate.value;
    const layout =
      tpl?.sheetLayout ?? (tpl?.preset ? presets.value?.find(p => p.key === tpl.preset)?.sheetLayout : undefined);
    return layout ? layout.columns * layout.rows : null;
  });

  const startAssetId = ref(1);
  const endAssetId = ref(30);
  const skip = ref(0);
  const pageSize = ref<"Letter" | "A4">("Letter");
  const showCutGuides = ref(false);
  const isRendering = ref(false);

  // Default to filling the rest of one sheet when a sheet template is picked
  watch([sheetCapacity, skip], ([capacity]) => {
    if (capacity) {
      endAssetId.value = startAssetId.value + Math.max(capacity - skip.value, 1) - 1;
    }
  });

  const labelCount = computed(() => endAssetId.value - startAssetId.value + 1);

  const rangeError = computed(() => {
    if (!Number.isInteger(startAssetId.value) || startAssetId.value < 1) {
      return t("reports.label_generator.errors.start_invalid");
    }
    if (!Number.isInteger(endAssetId.value) || endAssetId.value < startAssetId.value) {
      return t("reports.label_generator.errors.end_before_start");
    }
    if (labelCount.value > MAX_RANGE) {
      return t("reports.label_generator.errors.too_many", { max: MAX_RANGE });
    }
    if (!Number.isInteger(skip.value) || skip.value < 0) {
      return t("reports.label_generator.errors.skip_invalid");
    }
    return null;
  });

  const canGenerate = computed(() => !!selectedTemplateId.value && !rangeError.value && !isRendering.value);

  async function renderPDF(): Promise<Blob | null> {
    if (!canGenerate.value) return null;

    isRendering.value = true;
    try {
      return await api.labelTemplates.renderAssets(selectedTemplateId.value, {
        startAssetId: startAssetId.value,
        endAssetId: endAssetId.value,
        skipLabels: skip.value,
        format: "pdf",
        pageSize: pageSize.value,
        showCutGuides: showCutGuides.value,
      });
    } catch {
      toast.error(t("components.label_template.toast.render_failed"));
      return null;
    } finally {
      isRendering.value = false;
    }
  }

  async function handleDownload() {
    const blob = await renderPDF();
    if (!blob) return;

    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `asset-labels-${startAssetId.value}-${endAssetId.value}.pdf`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    toast.success(t("components.label_template.toast.downloaded"));
  }

  async function handlePrint() {
    const blob = await renderPDF();
    if (!blob) return;

    const url = URL.createObjectURL(blob);
    const printWindow = window.open(url, "_blank");
    if (printWindow) {
      printWindow.onload = () => {
        printWindow.print();
      };
    }
  }
</script>

<template>
  <BaseContainer>
    <div class="mb-6">
      <h1 class="text-2xl font-semibold">{{ $t("reports.label_generator.title") }}</h1>
      <p class="text-muted-foreground">{{ $t("reports.label_generator.description") }}</p>
    </div>

    <Card class="mb-6">
      <CardHeader>
        <CardTitle>{{ $t("pages.label_templates.batch.select_template") }}</CardTitle>
        <CardDescription>{{ $t("reports.label_generator.template_hint") }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="templatesPending" class="text-muted-foreground">{{ $t("global.loading") }}</div>
        <div v-else-if="!templates || templates.length === 0" class="text-muted-foreground">
          {{ $t("pages.label_templates.batch.no_templates") }}
          <NuxtLink to="/label-templates" class="text-primary underline">
            {{ $t("pages.label_templates.batch.create_template") }}
          </NuxtLink>
        </div>
        <div v-else class="space-y-2">
          <Label for="template">{{ $t("components.label_template.print_dialog.select_template") }}</Label>
          <Select v-model="selectedTemplateId">
            <SelectTrigger id="template">
              <SelectValue :placeholder="$t('components.label_template.print_dialog.select_placeholder')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="tpl in templates" :key="tpl.id" :value="tpl.id">
                {{ tpl.name }} ({{ measurement.formatSize(tpl.width, tpl.height) }})
              </SelectItem>
            </SelectContent>
          </Select>
          <p v-if="sheetCapacity" class="text-sm text-muted-foreground">
            {{ $t("reports.label_generator.sheet_capacity", { count: sheetCapacity }) }}
          </p>
        </div>
      </CardContent>
    </Card>

    <Card class="mb-6">
      <CardHeader>
        <CardTitle>{{ $t("reports.label_generator.range_title") }}</CardTitle>
        <CardDescription>{{ $t("reports.label_generator.range_hint") }}</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div class="space-y-2">
            <Label for="startAssetId">{{ $t("reports.label_generator.asset_start") }}</Label>
            <Input id="startAssetId" v-model.number="startAssetId" type="number" min="1" step="1" />
          </div>
          <div class="space-y-2">
            <Label for="endAssetId">{{ $t("reports.label_generator.asset_end") }}</Label>
            <Input id="endAssetId" v-model.number="endAssetId" type="number" min="1" step="1" />
          </div>
          <div class="space-y-2">
            <Label for="skip">{{ $t("reports.label_generator.skip_first_labels") }}</Label>
            <Input id="skip" v-model.number="skip" type="number" min="0" step="1" />
          </div>
        </div>
        <p v-if="rangeError" class="text-sm text-destructive">{{ rangeError }}</p>
        <p v-else class="text-sm text-muted-foreground">
          {{ $t("reports.label_generator.label_count", { count: labelCount }) }}
        </p>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div class="space-y-2">
            <Label for="pageSize">{{ $t("pages.label_templates.batch.page_size") }}</Label>
            <Select v-model="pageSize">
              <SelectTrigger id="pageSize">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="Letter">Letter ({{ measurement.formatSize(215.9, 279.4) }})</SelectItem>
                <SelectItem value="A4">A4 ({{ measurement.formatSize(210, 297) }})</SelectItem>
              </SelectContent>
            </Select>
            <p v-if="sheetCapacity" class="text-xs text-muted-foreground">
              {{ $t("reports.label_generator.page_size_sheet_hint") }}
            </p>
          </div>
          <div class="flex items-center gap-2 sm:mt-8">
            <Checkbox id="cutGuides" v-model="showCutGuides" />
            <Label for="cutGuides" class="cursor-pointer text-sm font-normal">
              {{ $t("pages.label_templates.batch.show_cut_guides") }}
            </Label>
          </div>
        </div>

        <div class="flex flex-wrap gap-2">
          <Button :disabled="!canGenerate" @click="handleDownload">
            <MdiLoading v-if="isRendering" class="mr-2 size-4 animate-spin" />
            <MdiDownload v-else class="mr-2 size-4" />
            {{ $t("reports.label_generator.download_pdf") }}
          </Button>
          <Button variant="outline" :disabled="!canGenerate" @click="handlePrint">
            <MdiPrinter class="mr-2 size-4" />
            {{ $t("reports.label_generator.print_pdf") }}
          </Button>
        </div>
        <p class="text-xs text-muted-foreground">{{ $t("reports.label_generator.print_tip") }}</p>
      </CardContent>
    </Card>
  </BaseContainer>
</template>
