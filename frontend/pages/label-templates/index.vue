<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import MdiPlus from "~icons/mdi/plus";
  import MdiPrinter from "~icons/mdi/printer";
  import MdiUpload from "~icons/mdi/upload";
  import { toast } from "@/components/ui/sonner";
  import { Button } from "@/components/ui/button";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import LabelTemplateCard from "~/components/LabelTemplate/Card.vue";
  import LabelTemplateCreateModal from "~/components/LabelTemplate/CreateModal.vue";
  import type { LabelTemplateCreate } from "~~/lib/api/types/data-contracts";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

  useHead({
    title: computed(() => `HomeBox | ${t("pages.label_templates.title")}`),
  });

  const { templates, refresh, pending } = useLabelTemplates();
  const { create } = useLabelTemplateActions();
  const { openDialog } = useDialog();

  const handleRefresh = () => refresh();

  // Import: the file is the payload written by a card's "Export JSON", which matches the create request.
  const importInput = ref<HTMLInputElement | null>(null);
  const importing = ref(false);
  // An export is a few KB of JSON plus any embedded images (the editor caps each at 512 KB),
  // so anything beyond this is not a template and must not be read into memory.
  const MAX_IMPORT_BYTES = 4 * 1024 * 1024;

  function parseTemplateExport(raw: unknown): LabelTemplateCreate | null {
    if (!raw || typeof raw !== "object") return null;
    const o = raw as Record<string, unknown>;
    const num = (v: unknown) => (typeof v === "number" && Number.isFinite(v) ? v : NaN);
    const width = num(o.width);
    const height = num(o.height);
    const dpi = num(o.dpi);
    if (typeof o.name !== "string" || !o.name.trim()) return null;
    if (!(width > 0 && width <= 1000) || !(height > 0 && height <= 1000)) return null;
    if (!(dpi >= 72 && dpi <= 1200)) return null;
    if (!o.canvasData || typeof o.canvasData !== "object" || Array.isArray(o.canvasData)) return null;
    return {
      name: o.name.trim().slice(0, 255),
      description: typeof o.description === "string" ? o.description.slice(0, 1000) : "",
      width,
      height,
      dpi,
      preset: typeof o.preset === "string" ? o.preset : null,
      outputFormat: o.outputFormat === "pdf" ? "pdf" : "png",
      // Imported designs start private; the owner can share them afterwards.
      isShared: false,
      canvasData: o.canvasData as Record<string, unknown>,
    };
  }

  async function handleImport(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = "";
    if (!file) return;
    if (file.size > MAX_IMPORT_BYTES) {
      toast.error(t("components.label_template.toast.import_too_large", { max: "4 MB" }));
      return;
    }
    importing.value = true;
    try {
      let parsed: unknown;
      try {
        parsed = JSON.parse(await file.text());
      } catch {
        toast.error(t("components.label_template.toast.import_invalid"));
        return;
      }
      const body = parseTemplateExport(parsed);
      if (!body) {
        toast.error(t("components.label_template.toast.import_invalid"));
        return;
      }
      const created = await create(body);
      toast.success(t("components.label_template.toast.imported", { name: created?.name ?? body.name }));
      await refresh();
    } catch {
      toast.error(t("components.label_template.toast.import_failed"));
    } finally {
      importing.value = false;
    }
  }

  const sharedCount = computed(() => (templates.value ?? []).filter(tpl => tpl.isShared).length);
</script>

<template>
  <BaseContainer>
    <div class="mb-4 flex flex-wrap items-start justify-between gap-3">
      <div>
        <BaseSectionHeader>{{ $t("pages.label_templates.title") }}</BaseSectionHeader>
        <p v-if="templates" class="text-sm text-muted-foreground">
          {{ $t("pages.label_templates.subtitle", { count: templates.length, shared: sharedCount }) }}
        </p>
      </div>
      <div class="flex gap-2">
        <Button variant="outline" as-child>
          <NuxtLink to="/label-templates/batch">
            <MdiPrinter class="mr-2" />
            {{ $t("pages.label_templates.batch.title") }}
          </NuxtLink>
        </Button>
        <Button variant="outline" :disabled="importing" @click="importInput?.click()">
          <MdiUpload class="mr-2" />
          {{ $t("pages.label_templates.import_json") }}
        </Button>
        <input ref="importInput" type="file" accept="application/json,.json" class="hidden" @change="handleImport" />
        <Button @click="openDialog(DialogID.CreateLabelTemplate)">
          <MdiPlus class="mr-2" />
          {{ $t("pages.label_templates.new_template") }}
        </Button>
      </div>
    </div>

    <LabelTemplateCreateModal @created="handleRefresh" />

    <div v-if="pending" class="flex items-center justify-center py-12">
      <div class="text-muted-foreground">{{ $t("global.loading") }}</div>
    </div>

    <div v-else-if="templates && templates.length > 0" class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <LabelTemplateCard
        v-for="tpl in templates"
        :key="tpl.id"
        :template="tpl"
        @deleted="handleRefresh"
        @duplicated="handleRefresh"
      />
    </div>

    <div v-else class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
      <p class="mb-1 font-medium">{{ $t("pages.label_templates.no_templates") }}</p>
      <p class="mb-4 max-w-md text-sm text-muted-foreground">{{ $t("pages.label_templates.empty_hint") }}</p>
      <Button @click="openDialog(DialogID.CreateLabelTemplate)">
        <MdiPlus class="mr-2" />
        {{ $t("components.label_template.create_modal.title") }}
      </Button>
    </div>
  </BaseContainer>
</template>
