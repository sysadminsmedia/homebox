<template>
  <BaseModal v-model="dialog">
    <template #title> {{ $t("components.app.import_dialog.title") }} </template>
    <p>
      {{ $t("components.app.import_dialog.description") }}
    </p>
    <div class="alert alert-warning mt-4 shadow-lg">
      <div>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="mb-auto size-6 shrink-0 stroke-current"
          fill="none"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        <span class="text-sm">
          {{ $t("components.app.import_dialog.change_warning") }}
        </span>
      </div>
    </div>

    <form @submit.prevent="submitCsvFile">
      <div class="flex flex-col gap-2 py-6">
        <input ref="importRef" type="file" class="hidden" accept=".csv,.tsv" @change="setFile" />

        <BaseButton type="button" @click="uploadCsv">
          <MdiUpload class="mr-2 size-5" />
          {{ $t("components.app.import_dialog.upload") }}
        </BaseButton>
        <p class="-mb-5 pt-4 text-center">
          {{ importCsv?.name }}
        </p>
      </div>

      <div class="modal-action">
        <BaseButton type="submit" :disabled="!importCsv"> {{ $t("global.submit") }} </BaseButton>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import MdiUpload from "~icons/mdi/upload";
  type Props = {
    modelValue: boolean;
  };

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
  });

  const emit = defineEmits(["update:modelValue"]);

  const dialog = useVModel(props, "modelValue", emit);

  const api = useUserApi();
  const toast = useNotifier();

  const importCsv = ref<File | null>(null);
  const importLoading = ref(false);
  const importRef = ref<HTMLInputElement>();
  whenever(
    () => !dialog.value,
    () => {
      importCsv.value = null;
    }
  );

  function setFile(e: Event) {
    const result = e.target as HTMLInputElement;
    if (!result.files || result.files.length === 0) {
      return;
    }

    importCsv.value = result.files[0];
  }

  function uploadCsv() {
    importRef.value?.click();
  }

  async function submitCsvFile() {
    if (!importCsv.value) {
      toast.error("Please select a file to import.");
      return;
    }

    importLoading.value = true;

    const { error } = await api.items.import(importCsv.value);

    if (error) {
      toast.error("Import failed. Please try again later.");
    }

    // Reset
    dialog.value = false;
    importLoading.value = false;
    importCsv.value = null;

    if (importRef.value) {
      importRef.value.value = "";
    }

    toast.success("Import successful!");
  }
</script>
