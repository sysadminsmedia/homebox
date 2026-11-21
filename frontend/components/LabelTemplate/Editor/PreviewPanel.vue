<script setup lang="ts">
  import { watchDebounced } from "@vueuse/core";
  import { useI18n } from "vue-i18n";
  import MdiRefresh from "~icons/mdi/refresh";
  import { Button } from "@/components/ui/button";
  import ItemSelector from "~/components/Item/Selector.vue";
  import type { EntitySummary } from "~/lib/api/types/data-contracts";

  const { t } = useI18n();
  const props = defineProps<{
    templateId: string;
    canvasData: Record<string, unknown>;
    canvasDataJson?: string; // Serialized canvas data for live preview
    width: number; // label width in mm, for the preview box aspect ratio
    height: number;
    dpi: number;
  }>();

  const emit = defineEmits<{
    "update:item": [item: EntitySummary | null];
  }>();

  const api = useUserApi();
  const { render } = useLabelTemplateActions();

  const { query, results, isLoading, triggerSearch } = useItemSearch(api);

  const selectedItem = ref<EntitySummary | null>(null);
  const previewUrl = ref<string | null>(null);
  const isRenderingPreview = ref(false);
  const previewError = ref<string | null>(null);

  const aspectRatio = computed(() => `${props.width} / ${props.height}`);

  async function renderPreview() {
    if (previewUrl.value) {
      URL.revokeObjectURL(previewUrl.value);
      previewUrl.value = null;
    }
    previewError.value = null;

    if (!selectedItem.value || !props.templateId) return;

    isRenderingPreview.value = true;
    try {
      const blob = await render(
        props.templateId,
        [selectedItem.value.id],
        "png",
        "Letter",
        false,
        props.canvasDataJson
      );
      previewUrl.value = URL.createObjectURL(blob);
    } catch (err) {
      previewError.value = t("components.label_template.editor.preview.render_failed");
      console.error("Preview render error:", err);
    } finally {
      isRenderingPreview.value = false;
    }
  }

  watchDebounced(
    () => props.canvasData,
    () => {
      if (selectedItem.value) renderPreview();
    },
    { debounce: 500, deep: true }
  );

  watch(selectedItem, item => {
    emit("update:item", item);
    renderPreview();
  });

  onBeforeUnmount(() => {
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value);
  });

  defineExpose({ selectedItem, renderPreview });
</script>

<template>
  <section
    class="flex min-w-0 flex-1 flex-col gap-2 2xl:min-w-[360px]"
    :aria-label="$t('components.label_template.editor.preview.print_preview')"
  >
    <div class="flex flex-wrap items-center justify-between gap-2">
      <span class="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
        {{ $t("components.label_template.editor.preview.print_preview") }}
      </span>
      <div class="w-56 max-w-full">
        <ItemSelector
          v-model="selectedItem"
          :items="results"
          item-text="name"
          item-value="id"
          :is-loading="isLoading"
          :trigger-search="triggerSearch"
          :search="query"
          @update:search="query = $event"
        >
          <template #display="{ item }">
            <span v-if="item" class="truncate">{{ (item as { name: string }).name }}</span>
            <span v-else class="text-muted-foreground">{{
              $t("components.label_template.editor.preview.pick_item")
            }}</span>
          </template>
        </ItemSelector>
      </div>
    </div>

    <div
      class="flex w-full max-w-[520px] items-center justify-center overflow-hidden rounded border bg-white shadow"
      :style="{ aspectRatio }"
    >
      <p v-if="!selectedItem" class="px-6 text-center text-xs text-muted-foreground">
        {{ $t("components.label_template.editor.preview.select_item_hint") }}
      </p>
      <div v-else-if="isRenderingPreview" class="text-center">
        <div class="mx-auto size-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
      <div v-else-if="previewError" class="flex flex-col items-center gap-2">
        <p class="text-xs text-destructive">{{ previewError }}</p>
        <Button variant="ghost" size="sm" @click="renderPreview">
          <MdiRefresh class="mr-1 size-3" />
          {{ $t("components.label_template.editor.preview.retry") }}
        </Button>
      </div>
      <img v-else-if="previewUrl" :src="previewUrl" alt="" class="size-full object-contain" />
    </div>

    <p v-if="selectedItem" class="max-w-[520px] text-xs leading-relaxed text-muted-foreground">
      {{ $t("components.label_template.editor.preview.render_hint", { dpi: props.dpi }) }}
    </p>
  </section>
</template>
