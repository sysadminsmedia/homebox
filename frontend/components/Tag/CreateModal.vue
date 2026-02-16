<template>
  <BaseModal :dialog-id="DialogID.CreateTag" :title="$t('components.tag.create_modal.title')">
    <form class="flex min-w-0 flex-col gap-2" @submit.prevent="create()">
      <FormTextField
        v-model="form.name"
        :trigger-focus="focused"
        :autofocus="true"
        :label="$t('components.tag.create_modal.tag_name')"
        :max-length="50"
        :min-length="1"
      />
      <FormTextArea
        v-model="form.description"
        :label="$t('components.tag.create_modal.tag_description')"
        :max-length="1000"
      />
      <TagSingleSelector
        v-model="form.parentTag"
        :tags="availableTags"
        :name="$t('components.tag.create_modal.tag_parent')"
      />
      <ColorSelector v-model="form.color" :label="$t('components.tag.create_modal.tag_color')" :show-hex="true" />
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
  import { useDialog, useDialogHotkey } from "~/components/ui/dialog-provider";
  import ColorSelector from "@/components/Form/ColorSelector.vue";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import { Button, ButtonGroup } from "~/components/ui/button";
  import TagSingleSelector from "~/components/Tag/SingleSelector.vue";
  import type { TagOut } from "~~/lib/api/types/data-contracts";
  import { useTagStore } from "~/stores/tags";

  const { t } = useI18n();

  const { closeDialog } = useDialog();
  const tagStore = useTagStore();

  useDialogHotkey(DialogID.CreateTag, { code: "Digit2", shift: true });

  const loading = ref(false);
  const focused = ref(false);
  const form = reactive({
    name: "",
    description: "",
    color: "",
    parentTag: null as TagOut | null,
  });

  onMounted(async () => {
    await tagStore.ensureAllTagsFetched();
  });

  const availableTags = computed(() => {
    return tagStore.tags;
  });

  function reset() {
    form.name = "";
    form.description = "";
    form.color = "";
    form.parentTag = null;
    focused.value = false;
    loading.value = false;
  }

  const api = useUserApi();
  const { shift } = useMagicKeys();

  async function create(close = true) {
    if (loading.value) {
      toast.error(t("components.tag.create_modal.toast.already_creating"));
      return;
    }
    if (form.name.length > 50) {
      toast.error(t("components.tag.create_modal.toast.tag_name_too_long"));
      return;
    }

    loading.value = true;

    if (shift?.value) close = false;

    const { error, data } = await api.tags.create({
      name: form.name,
      description: form.description,
      color: form.color,
      icon: "",
      parentId: form.parentTag?.id,
    });

    if (error) {
      toast.error(t("components.tag.create_modal.toast.create_failed"));
      loading.value = false;
      return;
    }

    toast.success(t("components.tag.create_modal.toast.create_success"));
    reset();

    if (close) {
      closeDialog(DialogID.CreateTag);
      await navigateTo(`/tag/${data.id}`);
    }
  }
</script>
