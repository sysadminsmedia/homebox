<script setup lang="ts">
  import { Dialog, DialogContent, DialogFooter, DialogTitle, DialogHeader } from "@/components/ui/dialog";
  import { Button } from "@/components/ui/button";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import type { ItemPatch, ItemSummary, TagOut, LocationSummary } from "~/lib/api/types/data-contracts";
  import LocationSelector from "~/components/Location/Selector.vue";
  import MdiLoading from "~icons/mdi/loading";
  import { toast } from "~/components/ui/sonner";
  import { useI18n } from "vue-i18n";
  import TagSelector from "~/components/Tag/Selector.vue";

  const { closeDialog, registerOpenDialogCallback } = useDialog();

  const api = useUserApi();
  const { t } = useI18n();
  const tagStore = useTagStore();

  const allTags = computed(() => tagStore.tags);

  const items = ref<ItemSummary[]>([]);
  const saving = ref(false);

  const enabled = reactive({
    changeLocation: false,
    addTags: false,
    removeTags: false,
  });

  const newLocation = ref<LocationSummary | null>(null);
  const addTags = ref<string[]>([]);
  const removeTags = ref<string[]>([]);

  const availableToAddTags = ref<TagOut[]>([]);
  const availableToRemoveTags = ref<TagOut[]>([]);

  const intersectTagIds = (items: ItemSummary[]): string[] => {
    if (items.length === 0) return [];
    const counts = new Map<string, number>();
    for (const it of items) {
      const seen = new Set<string>();
      for (const l of it.tags || []) seen.add(l.id);
      for (const id of seen) counts.set(id, (counts.get(id) || 0) + 1);
    }
    return [...counts.entries()].filter(([_, c]) => c === items.length).map(([id]) => id);
  };

  const unionTagIds = (items: ItemSummary[]): string[] => {
    const s = new Set<string>();
    for (const it of items) for (const l of it.tags || []) s.add(l.id);
    return Array.from(s);
  };

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.ItemChangeDetails, params => {
      items.value = params.items;
      enabled.changeLocation = params.changeLocation ?? false;
      enabled.addTags = params.addTags ?? false;
      enabled.removeTags = params.removeTags ?? false;

      if (params.changeLocation && params.items.length > 0) {
        // if all locations are the same then set the current location to said location
        if (
          params.items[0]!.location &&
          params.items.every(item => item.location?.id === params.items[0]!.location?.id)
        ) {
          newLocation.value = params.items[0]!.location;
        }
      }

      if (params.addTags && params.items.length > 0) {
        const intersection = intersectTagIds(params.items);
        availableToAddTags.value = allTags.value.filter(l => !intersection.includes(l.id));
      }

      if (params.removeTags && params.items.length > 0) {
        const union = unionTagIds(params.items);
        availableToRemoveTags.value = allTags.value.filter(l => union.includes(l.id));
      }
    });

    onUnmounted(cleanup);
  });

  const save = async () => {
    const location = newLocation.value;
    const tagsToAdd = addTags.value;
    const tagsToRemove = removeTags.value;
    if (!items.value.length || (enabled.changeLocation && !location)) {
      return;
    }

    saving.value = true;

    await Promise.allSettled(
      items.value.map(async item => {
        const patch: ItemPatch = {
          id: item.id,
        };

        if (enabled.changeLocation) {
          patch.locationId = location!.id;
        }

        let currentTags = item.tags.map(l => l.id);

        if (enabled.addTags) {
          currentTags = currentTags.concat(tagsToAdd);
        }

        if (enabled.removeTags) {
          currentTags = currentTags.filter(l => !tagsToRemove.includes(l));
        }

        if (enabled.addTags || enabled.removeTags) {
          patch.tagIds = Array.from(new Set(currentTags));
        }

        const { error, data } = await api.items.patch(item.id, patch);

        if (error) {
          console.error("failed to update item", item.id, data);
          toast.error(t("components.item.view.change_details.failed_to_update_item"));
          return;
        }
      })
    );

    closeDialog(DialogID.ItemChangeDetails, true);
    enabled.changeLocation = false;
    enabled.addTags = false;
    enabled.removeTags = false;
    items.value = [];
    addTags.value = [];
    removeTags.value = [];
    availableToAddTags.value = [];
    availableToRemoveTags.value = [];
    saving.value = false;
  };
</script>

<template>
  <Dialog :dialog-id="DialogID.ItemChangeDetails">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ $t("components.item.view.change_details.title") }}</DialogTitle>
      </DialogHeader>
      <LocationSelector v-if="enabled.changeLocation" v-model="newLocation" />
      <TagSelector
        v-if="enabled.addTags"
        v-model="addTags"
        :tags="availableToAddTags"
        :name="$t('components.item.view.change_details.add_tags')"
      />
      <TagSelector
        v-if="enabled.removeTags"
        v-model="removeTags"
        :tags="availableToRemoveTags"
        :name="$t('components.item.view.change_details.remove_tags')"
      />
      <DialogFooter>
        <Button type="submit" :disabled="saving || (enabled.changeLocation && !newLocation)" @click="save">
          <span v-if="!saving">{{ $t("global.save") }}</span>
          <MdiLoading v-else class="animate-spin" />
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
