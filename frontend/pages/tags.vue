<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import MdiCollapseAllOutline from "~icons/mdi/collapse-all-outline";
  import MdiExpandAllOutline from "~icons/mdi/expand-all-outline";
  import { useTreeState } from "~~/components/Tag/Tree/tree-state";
  import { Button, ButtonGroup } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import type { TagOut } from "~~/lib/api/types/data-contracts";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import BaseCard from "@/components/Base/Card.vue";
  import TagTreeRoot from "~/components/Tag/Tree/Root.vue";
  import type { TagTreeItem } from "~/components/Tag/Tree/types";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

  useHead({
    title: "HomeBox | " + t("global.tags"),
  });

  const api = useUserApi();

  const { data: allTags } = useAsyncData(async () => {
    const { data, error } = await api.tags.getAll();

    if (error) {
      return [];
    }

    return data;
  });

  function buildTagTree(tags: TagOut[]): TagTreeItem[] {
    const nodes = new Map<string, TagTreeItem>();

    for (const tag of tags) {
      nodes.set(tag.id, {
        ...tag,
        children: [],
      });
    }

    const roots: TagTreeItem[] = [];

    for (const tag of tags) {
      const node = nodes.get(tag.id);

      if (!node) {
        continue;
      }

      const parent = tag.parentId ? nodes.get(tag.parentId) : undefined;

      if (parent && parent.id !== node.id) {
        parent.children.push(node);
      } else {
        roots.push(node);
      }
    }

    return roots;
  }

  const tree = computed(() => {
    if (!allTags.value || !Array.isArray(allTags.value)) {
      return [];
    }

    return buildTagTree(allTags.value);
  });

  const tagTreeId = "tagTree";
  const treeState = useTreeState(tagTreeId);

  const route = useRouter();

  onMounted(() => {
    const query = route.currentRoute.value.query;

    if (query && query[tagTreeId]) {
      const data = JSON.parse(query[tagTreeId] as string);

      for (const key in data) {
        treeState.value[key] = data[key];
      }
    }
  });

  watch(
    treeState,
    () => {
      route.replace({
        query: {
          [tagTreeId]: JSON.stringify(treeState.value),
        },
      });
    },
    { deep: true }
  );

  function closeAll() {
    for (const key in treeState.value) {
      treeState.value[key] = false;
    }
  }

  function openTagChildren(items: TagTreeItem[]) {
    for (const item of items) {
      if (item.children.length > 0) {
        treeState.value[item.id.replace(/-/g, "").substring(0, 8)] = true;
        openTagChildren(item.children);
      }
    }
  }

  function openAll() {
    if (!tree.value) return;

    openTagChildren(tree.value);
  }
</script>

<template>
  <BaseContainer>
    <div class="mb-2 flex justify-between">
      <BaseSectionHeader>{{ $t("global.tags") }}</BaseSectionHeader>
      <div>
        <TooltipProvider :delay-duration="0">
          <ButtonGroup>
            <Tooltip>
              <TooltipTrigger>
                <Button size="icon" variant="outline" data-pos="start" @click="openAll">
                  <MdiExpandAllOutline />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{{ $t("locations.expand_tree") }}</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger>
                <Button size="icon" variant="outline" data-pos="end" @click="closeAll">
                  <MdiCollapseAllOutline />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{{ $t("locations.collapse_tree") }}</p>
              </TooltipContent>
            </Tooltip>
          </ButtonGroup>
        </TooltipProvider>
      </div>
    </div>

    <BaseCard>
      <div class="p-2">
        <TagTreeRoot v-if="tree && Array.isArray(tree)" :tags="tree" :tree-id="tagTreeId" />
      </div>
    </BaseCard>
  </BaseContainer>
</template>
