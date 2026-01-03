<template>
  <BaseModal :dialog-id="DialogID.JoinCollection" :title="$t('components.collection.join_modal.title')" hide-footer>
    <form class="flex min-w-0 flex-col gap-2" @submit.prevent="join()">
      <FormTextField
        v-model="form.inviteCode"
        :trigger-focus="focused"
        :autofocus="true"
        :required="true"
        :label="$t('components.collection.join_modal.invite_code_label')"
      />

      <div class="mt-2 flex flex-col">
        <span class="text-sm text-muted-foreground">
          {{ $t("components.collection.join_modal.invites_should_look_like") }}
        </span>
        <div>
          <Badge variant="outline"> AYQ4W4K5MT4CZOPB2PZRCZ4PTY </Badge>
          <Badge variant="outline"> {{ `${domain}?token=AYQ4W4K5MT4CZOPB2PZRCZ4PTY` }} </Badge>
        </div>
      </div>

      <Button class="mt-2" variant="outline" @click="openDialog(DialogID.CreateCollection)">
        {{ $t("components.collection.join_modal.no_invite_code_create_one") }}
      </Button>

      <div class="mt-2 flex flex-row-reverse">
        <Button :disabled="loading" type="submit">{{ $t("components.collection.join_modal.join") }}</Button>
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
  import { Button } from "~/components/ui/button";
  import { Badge } from "~/components/ui/badge";

  const { t } = useI18n();

  const { activeDialog, closeDialog, openDialog } = useDialog();

  const loading = ref(false);
  const focused = ref(false);

  const form = reactive({ inviteCode: "" });

  const api = useUserApi();

  const collectionStore = useCollectionStore();

  const domain = window.location.protocol + "//" + window.location.host;

  watch(
    () => activeDialog.value,
    active => {
      if (active && active === DialogID.JoinCollection) {
        // reset and focus on open
        form.inviteCode = "";
        focused.value = true;
      }
    }
  );

  async function join() {
    if (loading.value) {
      toast.error(t("components.collection.join_modal.toast.already_joining"));
      return;
    }

    if (!form.inviteCode || form.inviteCode.trim().length === 0) {
      toast.error(t("components.collection.join_modal.toast.please_enter_valid_join_code"));
      return;
    }

    // remove everything before the first '?' character
    const code = form.inviteCode.includes("?")
      ? form.inviteCode.split("?")[1]?.replace("token=", "").trim()
      : form.inviteCode.trim();

    if (!code || code.length !== 26) {
      toast.error(t("components.collection.join_modal.toast.please_enter_valid_join_code"));
      return;
    }

    loading.value = true;

    const { data, error } = await api.group.acceptInvitation(code);

    if (error) {
      loading.value = false;
      toast.error(t("components.collection.join_modal.toast.join_failed"));
      return;
    }

    if (data) {
      toast.success(t("components.collection.join_modal.toast.join_success"));
    }

    form.inviteCode = "";
    loading.value = false;

    closeDialog(DialogID.CreateCollection);
    if (data) {
      const joinedId = data.id;

      collectionStore.set(joinedId);
      // reload page to reflect joined collection
      window.location.reload();
    }
  }
</script>
