<template>
  <Dialog :dialog-id="DialogID.EditMaintenance">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          {{ entry.id ? $t("maintenance.modal.edit_title") : $t("maintenance.modal.new_title") }}
        </DialogTitle>
      </DialogHeader>

      <form class="flex flex-col gap-2" @submit.prevent="dispatchFormSubmit">
        <FormTextField v-model="entry.name" autofocus :label="$t('maintenance.modal.entry_name')" />
        <DatePicker v-model="entry.completedDate" :label="$t('maintenance.modal.completed_date')" />
        <DatePicker v-model="entry.scheduledDate" :label="$t('maintenance.modal.scheduled_date')" />
        <FormTextArea v-model="entry.description" :label="$t('maintenance.modal.notes')" />
        <FormTextField v-model="entry.cost" autofocus :label="$t('maintenance.modal.cost')" />

        <DialogFooter>
          <Button type="submit">
            <MdiPost />
            {{ entry.id ? $t("maintenance.modal.edit_action") : $t("maintenance.modal.new_action") }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { toast } from "@/components/ui/sonner";
  import type { MaintenanceEntry, MaintenanceEntryWithDetails } from "~~/lib/api/types/data-contracts";
  import MdiPost from "~icons/mdi/post";
  import DatePicker from "~~/components/Form/DatePicker.vue";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import Button from "@/components/ui/button/Button.vue";

  const { openDialog, closeDialog } = useDialog();

  const { t } = useI18n();
  const api = useUserApi();

  const emit = defineEmits(["changed"]);

  const entry = reactive({
    id: null as string | null,
    name: "",
    completedDate: null as Date | null,
    scheduledDate: null as Date | null,
    description: "",
    cost: "",
    itemIds: null as string[] | null,
  });

  async function dispatchFormSubmit() {
    if (entry.id) {
      await editEntry();
      return;
    }

    await createEntry();
  }

  async function createEntry() {
    if (!entry.itemIds || !entry.itemIds.length) {
      return;
    }
    try {
      await Promise.all(
        entry.itemIds.map(async itemId => {
          const { error } = await api.items.maintenance.create(itemId, {
            name: entry.name,
            completedDate: entry.completedDate ?? "",
            scheduledDate: entry.scheduledDate ?? "",
            description: entry.description,
            cost: parseFloat(entry.cost) ? entry.cost : "0",
          });

          if (error) {
            throw new Error("failed");
          }
        })
      );
    } catch (err) {
      toast.error(t("maintenance.toast.failed_to_create"));
      return;
    }
    toast.success(t("maintenance.toast.successfully_created"));
    closeDialog(DialogID.EditMaintenance);
    emit("changed");
  }

  async function editEntry() {
    if (!entry.id) {
      return;
    }

    const { error } = await api.maintenance.update(entry.id, {
      name: entry.name,
      completedDate: entry.completedDate ?? "null",
      scheduledDate: entry.scheduledDate ?? "null",
      description: entry.description,
      cost: entry.cost,
    });

    if (error) {
      toast.error(t("maintenance.toast.failed_to_update"));
      return;
    }

    closeDialog(DialogID.EditMaintenance);
    emit("changed");
  }

  const openCreateModal = (itemIds: string[]) => {
    entry.id = null;
    entry.name = "";
    entry.completedDate = null;
    entry.scheduledDate = null;
    entry.description = "";
    entry.cost = "";
    entry.itemIds = itemIds;
    openDialog(DialogID.EditMaintenance);
  };

  const openUpdateModal = (maintenanceEntry: MaintenanceEntry | MaintenanceEntryWithDetails) => {
    entry.id = maintenanceEntry.id;
    entry.name = maintenanceEntry.name;
    entry.completedDate = new Date(maintenanceEntry.completedDate);
    entry.scheduledDate = new Date(maintenanceEntry.scheduledDate);
    entry.description = maintenanceEntry.description;
    entry.cost = maintenanceEntry.cost;
    entry.itemIds = null;
    openDialog(DialogID.EditMaintenance);
  };

  const confirm = useConfirm();

  async function deleteEntry(id: string) {
    const result = await confirm.open(t("maintenance.modal.delete_confirmation"));
    if (result.isCanceled) {
      return;
    }

    const { error } = await api.maintenance.delete(id);

    if (error) {
      toast.error(t("maintenance.toast.failed_to_delete"));
      return;
    }
    emit("changed");
  }

  async function complete(maintenanceEntry: MaintenanceEntry) {
    const { error } = await api.maintenance.update(maintenanceEntry.id, {
      name: maintenanceEntry.name,
      completedDate: new Date(Date.now()),
      scheduledDate: maintenanceEntry.scheduledDate ?? "null",
      description: maintenanceEntry.description,
      cost: maintenanceEntry.cost,
    });
    if (error) {
      toast.error(t("maintenance.toast.failed_to_update"));
    }
    emit("changed");
  }

  function duplicate(maintenanceEntry: MaintenanceEntry | MaintenanceEntryWithDetails, itemId: string) {
    entry.id = null;
    entry.name = maintenanceEntry.name;
    entry.completedDate = null;
    entry.scheduledDate = null;
    entry.description = maintenanceEntry.description;
    entry.cost = maintenanceEntry.cost;
    entry.itemIds = [itemId];
    openDialog(DialogID.EditMaintenance);
  }

  defineExpose({ openCreateModal, openUpdateModal, deleteEntry, complete, duplicate });
</script>
