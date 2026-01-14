<template>
  <BaseModal :dialog-id="DialogID.CreateCollection" :title="$t('components.collection.create_modal.title')">
    <form class="flex min-w-0 flex-col gap-2" @submit.prevent="create()">
      <FormTextField
        v-model="form.name"
        :trigger-focus="focused"
        :autofocus="true"
        :required="true"
        :label="$t('components.collection.create_modal.name_label')"
        :max-length="255"
        :min-length="1"
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
  import { useI18n } from "vue-i18n";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { toast } from "@/components/ui/sonner";
  import BaseModal from "@/components/App/CreateModal.vue";
  import { useDialog } from "~/components/ui/dialog-provider";
  import FormTextField from "~/components/Form/TextField.vue";
  import { Button, ButtonGroup } from "~/components/ui/button";
  import { useUserApi } from "~/composables/use-api";
  import { useCollections } from "~/composables/use-collections";

  const { t } = useI18n();

  const { activeDialog, closeDialog, registerOpenDialogCallback } = useDialog();

  const loading = ref(false);
  const focused = ref(false);

  const form = reactive({ name: "" });

  const api = useUserApi();
  const collections = useCollections();
  const { shift } = useMagicKeys();

  const redirectTo = ref("");

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.CreateCollection, params => {
      redirectTo.value = params?.redirectTo || "";
    });

    onUnmounted(cleanup);
  });

  watch(
    () => activeDialog.value,
    active => {
      if (active && active === DialogID.CreateCollection) {
        // reset and focus on open
        form.name = "";
        focused.value = true;
      }
    }
  );

  async function create(close = true) {
    if (loading.value) {
      toast.error(t("components.collection.create_modal.toast.already_creating"));
      return;
    }

    if (!form.name || form.name.trim().length === 0) {
      toast.error(t("components.collection.create_modal.toast.please_enter_name"));
      return;
    }

    loading.value = true;

    if (shift?.value) close = false;

    const { data, error } = await api.group.create(form.name.trim());

    if (error) {
      loading.value = false;
      toast.error(t("components.collection.create_modal.toast.create_failed"));
      return;
    }

    if (data) {
      toast.success(t("components.collection.create_modal.toast.create_success"));
    }

    form.name = "";
    loading.value = false;

    if (close) {
      closeDialog(DialogID.CreateCollection);
      if (data) {
        const createdId = data.id;
        collections.set(createdId);
        if (redirectTo.value === "") {
          // reload page to reflect new collection
          window.location.reload();
        } else {
          window.location.href = redirectTo.value;
        }
      }
    } else {
      // refresh global collections list
      await collections.load();
    }
  }
</script>
