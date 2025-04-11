<template>
  <Dialog dialog-id="import">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ $t("components.app.import_dialog.title") }}</DialogTitle>
        <DialogDescription> {{ $t("components.app.import_dialog.description") }} </DialogDescription>
      </DialogHeader>

      <div class="bg-destructive text-destructive-foreground flex gap-2 rounded p-2 shadow-lg">
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

      <form class="flex flex-col gap-4" @submit.prevent="submitCsvFile">
        <Input ref="importRef" type="file" accept=".csv,.tsv" @change="setFile" />

        <DialogFooter>
          <Button type="submit" :disabled="!importCsv"> {{ $t("global.submit") }} </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
  } from "@/components/ui/dialog";
  import { Button } from "@/components/ui/button";
  import { Input } from "@/components/ui/input";
  type Props = {
    modelValue: boolean;
  };

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
  });

  const emit = defineEmits(["update:modelValue"]);

  const dialog = useVModel(props, "modelValue", emit);

  const api = useUserApi();

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
