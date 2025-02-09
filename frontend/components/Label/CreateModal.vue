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
        <div class="modal-action">
          <div class="flex justify-center">
            <BaseButton class="rounded-r-none" :loading="loading" type="submit"> {{ $t("global.create") }} </BaseButton>
            <div class="dropdown dropdown-top">
              <label tabindex="0" class="btn rounded-l-none rounded-r-xl">
                <MdiChevronDown class="size-5" />
              </label>
              <ul tabindex="0" class="dropdown-content menu rounded-box bg-base-100 right-0 w-64 p-2 shadow">
                <li>
                  <button type="button" @click="create(false)">{{ $t("global.create_and_add") }}</button>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </form>

      <DialogFooter>
        use <kbd class="kbd kbd-xs">Shift</kbd> + <kbd class="kbd kbd-xs"> Enter </kbd> to create and add another
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
