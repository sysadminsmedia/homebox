<template>
  <Dialog :dialog-id="DialogID.ChangeItemLocation">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          <span> Change location </span>
        </DialogTitle>
      </DialogHeader>
      <LocationSelector v-model="location" />
      <Button @click="updateLocation"> Update location </Button>
    </DialogContent>
  </Dialog>
  <Dialog :dialog-id="DialogID.ChangeItemLabels">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          <span> Change labels </span>
        </DialogTitle>
      </DialogHeader>
      <span> Add labels </span>
      <LabelSelector v-model="addedLabels" :labels="labels" />
      <span> Remove labels </span>
      <LabelSelector v-model="removedLabels" :labels="labels" />
      <Button @click="updateLabels">{{ $t("labels.update_labels") }}</Button>
    </DialogContent>
  </Dialog>
  <MaintenanceEditModal ref="maintenanceEditModal"></MaintenanceEditModal>
  <div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
    <ItemCard
      v-for="(item, index) in props.items"
      :key="item.id"
      v-model="cards[index]"
      :item="item"
      :location-flat-tree="props.locationFlatTree"
    />
  </div>
  <div class="DropdownMenu">
    <DropdownMenu>
      <DropdownMenuTrigger aria-label="Open quick actions menu">
        <Ellipsis v-if="cards.some(c => c.selected)" class="cards-quick-menu-icon size-9 rounded-full border-solid" />
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuItem
          @click="maintenanceEditModal?.openCreateModal(cards.filter(c => c.selected).map(({ item }) => item.id))"
        >
          Add maintenance entry
        </DropdownMenuItem>
        <DropdownMenuItem @click="openDialog(DialogID.ChangeItemLocation)"> Change location </DropdownMenuItem>
        <DropdownMenuItem @click="openDialog(DialogID.ChangeItemLabels)"> Change labels </DropdownMenuItem>
        <DropdownMenuItem @click="deleteItems">{{ $t("global.delete") }}</DropdownMenuItem>
        <DropdownMenuItem @click="duplicateItems">{{ $t("global.duplicate") }}</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { Ellipsis } from "lucide-vue-next";
  import MaintenanceEditModal from "@/components/Maintenance/EditModal.vue";
  import type { ItemSummary, LocationSummary } from "~~/lib/api/types/data-contracts";
  import { useLabelStore } from "~~/stores/labels";
  import { toast } from "@/components/ui/sonner";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DropdownMenu, DropdownMenuContent, DropdownMenuItem } from "~/components/ui/dropdown-menu";

  const maintenanceEditModal = ref<InstanceType<typeof MaintenanceEditModal>>();

  const { t } = useI18n();
  const confirm = useConfirm();
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

  const emit = defineEmits(["refreshItems"]);

  const api = useUserApi();
  const labelStore = useLabelStore();
  const labels = computed(() => labelStore.labels);

  const location = ref<LocationSummary | null>();
  const addedLabels = ref<string[]>([]);
  const removedLabels = ref<string[]>([]);
  let initialLabelsToRemove: string[] = [];

  watch(
    () => props.items,
    () => {
      cards.value = props.items.map(item => {
        return { selected: false, item };
      });
    }
  );

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

      const initialLabels = new Set(
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
      removedLabels.value = [...initialLabels];
      initialLabelsToRemove = removedLabels.value;
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

  async function updateLocation() {
    try {
      await Promise.all(
        cards.value.map(async ({ selected, item }) => {
          if (!selected) return;
          if (!location.value?.id) return;
          const { error } = await api.items.patch(item.id, { id: item.id, location: location.value.id });

          if (error) {
            throw new Error("failed");
          }
        })
      );
    } catch (e) {
      toast.error(t("locations.toast.failed_update_location"));
      return;
    }
    location.value = null;
    closeDialog(DialogID.ChangeItemLocation);
    toast.success(t("locations.toast.location_updated"));
  }

  async function updateLabels() {
    const labelsToRemove = initialLabelsToRemove.filter(il => !removedLabels.value.includes(il));
    if (!addedLabels.value.length && !labelsToRemove.length) return;
    try {
      await Promise.all(
        cards.value.map(async ({ selected, item }) => {
          if (!selected) return;
          let needsUpdate = false;
          addedLabels.value.forEach(l => {
            const index = item.labels.findIndex(il => {
              return il.id === l;
            });
            if (index === -1) {
              item.labels.push(labels.value.find(ls => ls.id === l) as any);
              needsUpdate = true;
            }
          });
          labelsToRemove.forEach(l => {
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
    addedLabels.value = [];
    toast.success(t("labels.toast.labels_updated"));
    closeDialog(DialogID.ChangeItemLabels);
  }

  async function deleteItems() {
    const confirmed = await confirm.open(t("items.delete_items_confirm"));

    if (!confirmed.data) {
      return;
    }

    try {
      await Promise.all(
        cards.value.map(async ({ selected, item }) => {
          if (!selected) return;
          console.log("deleting item", item.name);
          const { error } = await api.items.delete(item.id);
          if (error) {
            throw new Error("failed");
          }
        })
      );
    } catch (e) {
      toast.error(t("items.toast.failed_delete_items"));
    }
    selectedAllCards.value = false;
    emit("refreshItems");
    toast.success(t("items.toast.items_deleted"));
  }

  async function duplicateItems() {
    try {
      await Promise.all(
        cards.value.map(async ({ selected, item }) => {
          if (!selected) return;
          const { error } = await api.items.duplicate(item.id);
          if (error) {
            throw new Error("failed");
          }
        })
      );
    } catch (e) {
      toast.error(t("items.toast.failed_duplicate_item"));
    }
    selectedAllCards.value = false;
    emit("refreshItems");
  }
</script>

<style lang="css">
  .cards-quick-menu-icon {
    position: fixed;
    left: 50%;
    bottom: 5%;
    background-color: whitesmoke;
  }
  .DropdownMenu {
    position: fixed;
    left: 50%;
    bottom: 27%;
  }
</style>
