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
  <ul v-if="selectedCards.length" class="flex flex-row cards-quick-menu">
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
          <a @click="maintenanceEditModal?.openCreateModal(selectedCards.map(i => i.id))">
            <AlphaMCircle class="size-10" aria-hidden="true" />
          </a>
        </TooltipTrigger>
        <TooltipContent> Add maintenance entry </TooltipContent>
      </Tooltip>
    </li>
  </ul>
  <div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
    <ItemCard
      v-for="item in props.items"
      :key="item.id"
      v-model="selectedCards"
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

  const { openDialog } = useDialog();

  const props = defineProps<{
    items: ItemSummary[];
    action?: { action: "selectAll" | "clearAll" }; // using an object instead of a bare string to force watch update
    locationFlatTree?: FlatTreeItem[];
  }>();

  const api = useUserApi();
  const labelStore = useLabelStore();
  const labels = computed(() => labelStore.labels);

  const location = ref();
  const currentLabels = ref<string[]>([]);
  const selectedCards = ref<ItemSummary[]>([]);
  let initialLabels: Set<string>;
  let addedLabels: Set<string>;
  let removedLabels: Set<string>;

  watch(
    () => props.action,
    () => {
      if (props.action?.action === "selectAll") {
        selectedCards.value = props.items;
      } else if (props.action?.action === "clearAll") {
        selectedCards.value = [];
      }
    }
  );

  watch(location, async newLoc => {
    try {
      await Promise.all(
        selectedCards.value.map(async item => {
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

  watch(selectedCards, () => {
    initialLabels = new Set(
      selectedCards.value
        .map(item => {
          return item.labels.map(l => {
            return l.id;
          });
        })
        .flat()
    );
    currentLabels.value = [...initialLabels];
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
        selectedCards.value.map(async item => {
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
