<template>
  <Dialog :dialog-id="DialogID.EditMaintenance">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          {{ entry.id ? $t("maintenance.modal.edit_title") : $t("maintenance.modal.new_title") }}
        </DialogTitle>
      </DialogHeader>

      <form class="flex flex-col gap-2" @submit.prevent="dispatchFormSubmit">
        <FormTextField v-model="entry.name" autofocus :tag="$t('maintenance.modal.entry_name')" />
        <DatePicker v-model="entry.completedDate" :tag="$t('maintenance.modal.completed_date')" />
        <DatePicker v-model="entry.scheduledDate" :tag="$t('maintenance.modal.scheduled_date')" />
        <FormTextArea v-model="entry.description" :tag="$t('maintenance.modal.notes')" />
        <FormTextField v-model="entry.cost" autofocus :tag="$t('maintenance.modal.cost')" />

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
  import MdiPost from "~icons/mdi/post";
  import DatePicker from "~~/components/Form/DatePicker.vue";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import Button from "@/components/ui/button/Button.vue";

  const { closeDialog, registerOpenDialogCallback } = useDialog();

  const { t } = useI18n();
  const api = useUserApi();

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
    if (!entry.itemIds) {
      return;
    }

    await Promise.allSettled(
      entry.itemIds.map(async itemId => {
        const { error } = await api.items.maintenance.create(itemId, {
          name: entry.name,
          completedDate: entry.completedDate ?? "",
          scheduledDate: entry.scheduledDate ?? "",
          description: entry.description,
          cost: parseFloat(entry.cost) ? entry.cost : "0",
        });

        if (error) {
          toast.error(t("maintenance.toast.failed_to_create"));
          return;
        }
      })
    );

    closeDialog(DialogID.EditMaintenance, true);
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

    closeDialog(DialogID.EditMaintenance, true);
  }

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.EditMaintenance, params => {
      switch (params.type) {
        case "create":
          entry.id = null;
          entry.name = "";
          entry.completedDate = null;
          entry.scheduledDate = null;
          entry.description = "";
          entry.cost = "";
          entry.itemIds = typeof params.itemId === "string" ? [params.itemId] : params.itemId;
          break;
        case "update":
          entry.id = params.maintenanceEntry.id;
          entry.name = params.maintenanceEntry.name;
          entry.completedDate = new Date(params.maintenanceEntry.completedDate);
          entry.scheduledDate = new Date(params.maintenanceEntry.scheduledDate);
          entry.description = params.maintenanceEntry.description;
          entry.cost = params.maintenanceEntry.cost;
          entry.itemIds = null;
          break;
        case "duplicate":
          entry.id = null;
          entry.name = params.maintenanceEntry.name;
          entry.completedDate = null;
          entry.scheduledDate = null;
          entry.description = params.maintenanceEntry.description;
          entry.cost = params.maintenanceEntry.cost;
          entry.itemIds = [params.itemId];
          break;
      }
    });

    onUnmounted(cleanup);
  });
</script>
