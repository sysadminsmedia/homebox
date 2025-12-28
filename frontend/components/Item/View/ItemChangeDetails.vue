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
  import LabelSelector from "~/components/Label/Selector.vue";

  const { closeDialog, registerOpenDialogCallback } = useDialog();

  const api = useUserApi();
  const { t } = useI18n();
  const labelStore = useTagStore();

  const allLabels = computed(() => labelStore.tags);

  const items = ref<ItemSummary[]>([]);
  const saving = ref(false);

  const enabled = reactive({
    changeLocation: false,
    addLabels: false,
    removeLabels: false,
  });

  const newLocation = ref<LocationSummary | null>(null);
  const addLabels = ref<string[]>([]);
  const removeLabels = ref<string[]>([]);

  const availableToAddLabels = ref<TagOut[]>([]);
  const availableToRemoveLabels = ref<TagOut[]>([]);

  const intersectLabelIds = (items: ItemSummary[]): string[] => {
    if (items.length === 0) return [];
    const counts = new Map<string, number>();
    for (const it of items) {
      const seen = new Set<string>();
      for (const l of it.tags || []) seen.add(l.id);
      for (const id of seen) counts.set(id, (counts.get(id) || 0) + 1);
    }
    return [...counts.entries()].filter(([_, c]) => c === items.length).map(([id]) => id);
  };

  const unionLabelIds = (items: ItemSummary[]): string[] => {
    const s = new Set<string>();
    for (const it of items) for (const l of it.tags || []) s.add(l.id);
    return Array.from(s);
  };

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.ItemChangeDetails, params => {
      items.value = params.items;
      enabled.changeLocation = params.changeLocation ?? false;
      enabled.addLabels = params.addLabels ?? false;
      enabled.removeLabels = params.removeLabels ?? false;

      if (params.changeLocation && params.items.length > 0) {
        // if all locations are the same then set the current location to said location
        if (
          params.items[0]!.location &&
          params.items.every(item => item.location?.id === params.items[0]!.location?.id)
        ) {
          newLocation.value = params.items[0]!.location;
        }
      }

      if (params.addLabels && params.items.length > 0) {
        const intersection = intersectLabelIds(params.items);
        availableToAddLabels.value = allLabels.value.filter(l => !intersection.includes(l.id));
      }

      if (params.removeLabels && params.items.length > 0) {
        const union = unionLabelIds(params.items);
        availableToRemoveLabels.value = allLabels.value.filter(l => union.includes(l.id));
      }
    });

    onUnmounted(cleanup);
  });

  const save = async () => {
    const location = newLocation.value;
    const labelsToAdd = addLabels.value;
    const labelsToRemove = removeLabels.value;
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

        let currentLabels = item.tags.map(l => l.id);

        if (enabled.addLabels) {
          currentLabels = currentLabels.concat(labelsToAdd);
        }

        if (enabled.removeLabels) {
          currentLabels = currentLabels.filter(l => !labelsToRemove.includes(l));
        }

        if (enabled.addLabels || enabled.removeLabels) {
          patch.tagIds = Array.from(new Set(currentLabels));
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
    enabled.addLabels = false;
    enabled.removeLabels = false;
    items.value = [];
    addLabels.value = [];
    removeLabels.value = [];
    availableToAddLabels.value = [];
    availableToRemoveLabels.value = [];
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
      <LabelSelector
        v-if="enabled.addLabels"
        v-model="addLabels"
        :tags="availableToAddLabels"
        :name="$t('components.item.view.change_details.add_tags')"
      />
      <LabelSelector
        v-if="enabled.removeLabels"
        v-model="removeLabels"
        :tags="availableToRemoveLabels"
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
