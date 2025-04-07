<script setup lang="ts">
  import { toast } from "vue-sonner";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";

  definePageMeta({
    middleware: ["auth"],
  });

  const route = useRoute();
  const api = useUserApi();

  const labelId = computed<string>(() => route.params.id as string);

  const { data: label } = useAsyncData(labelId.value, async () => {
    const { data, error } = await api.labels.get(labelId.value);
    if (error) {
      toast.error("Failed to load label");
      navigateTo("/home");
      return;
    }
    return data;
  });

  const confirm = useConfirm();

  async function confirmDelete() {
    const { isCanceled } = await confirm.open(
      "Are you sure you want to delete this label? This action cannot be undone."
    );

    if (isCanceled) {
      return;
    }

    const { error } = await api.labels.delete(labelId.value);

    if (error) {
      toast.error("Failed to delete label");
      return;
    }
    toast.success("Label deleted");
    navigateTo("/home");
  }

  const updateModal = ref(false);
  const updating = ref(false);
  const updateData = reactive({
    name: "",
    description: "",
    color: "",
  });

  function openUpdate() {
    updateData.name = label.value?.name || "";
    updateData.description = label.value?.description || "";
    updateModal.value = true;
  }

  async function update() {
    updating.value = true;
    const { error, data } = await api.labels.update(labelId.value, updateData);

    if (error) {
      updating.value = false;
      toast.error("Failed to update label");
      return;
    }

    toast.success("Label updated");
    label.value = data;
    updateModal.value = false;
    updating.value = false;
  }

  const items = computedAsync(async () => {
    if (!label.value) {
      return {
        items: [],
        totalPrice: null,
      };
    }

    const resp = await api.items.getAll({
      labels: [label.value.id],
    });

    if (resp.error) {
      toast.error("Failed to load items");
      return {
        items: [],
        totalPrice: null,
      };
    }

    return resp.data;
  });
</script>

<template>
  <BaseContainer>
    <BaseModal v-model="updateModal">
      <template #title> {{ $t("labels.update_label") }} </template>
      <form v-if="label" class="flex flex-col gap-2" @submit.prevent="update">
        <FormTextField
          v-model="updateData.name"
          :autofocus="true"
          :label="$t('components.label.create_modal.label_name')"
          :max-length="255"
          :min-length="1"
        />
        <FormTextArea
          v-model="updateData.description"
          :label="$t('components.label.create_modal.label_description')"
          :max-length="255"
        />
        <div class="modal-action">
          <BaseButton type="submit" :loading="updating"> {{ $t("global.update") }} </BaseButton>
        </div>
      </form>
    </BaseModal>

    <BaseContainer v-if="label">
      <div class="rounded bg-base-100 p-3">
        <header class="mb-2">
          <div class="flex flex-wrap items-end gap-2">
            <div class="avatar placeholder mb-auto">
              <div class="w-12 rounded-full bg-neutral-focus text-neutral-content">
                <MdiPackageVariant class="size-7" />
              </div>
            </div>
            <div>
              <h1 class="flex items-center gap-3 pb-1 text-2xl">
                {{ label ? label.name : "" }}

                <div
                  v-if="items && items.totalPrice"
                  class="rounded-full bg-secondary px-2 py-1 text-xs text-secondary-content"
                >
                  <div>
                    <Currency :amount="items.totalPrice" />
                  </div>
                </div>
              </h1>
              <div class="flex flex-wrap gap-1 text-xs">
                <div>
                  Created
                  <DateTime :date="label?.createdAt" />
                </div>
              </div>
            </div>
            <div class="ml-auto mt-2 flex flex-wrap items-center justify-between gap-3">
              <div class="btn-group">
                <PageQRCode class="dropdown-left" />
                <BaseButton size="sm" @click="openUpdate">
                  <MdiPencil class="mr-1" />
                  Edit
                </BaseButton>
              </div>
              <BaseButton class="btn btn-sm" @click="confirmDelete()">
                <MdiDelete class="mr-2" />
                Delete
              </BaseButton>
            </div>
          </div>
        </header>
        <div class="divider my-0 mb-1"></div>
        <Markdown v-if="label && label.description" class="text-base" :source="label.description"> </Markdown>
      </div>
      <section v-if="label && items">
        <ItemViewSelectable :items="items.items" />
      </section>
    </BaseContainer>
  </BaseContainer>
</template>
