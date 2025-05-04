<template>
  <BaseModal dialog-id="create-label" :title="$t('components.label.create_modal.title')">
    <form class="flex flex-col gap-2" @submit.prevent="create()">
      <FormTextField
        v-model="form.name"
        :trigger-focus="focused"
        :autofocus="true"
        :label="$t('components.label.create_modal.label_name')"
        :max-length="50"
        :min-length="1"
      />
      <FormTextArea
        v-model="form.description"
        :label="$t('components.label.create_modal.label_description')"
        :max-length="255"
      />
      <div class="mt-4 flex flex-row-reverse">
        <ButtonGroup>
          <Button :disabled="loading" type="submit">{{ $t("global.create") }}</Button>
          <Button variant="outline" :disabled="loading" type="button" @click="create(false)">
            {{ $t("global.create_and_add") }}
          </Button>
        </ButtonGroup>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import BaseModal from "@/components/App/CreateModal.vue";
  import { useDialog, useDialogHotkey } from "~/components/ui/dialog-provider";

  const { closeDialog } = useDialog();

  useDialogHotkey("create-label", { code: "Digit2", shift: true });

  const loading = ref(false);
  const focused = ref(false);
  const form = reactive({
    name: "",
    description: "",
    color: "", // Future!
  });

  function reset() {
    form.name = "";
    form.description = "";
    form.color = "";
    focused.value = false;
    loading.value = false;
  }

  const api = useUserApi();
  const { shift } = useMagicKeys();

  async function create(close = true) {
    if (loading.value) {
      toast.error("Already creating a label");
      return;
    }
    if (form.name.length > 50) {
      toast.error("Label name must not be longer than 50 characters");
      return;
    }

    loading.value = true;

    if (shift.value) close = false;

    const { error, data } = await api.labels.create(form);

    if (error) {
      toast.error("Couldn't create label");
      loading.value = false;
      return;
    }

    toast.success("Label created");
    reset();

    if (close) {
      closeDialog("create-label");
      navigateTo(`/label/${data.id}`);
    }
  }
</script>
