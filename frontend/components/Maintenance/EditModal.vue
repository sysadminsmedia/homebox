<template>
  <BaseModal v-model="visible">
    <template #title>
      {{ entry.id ? "Edit Entry" : "New Entry" }}
    </template>
    <form @submit.prevent="dispatchFormSubmit">
      <FormTextField v-model="entry.name" autofocus label="Entry Name" />
      <DatePicker v-model="entry.completedDate" label="Completed Date" />
      <DatePicker v-model="entry.scheduledDate" label="Scheduled Date" />
      <FormTextArea v-model="entry.description" label="Notes" />
      <FormTextField v-model="entry.cost" autofocus label="Cost" />
      <div class="flex justify-end py-2">
        <BaseButton type="submit" class="ml-2 mt-2">
          <template #icon>
            <MdiPost />
          </template>
          {{ entry.id ? "Update" : "Create" }}
        </BaseButton>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import type { MaintenanceEntry, MaintenanceEntryWithDetails } from "~~/lib/api/types/data-contracts";
  import MdiPost from "~icons/mdi/post";
  import DatePicker from "~~/components/Form/DatePicker.vue";

  const api = useUserApi();
  const toast = useNotifier();

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
      toast.error("Failed to create entry");
      return;
    }

    visible.value = false;
    emit("changed");
  }

  async function editEntry() {
    if (!entry.id) {
      return;
    }

    const { error } = await api.maintenances.update(entry.id, {
      name: entry.name,
      completedDate: entry.completedDate ?? "null",
      scheduledDate: entry.scheduledDate ?? "null",
      description: entry.description,
      cost: entry.cost,
    });

    if (error) {
      toast.error("Failed to update entry");
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
  }

  const openUpdateModal = (maintenanceEntry: MaintenanceEntry | MaintenanceEntryWithDetails) => {
    entry.id = maintenanceEntry.id;
    entry.name = maintenanceEntry.name;
    entry.completedDate = new Date(maintenanceEntry.completedDate);
    entry.scheduledDate = new Date(maintenanceEntry.scheduledDate);
    entry.description = maintenanceEntry.description;
    entry.cost = maintenanceEntry.cost;
    entry.itemId = null;
    visible.value = true;
  }

  defineExpose({openCreateModal, openUpdateModal});
</script>
