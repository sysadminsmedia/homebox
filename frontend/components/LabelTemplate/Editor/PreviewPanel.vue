<script setup lang="ts">
  import { watchDebounced } from "@vueuse/core";
  import MdiRefresh from "~icons/mdi/refresh";
  import MdiEye from "~icons/mdi/eye";
  import { Button } from "@/components/ui/button";
  import ItemSelector from "~/components/Item/Selector.vue";
  import type { ItemSummary } from "~/lib/api/types/data-contracts";

  const props = defineProps<{
    templateId: string;
    canvasData: Record<string, unknown>;
    canvasDataJson?: string; // Serialized canvas data for live preview
  }>();

  const api = useUserApi();
  const { render } = useLabelTemplateActions();

  // Item search for preview
  const { query, results, isLoading, triggerSearch } = useItemSearch(api);

  // Selected item for preview
  const selectedItem = ref<ItemSummary | null>(null);

  // Preview state
  const previewUrl = ref<string | null>(null);
  const isRenderingPreview = ref(false);
  const previewError = ref<string | null>(null);

  // Render preview with selected item
  async function renderPreview() {
    // Clean up previous preview
    if (previewUrl.value) {
      URL.revokeObjectURL(previewUrl.value);
      previewUrl.value = null;
    }

    previewError.value = null;

    if (!selectedItem.value || !props.templateId) {
      return;
    }

    isRenderingPreview.value = true;
    try {
      // Pass canvas data for live preview (use current unsaved state)
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
      previewError.value = "Failed to render preview";
      console.error("Preview render error:", err);
    } finally {
      isRenderingPreview.value = false;
    }
  }

  // Watch for canvas data changes and re-render preview (debounced)
  watchDebounced(
    () => props.canvasData,
    () => {
      if (selectedItem.value) {
        renderPreview();
      }
    },
    { debounce: 500, deep: true }
  );

  // Watch for item selection changes
  watch(selectedItem, () => {
    renderPreview();
  });

  // Clean up on unmount
  onBeforeUnmount(() => {
    if (previewUrl.value) {
      URL.revokeObjectURL(previewUrl.value);
    }
  });
</script>

<template>
  <div class="rounded-lg border bg-card p-4">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-start">
      <!-- Left: Controls -->
      <div class="flex shrink-0 flex-col gap-3 sm:w-64">
        <div class="flex items-center gap-2">
          <MdiEye class="size-4 text-muted-foreground" />
          <h3 class="text-sm font-medium">{{ $t("components.label_template.editor.preview.title") }}</h3>
        </div>

        <!-- Item Selector -->
        <ItemSelector
          v-model="selectedItem"
          :label="$t('components.label_template.editor.preview.select_item')"
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
              $t("components.label_template.editor.preview.no_item_selected")
            }}</span>
          </template>
        </ItemSelector>

        <!-- Refresh Button -->
        <Button
          v-if="selectedItem && !isRenderingPreview"
          variant="outline"
          size="sm"
          class="w-full"
          @click="renderPreview"
        >
          <MdiRefresh class="mr-2 size-4" />
          {{ $t("components.label_template.editor.preview.refresh") }}
        </Button>
      </div>

      <!-- Right: Preview Area -->
      <div class="min-h-[120px] flex-1 rounded border bg-gray-50 p-2">
        <div v-if="!selectedItem" class="flex h-full items-center justify-center py-8">
          <p class="text-center text-xs text-muted-foreground">
            {{ $t("components.label_template.editor.preview.select_item_hint") }}
          </p>
        </div>

        <div v-else-if="isRenderingPreview" class="flex h-full items-center justify-center py-8">
          <div class="text-center">
            <div class="mx-auto size-5 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
            <p class="mt-2 text-xs text-muted-foreground">
              {{ $t("components.label_template.editor.preview.rendering") }}
            </p>
          </div>
        </div>

        <div v-else-if="previewError" class="flex h-full flex-col items-center justify-center py-8">
          <p class="text-center text-xs text-destructive">{{ previewError }}</p>
          <Button variant="ghost" size="sm" class="mt-2" @click="renderPreview">
            <MdiRefresh class="mr-1 size-3" />
            {{ $t("components.label_template.editor.preview.retry") }}
          </Button>
        </div>

        <div v-else-if="previewUrl" class="flex flex-col items-center">
          <img :src="previewUrl" alt="Label preview" class="max-w-full" />
        </div>
      </div>
    </div>
  </div>
</template>
