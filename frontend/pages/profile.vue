<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import type { Detail } from "~~/components/global/DetailsSection/types";
  import MdiLoading from "~icons/mdi/loading";
  import MdiAccount from "~icons/mdi/account";
  import MdiDelete from "~icons/mdi/delete";
  import MdiFill from "~icons/mdi/fill";
  import { Button } from "@/components/ui/button";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import LanguageSelector from "~/components/App/LanguageSelector.vue";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import ThemePicker from "~/components/App/ThemePicker.vue";
  import ItemDuplicateSettings from "~/components/Item/DuplicateSettings.vue";
  import FormPassword from "~/components/Form/Password.vue";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseCard from "@/components/Base/Card.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import DetailsSection from "@/components/global/DetailsSection/DetailsSection.vue";
  import PasswordScore from "~/components/global/PasswordScore.vue";

  const { t } = useI18n();

  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "HomeBox | " + t("menu.profile"),
  });

  const api = useUserApi();
  const confirm = useConfirm();

  const { openDialog, closeDialog } = useDialog();

  const preferences = useViewPreferences();
  function setDisplayHeader() {
    preferences.value.displayLegacyHeader = !preferences.value.displayLegacyHeader;
  }
  function setLegacyImageFit() {
    preferences.value.legacyImageFit = !preferences.value.legacyImageFit;
  }

  const auth = useAuthContext();

  const details = computed(() => {
    return [
      {
        name: "global.name",
        text: auth.user?.name || t("global.unknown"),
      },
      {
        name: "global.email",
        text: auth.user?.email || t("global.unknown"),
      },
    ] as Detail[];
  });

  async function deleteProfile() {
    const result = await confirm.open(t("profile.delete_account_confirm"));

    if (result.isCanceled) {
      return;
    }

    const { response } = await api.user.delete();

    if (response?.status === 204) {
      toast.success(t("profile.toast.account_deleted"));
      auth.logout(api);
      navigateTo("/");
    }

    toast.error(t("profile.toast.failed_delete_account"));
  }

  const passwordChange = reactive({
    loading: false,
    current: "",
    new: "",
    isValid: false,
  });

  async function changePassword() {
    passwordChange.loading = true;
    if (!passwordChange.isValid) {
      return;
    }

    const { error } = await api.user.changePassword(passwordChange.current, passwordChange.new);

    if (error) {
      toast.error(t("profile.toast.failed_change_password"));
      passwordChange.loading = false;
      return;
    }

    toast.success(t("profile.toast.password_changed"));
    closeDialog(DialogID.ChangePassword);
    passwordChange.new = "";
    passwordChange.current = "";
    passwordChange.loading = false;
  }
</script>

<template>
  <div>
    <Dialog :dialog-id="DialogID.DuplicateSettings">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("items.duplicate.title") }}</DialogTitle>
        </DialogHeader>
        <ItemDuplicateSettings v-model="preferences.duplicateSettings" />
        <p class="text-sm text-muted-foreground">
          {{ $t("items.duplicate.override_instructions") }}
        </p>
      </DialogContent>
    </Dialog>

    <Dialog :dialog-id="DialogID.ChangePassword">
      <DialogContent>
        <DialogHeader>
          <DialogTitle> {{ $t("profile.change_password") }} </DialogTitle>
        </DialogHeader>

        <FormPassword
          v-model="passwordChange.current"
          :label="$t('profile.current_password')"
          placeholder=""
          class="mb-2"
        />
        <FormPassword v-model="passwordChange.new" :label="$t('profile.new_password')" placeholder="" />
        <PasswordScore v-model:valid="passwordChange.isValid" :password="passwordChange.new" />

        <form @submit.prevent="changePassword">
          <DialogFooter>
            <Button :disabled="!passwordChange.isValid || passwordChange.loading" type="submit">
              <MdiLoading v-if="passwordChange.loading" class="animate-spin" />
              {{ $t("global.submit") }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <BaseContainer class="flex flex-col gap-4">
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiAccount class="-mt-1 mr-2" />
            <span> {{ $t("profile.user_profile") }} </span>
            <template #description> {{ $t("profile.user_profile_sub") }} </template>
          </BaseSectionHeader>
        </template>

        <DetailsSection :details="details" />

        <div class="p-4">
          <div class="flex gap-2">
            <Button variant="secondary" size="sm" @click="openDialog(DialogID.ChangePassword)">
              {{ $t("profile.change_password") }}
            </Button>
            <Button variant="secondary" size="sm" @click="openDialog(DialogID.DuplicateSettings)">
              {{ $t("items.duplicate.title") }}
            </Button>
          </div>
        </div>
        <LanguageSelector />
      </BaseCard>

      <!-- TODO: Remove this notice once users are familiar with the collection-based settings. -->
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <span> {{ $t("profile.moved_notice_title") }} </span>
            <template #description>
              {{ $t("profile.moved_notice_description") }}
            </template>
          </BaseSectionHeader>
        </template>

        <div class="space-y-2 px-4 pb-4 text-sm text-muted-foreground">
          <p>
            {{ $t("profile.moved_notice_body") }}
          </p>
          <div class="flex flex-wrap gap-2">
            <NuxtLink to="/collection/settings" class="text-primary underline">
              {{ $t("profile.moved_notice_link_settings") }}
            </NuxtLink>
            <NuxtLink to="/collection/notifiers" class="text-primary underline">
              {{ $t("profile.moved_notice_link_notifiers") }}
            </NuxtLink>
            <NuxtLink to="/collection/invites" class="text-primary underline">
              {{ $t("profile.moved_notice_link_invites") }}
            </NuxtLink>
          </div>
        </div>
      </BaseCard>
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiFill class="mr-2" />
            <span> {{ $t("profile.theme_settings") }} </span>
            <template #description>
              {{ $t("profile.theme_settings_sub") }}
            </template>
          </BaseSectionHeader>
        </template>

        <div class="px-4 pb-4">
          <div class="mb-3 flex gap-2">
            <Button variant="secondary" size="sm" @click="setDisplayHeader">
              {{ $t("profile.display_legacy_header", { currentValue: preferences.displayLegacyHeader }) }}
            </Button>
            <Button variant="secondary" size="sm" @click="setLegacyImageFit">
              {{ $t("profile.legacy_image_fit", { currentValue: preferences.legacyImageFit }) }}
            </Button>
          </div>
          <ThemePicker />
        </div>
      </BaseCard>

      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiDelete class="-mt-1 mr-2" />
            <span> {{ $t("profile.delete_account") }} </span>
            <template #description> {{ $t("profile.delete_account_sub") }} </template>
          </BaseSectionHeader>
        </template>
        <div class="border-t-2 p-4 px-6">
          <Button size="sm" variant="destructive" @click="deleteProfile">
            {{ $t("profile.delete_account") }}
          </Button>
        </div>
      </BaseCard>
    </BaseContainer>
  </div>
</template>

<style scoped></style>
