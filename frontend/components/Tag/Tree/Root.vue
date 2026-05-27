<script setup lang="ts">
  import { Button } from "~~/components/ui/button";
  import { useDialog } from "~/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import type { TagTreeItem } from "./types";
  import TagTreeNode from "./Node.vue";

  type Props = {
    tags: TagTreeItem[];
    treeId: string;
  };

  const { openDialog } = useDialog();

  const props = defineProps<Props>();

  const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: "base" });

  const sortedTags = computed(() => {
    const list = props.tags ?? [];
    return [...list].sort((a, b) => collator.compare(a.name, b.name));
  });
</script>

<template>
  <div>
    <div
      v-if="sortedTags.length === 0"
      class="py-6 text-center text-sm text-muted-foreground"
      role="status"
      aria-live="polite"
    >
      <p class="mx-auto max-w-xs">
        {{ $t("tags.no_results") }}
      </p>
      <Button
        class="mt-3"
        variant="outline"
        size="sm"
        type="button"
        :aria-label="$t('components.tag.create_modal.title') || $t('global.create')"
        @click="openDialog(DialogID.CreateTag)"
      >
        {{ $t("components.tag.create_modal.title") || $t("global.create") }}
      </Button>
    </div>

    <ul role="tree" :aria-labelledby="treeId" class="space-y-1">
      <li v-for="item in sortedTags" :key="item.id" role="treeitem">
        <TagTreeNode :item="item" :tree-id="treeId" />
      </li>
    </ul>
  </div>
</template>
