<script setup lang="ts">
  import type { LocationSummary, LocationUpdate } from "~~/lib/api/types/data-contracts";
  import { useLocationStore } from "~~/stores/locations";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";

  definePageMeta({
    middleware: ["auth"],
  });

  const route = useRoute();
  const api = useUserApi();
  const toast = useNotifier();

  const locationId = computed<string>(() => route.params.id as string);

  const { data: location } = useAsyncData(locationId.value, async () => {
    const { data, error } = await api.locations.get(locationId.value);
    if (error) {
      toast.error("Failed to load location");
      navigateTo("/home");
      return;
    }

    if (data.parent) {
      parent.value = locations.value.find(l => l.id === data.parent.id);
    }

    return data;
  });

  const confirm = useConfirm();

  async function confirmDelete() {
    const { isCanceled } = await confirm.open(
      "Are you sure you want to delete this location and all of its items? This action cannot be undone."
    );
    if (isCanceled) {
      return;
    }

    const { error } = await api.locations.delete(locationId.value);
    if (error) {
      toast.error("Failed to delete location");
      return;
    }

    toast.success("Location deleted");
    navigateTo("/home");
  }

  const updateModal = ref(false);
  const updating = ref(false);
  const updateData = reactive<LocationUpdate>({
    id: locationId.value,
    name: "",
    description: "",
    parentId: null,
  });

  function openUpdate() {
    updateData.name = location.value?.name || "";
    updateData.description = location.value?.description || "";
    updateModal.value = true;
  }

  async function update() {
    updating.value = true;
    updateData.parentId = parent.value?.id || null;
    const { error, data } = await api.locations.update(locationId.value, updateData);

    if (error) {
      updating.value = false;
      toast.error("Failed to update location");
      return;
    }

    toast.success("Location updated");
    location.value = data;
    updateModal.value = false;
    updating.value = false;
  }

  const locationStore = useLocationStore();
  const locations = computed(() => locationStore.allLocations);

  const parent = ref<LocationSummary | any>({});

  const items = computedAsync(async () => {
    if (!location.value) {
      return [];
    }

    const resp = await api.items.getAll({
      locations: [location.value.id],
    });

    if (resp.error) {
      toast.error("Failed to load items");
      return [];
    }

    return resp.data.items;
  });
</script>

<template>
  <div>
    <!-- Update Dialog -->
    <BaseModal v-model="updateModal">
      <template #title> Update Location </template>
      <form v-if="location" @submit.prevent="update">
        <FormTextField v-model="updateData.name" :autofocus="true" label="Location Name" />
        <FormTextArea v-model="updateData.description" label="Location Description" />
        <LocationSelector v-model="parent" />
        <div class="modal-action">
          <BaseButton type="submit" :loading="updating"> Update </BaseButton>
        </div>
      </form>
    </BaseModal>

    <BaseContainer v-if="location">
      <div class="bg-base-100 rounded p-3">
        <header class="mb-2">
          <div class="flex flex-wrap items-end gap-2">
            <div class="avatar placeholder mb-auto">
              <div class="bg-neutral-focus text-neutral-content rounded-full w-12">
                <MdiPackageVariant name="mdi-package-variant" class="h-7 w-7" />
              </div>
            </div>
            <div>
              <div v-if="location?.parent" class="text-sm breadcrumbs pt-0 pb-0">
                <ul class="text-base-content/70">
                  <li>
                    <NuxtLink :to="`/location/${location.parent.id}`"> {{ location.parent.name }}</NuxtLink>
                  </li>
                  <li>{{ location.name }}</li>
                </ul>
              </div>
              <h1 class="text-2xl pb-1 flex items-center gap-3">
                {{ location ? location.name : "" }}

                <div
                  v-if="location && location.totalPrice"
                  class="text-xs bg-secondary text-secondary-content rounded-full px-2 py-1"
                >
                  <div>
                    <Currency :amount="location.totalPrice" />
                  </div>
                </div>
              </h1>
              <div class="flex gap-1 flex-wrap text-xs">
                <div>
                  Created
                  <DateTime :date="location?.createdAt" />
                </div>
              </div>
            </div>
            <div class="ml-auto mt-2 flex flex-wrap items-center justify-between gap-3">
              <div class="btn-group">
                <PageQRCode class="dropdown-left" />
                <BaseButton size="sm" @click="openUpdate">
                  <MdiPencil class="mr-1" name="mdi-pencil" />
                  Edit
                </BaseButton>
              </div>
              <BaseButton class="btn btn-sm" @click="confirmDelete()">
                <MdiDelete name="mdi-delete" class="mr-2" />
                Delete
              </BaseButton>
            </div>
          </div>
        </header>
        <div class="divider my-0 mb-1"></div>
        <Markdown v-if="location && location.description" class="text-base" :source="location.description"> </Markdown>
      </div>
      <section v-if="location && items">
        <ItemViewSelectable :items="items" />
      </section>

      <section v-if="location && location.children.length > 0" class="mt-6">
        <BaseSectionHeader class="mb-5"> Child Locations </BaseSectionHeader>
        <div class="grid gap-2 grid-cols-1 sm:grid-cols-3">
          <LocationCard v-for="item in location.children" :key="item.id" :location="item" />
        </div>
      </section>
    </BaseContainer>
  </div>
</template>
