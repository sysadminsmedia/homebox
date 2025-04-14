<script setup lang="ts">
  import { useTreeState } from "~~/components/Location/Tree/tree-state";
  import MdiCollapseAllOutline from "~icons/mdi/collapse-all-outline";
  import MdiExpandAllOutline from "~icons/mdi/expand-all-outline";

  import { ButtonGroup, Button } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import type { TreeItem } from "~/lib/api/types/data-contracts";

  // TODO: eventually move to https://reka-ui.com/docs/components/tree#draggable-sortable-tree

  definePageMeta({
    middleware: ["auth"],
  });

  useHead({
    title: "Homebox | Items",
  });

  const api = useUserApi();

  const { data: tree } = useAsyncData(async () => {
    const { data, error } = await api.locations.getTree({
      withItems: true,
    });

    if (error) {
      return [];
    }

    return data;
  });

  const locationTreeId = "locationTree";

  const treeState = useTreeState(locationTreeId);

  const route = useRouter();

  onMounted(() => {
    // set tree state from query params
    const query = route.currentRoute.value.query;

    if (query && query[locationTreeId]) {
      console.debug("setting tree state from query params");
      const data = JSON.parse(query[locationTreeId] as string);

      for (const key in data) {
        treeState.value[key] = data[key];
      }
    }
  });

  watch(
    treeState,
    () => {
      // Push the current state to the URL
      route.replace({ query: { [locationTreeId]: JSON.stringify(treeState.value) } });
    },
    { deep: true }
  );

  function closeAll() {
    for (const key in treeState.value) {
      treeState.value[key] = false;
    }
  }

  function openItemChildren(items: TreeItem[]) {
    for (const item of items) {
      if (item.children.length > 0) {
        treeState.value[item.id.replace(/-/g, "").substring(0, 8)] = true;
        openItemChildren(item.children);
      }
    }
  }

  function openAll() {
    if (!tree.value) return;

    openItemChildren(tree.value);
  }
</script>

<template>
  <BaseContainer class="mb-16">
    <BaseSectionHeader> {{ $t("menu.locations") }} </BaseSectionHeader>
    <BaseCard>
      <div class="p-4">
        <div class="mb-2 flex justify-end">
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
        <LocationTreeRoot v-if="tree" :locs="tree" :tree-id="locationTreeId" />
      </div>
    </BaseCard>
  </BaseContainer>
</template>
