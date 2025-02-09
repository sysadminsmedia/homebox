<template>
  <Dialog dialog-id="create-label">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ $t("components.label.create_modal.title") }}</DialogTitle>
      </DialogHeader>

      <form @submit.prevent="create()">
        <FormTextField
          ref="locationNameRef"
          v-model="form.name"
          :trigger-focus="focused"
          :autofocus="true"
          :label="$t('components.label.create_modal.label_name')"
          :max-length="255"
          :min-length="1"
        />
        <FormTextArea
          v-model="form.description"
          :label="$t('components.label.create_modal.label_description')"
          :max-length="255"
        />
                <div class="mt-4 flex flex-row-reverse">
          <ButtonGroup>
            <Button  :disabled="loading" type="submit">{{ $t("global.create") }}</Button>
            <Button variant="outline" :disabled="loading" type="button" @click="create(false)">{{ $t("global.create_and_add") }}</Button>
          </ButtonGroup>
        </div>
      </form>

      <DialogFooter>
        <span class="flex items-center gap-1 text-sm">
          Use <Shortcut size="sm" :keys="['Shift']" /> + <Shortcut size="sm" :keys="['Enter']" /> to create and add
          another.
        </span>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { toast } from "vue-sonner";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import MdiChevronDown from "~icons/mdi/chevron-down";
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

  // watch(
  //   () => modal.value,
  //   open => {
  //     if (open)
  //       useTimeoutFn(() => {
  //         focused.value = true;
  //       }, 50);
  //     else focused.value = false;
  //   }
  // );

  const api = useUserApi();

  const { shift } = useMagicKeys();

  async function create(close = true) {
    if (loading.value) {
      toast.error("Already creating a label");
      return;
    }
    loading.value = true;

    if (shift.value) {
      close = false;
    }

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
