<template>
  <Dialog :dialog-id="DialogID.EditLocationAndLabels">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          <span> Manage location and labels </span>
        </DialogTitle>
      </DialogHeader>
      <LocationSelector v-model="location" />
      <LabelSelector v-model="currentLabels" :labels="labels" />
      <Button @click="updateLabels">{{ $t("labels.update_labels") }}</Button>
    </DialogContent>
  </Dialog>
  <MaintenanceEditModal ref="maintenanceEditModal"></MaintenanceEditModal>
  <ul v-if="cards.some(c => c.selected)" class="cards-quick-menu flex flex-row">
    <li>
      <Tooltip>
        <TooltipTrigger>
          <a @click="openDialog(DialogID.EditLocationAndLabels)">
            <AlphaLCircle class="size-10" aria-hidden="true" />
          </a>
        </TooltipTrigger>
        <TooltipContent> Manage location and labels </TooltipContent>
      </Tooltip>
    </li>
    <li>
      <Tooltip>
        <TooltipTrigger>
          <a @click="maintenanceEditModal?.openCreateModal(cards.map(({ item }) => item.id))">
            <AlphaMCircle class="size-10" aria-hidden="true" />
          </a>
        </TooltipTrigger>
        <TooltipContent> Add maintenance entry </TooltipContent>
      </Tooltip>
    </li>
  </ul>
  <div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
    <ItemCard
      v-for="(item, index) in props.items"
      :key="item.id"
      v-model="cards[index]"
      :item="item"
      :location-flat-tree="props.locationFlatTree"
    />
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { MaintenanceEditModal } from "#components";
  import type { ItemSummary } from "~~/lib/api/types/data-contracts";
  import { useLabelStore } from "~~/stores/labels";
  import { toast } from "@/components/ui/sonner";
  import AlphaLCircle from "~icons/mdi/alpha-l-circle";
  import AlphaMCircle from "~icons/mdi/alpha-m-circle";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

  const maintenanceEditModal = ref<InstanceType<typeof MaintenanceEditModal>>();

  const { t } = useI18n();

  const { openDialog, closeDialog } = useDialog();

  const props = defineProps<{
    items: ItemSummary[];
    locationFlatTree?: FlatTreeItem[];
  }>();
  const selectedAllCards = defineModel<boolean | "indeterminate">({ default: false });

  const cards = ref<Array<{ selected: boolean; item: ItemSummary }>>(
    props.items.map(item => {
      return { selected: false, item };
    })
  );

  const api = useUserApi();
  const labelStore = useLabelStore();
  const labels = computed(() => labelStore.labels);

  const location = ref();
  const currentLabels = ref<string[]>([]);
  let initialLabels: Set<string>;
  let addedLabels: Set<string>;
  let removedLabels: Set<string>;

  watch(
    cards,
    n => {
      const notSelected = n.filter(c => {
        return !c.selected;
      });
      if (notSelected.length) {
        if (notSelected.length !== props.items.length) {
          selectedAllCards.value = "indeterminate";
        } else {
          selectedAllCards.value = false;
        }
      } else {
        selectedAllCards.value = true;
      }

      initialLabels = new Set(
        cards.value
          .filter(({ selected }) => {
            return !!selected;
          })
          .map(({ item }) => {
            return item.labels.map(l => {
              return l.id;
            });
          })
          .flat()
      );
      currentLabels.value = [...initialLabels];
    },
    { deep: true }
  );

  watch(selectedAllCards, () => {
    if (selectedAllCards.value === "indeterminate") return;
    if (selectedAllCards.value) {
      cards.value.forEach(o => {
        o.selected = true;
      });
    } else {
      cards.value.forEach(o => {
        o.selected = false;
      });
    }
  });

  watch(location, async newLoc => {
    if (!newLoc || !newLoc.id) return;
    try {
      await Promise.all(
        cards.value.map(async ({ selected, item }) => {
          if (!selected) return;
          const { error } = await api.items.patch(item.id, { id: item.id, location: newLoc.id });

          if (error) {
            throw new Error("failed");
          }
        })
      );
    } catch (e) {
      toast.error(t("locations.toast.failed_update_location"));
      return;
    }
    toast.success(t("locations.toast.location_updated"));
  });

  watch(currentLabels, clNew => {
    const clNewSet = new Set(clNew);
    addedLabels = clNewSet.difference(initialLabels);
    removedLabels = initialLabels.difference(clNewSet);
  });

  async function updateLabels() {
    initialLabels = new Set(currentLabels.value);
    try {
      await Promise.all(
        cards.value.map(async ({ selected, item }) => {
          if (!selected) return;
          let needsUpdate = false;
          addedLabels.forEach(l => {
            const index = item.labels.findIndex(il => {
              return il.id === l;
            });
            if (index === -1) {
              item.labels.push(labels.value.find(ls => ls.id === l) as any);
              needsUpdate = true;
            }
          });
          removedLabels.forEach(l => {
            const index = item.labels.findIndex(il => {
              return il.id === l;
            });
            if (index >= 0) {
              item.labels.splice(index, 1);
              needsUpdate = true;
            }
          });

          if (needsUpdate) {
            const { error } = await api.items.patch(item.id, { id: item.id, labelIds: item.labels.map(l => l.id) });
            if (error) {
              throw new Error("failed");
            }
          }
        })
      );
    } catch (e) {
      toast.error(t("labels.toast.failed_update_labels"));
    }
    toast.success(t("labels.toast.labels_updated"));
    closeDialog(DialogID.EditLocationAndLabels);
  }
</script>

<style lang="css">
  .cards-quick-menu {
    position: fixed;
    left: 50%;
    bottom: 5%;
    z-index: 2;
  }
</style>
