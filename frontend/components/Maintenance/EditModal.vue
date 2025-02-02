<template>
  <BaseModal v-model="visible">
    <template #title>
      {{ entry.id ? $t("maintenance.modal.edit_title") : $t("maintenance.modal.new_title") }}
    </template>
    <form @submit.prevent="dispatchFormSubmit">
      <FormTextField v-model="entry.name" autofocus :label="$t('maintenance.modal.entry_name')" />
      <DatePicker v-model="entry.completedDate" :label="$t('maintenance.modal.completed_date')" />
      <DatePicker v-model="entry.scheduledDate" :label="$t('maintenance.modal.scheduled_date')" />
      <FormTextArea v-model="entry.description" :label="$t('maintenance.modal.notes')" />
      <FormTextField v-model="entry.cost" autofocus :label="$t('maintenance.modal.cost')" />
      <div class="flex justify-end py-2">
        <BaseButton type="submit" class="ml-2 mt-2">
          <template #icon>
            <MdiPost />
          </template>
          {{ entry.id ? $t("maintenance.modal.edit_action") : $t("maintenance.modal.new_action") }}
        </BaseButton>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import { toast } from "vue-sonner";
  import { useI18n } from "vue-i18n";
  import type { MaintenanceEntry, MaintenanceEntryWithDetails } from "~~/lib/api/types/data-contracts";
  import MdiPost from "~icons/mdi/post";
  import DatePicker from "~~/components/Form/DatePicker.vue";

  const { t } = useI18n();
  const api = useUserApi();

  const emit = defineEmits(["changed"]);

  const visible = ref(false);
  const entry = reactive({
    id: null as string | null,
    name: "",
    completedDate: null as Date | null,
    scheduledDate: null as Date | null,
    description: "",
    cost: "",
    itemId: null as string | null,
  });

  async function dispatchFormSubmit() {
    if (entry.id) {
      await editEntry();
      return;
    }

    await createEntry();
  }

  async function createEntry() {
    if (!entry.itemId) {
      return;
    }
    const { error } = await api.items.maintenance.create(entry.itemId, {
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

    visible.value = false;
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

    visible.value = false;
    emit("changed");
  }

  const openCreateModal = (itemId: string) => {
    entry.id = null;
    entry.name = "";
    entry.completedDate = null;
    entry.scheduledDate = null;
    entry.description = "";
    entry.cost = "";
    entry.itemId = itemId;
    visible.value = true;
  };

  const openUpdateModal = (maintenanceEntry: MaintenanceEntry | MaintenanceEntryWithDetails) => {
    entry.id = maintenanceEntry.id;
    entry.name = maintenanceEntry.name;
    entry.completedDate = new Date(maintenanceEntry.completedDate);
    entry.scheduledDate = new Date(maintenanceEntry.scheduledDate);
    entry.description = maintenanceEntry.description;
    entry.cost = maintenanceEntry.cost;
    entry.itemId = null;
    visible.value = true;
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
    entry.itemId = itemId;
    visible.value = true;
  }

  defineExpose({ openCreateModal, openUpdateModal, deleteEntry, complete, duplicate });
</script>
