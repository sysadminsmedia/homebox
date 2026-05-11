<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import type { Detail } from "~~/components/global/DetailsSection/types";
  import MdiLoading from "~icons/mdi/loading";
  import MdiAccount from "~icons/mdi/account";
  import MdiDelete from "~icons/mdi/delete";
  import MdiFill from "~icons/mdi/fill";
  import MdiKeyVariant from "~icons/mdi/key-variant";
  import MdiContentCopy from "~icons/mdi/content-copy";
  import { Button } from "@/components/ui/button";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import LanguageSelector from "~/components/App/LanguageSelector.vue";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import ThemePicker from "~/components/App/ThemePicker.vue";
  import ItemDuplicateSettings from "~/components/Item/DuplicateSettings.vue";
  import FormPassword from "~/components/Form/Password.vue";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormCheckbox from "~/components/Form/Checkbox.vue";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseCard from "@/components/Base/Card.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import DetailsSection from "@/components/global/DetailsSection/DetailsSection.vue";
  import DateTime from "@/components/global/DateTime.vue";
  import PasswordScore from "~/components/global/PasswordScore.vue";
  import { PASSWORD_MIN_LENGTH, PASSWORD_RULES } from "~/lib/passwords";
  import type { APIKeyOut } from "~~/lib/api/types/data-contracts";

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
      passwordChange.loading = false;
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

  // ---------------------------------------------------------------------------
  // API keys

  const apiKeys = ref<APIKeyOut[]>([]);
  const apiKeysLoading = ref(false);

  async function loadApiKeys() {
    apiKeysLoading.value = true;
    const { data, error } = await api.user.listApiKeys();
    apiKeysLoading.value = false;
    if (error) {
      toast.error(t("errors.api_failure") + String(error));
      return;
    }
    apiKeys.value = data ?? [];
  }

  onMounted(() => {
    void loadApiKeys();
  });

  const apiKeyForm = reactive({
    name: "",
    setExpiration: false,
    expiresAt: "",
    submitting: false,
  });

  function openCreateApiKey() {
    apiKeyForm.name = "";
    apiKeyForm.setExpiration = false;
    apiKeyForm.expiresAt = "";
    apiKeyForm.submitting = false;
    openDialog(DialogID.CreateApiKey);
  }

  const newApiKeyToken = ref<string | null>(null);

  async function submitCreateApiKey() {
    if (!apiKeyForm.name.trim()) return;
    apiKeyForm.submitting = true;

    let expiresAt: string | null = null;
    if (apiKeyForm.setExpiration && apiKeyForm.expiresAt) {
      // <input type="date"> gives YYYY-MM-DD; treat as end-of-day UTC so the
      // key remains valid through the chosen calendar day.
      expiresAt = new Date(apiKeyForm.expiresAt + "T23:59:59Z").toISOString();
    }

    const { data, error } = await api.user.createApiKey({
      name: apiKeyForm.name.trim(),
      expiresAt,
    });

    apiKeyForm.submitting = false;

    if (error || !data) {
      toast.error(t("profile.toast.failed_api_key_create"));
      return;
    }

    closeDialog(DialogID.CreateApiKey);
    newApiKeyToken.value = data.token;
    openDialog(DialogID.CreateApiKeyResult);
    toast.success(t("profile.toast.api_key_created"));
    await loadApiKeys();
  }

  async function copyApiKey() {
    if (!newApiKeyToken.value) return;
    try {
      await navigator.clipboard.writeText(newApiKeyToken.value);
      toast.success(t("profile.api_key_copied"));
    } catch {
      // Clipboard API can be blocked by the browser; surface no error toast,
      // the user can still copy manually from the visible field.
    }
  }

  function dismissApiKeyResult() {
    closeDialog(DialogID.CreateApiKeyResult);
    newApiKeyToken.value = null;
  }

  async function revokeApiKey(key: APIKeyOut) {
    const result = await confirm.open(t("profile.api_key_delete_confirm"));
    if (result.isCanceled) return;

    const { error } = await api.user.deleteApiKey(key.id);
    if (error) {
      toast.error(t("profile.toast.failed_api_key_delete"));
      return;
    }
    toast.success(t("profile.toast.api_key_deleted"));
    await loadApiKeys();
  }

  // ---------------------------------------------------------------------------
  // Integration settings

  const integrationSettings = reactive({
    paperlessUrl: "",
    paperlessToken: "",
    immichUrl: "",
    immichToken: "",
    loading: false,
    saving: false,
  });

  async function loadIntegrationSettings() {
    integrationSettings.loading = true;
    const { data, error } = await api.user.getSettings();
    integrationSettings.loading = false;

    if (error || !data?.item) {
      toast.error(t("errors.api_failure"));
      return;
    }

    const settings = data.item as Record<string, unknown>;
    integrationSettings.paperlessUrl = (settings.paperless_url as string) || "";
    integrationSettings.paperlessToken = (settings.paperless_token as string) || "";
    integrationSettings.immichUrl = (settings.immich_url as string) || "";
    integrationSettings.immichToken = (settings.immich_token as string) || "";
  }

  async function saveIntegrationSettings() {
    integrationSettings.saving = true;
    const { data: current, error: currentError } = await api.user.getSettings();
    if (currentError) {
      integrationSettings.saving = false;
      toast.error(t("profile.toast.failed_settings_save"));
      return;
    }

    const { error } = await api.user.setSettings({
      ...(current?.item ?? {}),
      paperless_url: integrationSettings.paperlessUrl,
      paperless_token: integrationSettings.paperlessToken,
      immich_url: integrationSettings.immichUrl,
      immich_token: integrationSettings.immichToken,
    });
    integrationSettings.saving = false;

    if (error) {
      toast.error(t("profile.toast.failed_settings_save"));
      return;
    }

    toast.success(t("profile.toast.settings_saved"));
  }

  onMounted(() => {
    void loadIntegrationSettings();
  });
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

    <Dialog :dialog-id="DialogID.CreateApiKey">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("profile.api_key_create") }}</DialogTitle>
        </DialogHeader>

        <form @submit.prevent="submitCreateApiKey">
          <FormTextField
            v-model="apiKeyForm.name"
            :label="$t('profile.api_key_name')"
            :placeholder="$t('profile.api_key_name_placeholder')"
            :required="true"
            class="mb-2"
          />
          <div class="mb-2 max-w-[260px]">
            <FormCheckbox v-model="apiKeyForm.setExpiration" :label="$t('profile.api_key_set_expiration')" />
          </div>
          <div v-if="apiKeyForm.setExpiration" class="mb-2">
            <label class="mb-1 block text-sm font-medium">{{ $t("profile.api_key_expires_at") }}</label>
            <input
              v-model="apiKeyForm.expiresAt"
              type="date"
              class="w-full rounded-md border bg-background px-3 py-2 text-sm"
              :required="apiKeyForm.setExpiration"
            />
          </div>

          <DialogFooter>
            <Button :disabled="apiKeyForm.submitting || !apiKeyForm.name.trim()" type="submit">
              <MdiLoading v-if="apiKeyForm.submitting" class="animate-spin" />
              {{ $t("global.create") }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog :dialog-id="DialogID.CreateApiKeyResult">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("profile.api_key_token_label") }}</DialogTitle>
        </DialogHeader>

        <p class="text-sm text-muted-foreground">{{ $t("profile.api_key_token_warning") }}</p>

        <div class="flex items-center gap-2">
          <input
            :value="newApiKeyToken ?? ''"
            readonly
            class="w-full rounded-md border bg-muted px-3 py-2 font-mono text-sm"
            @focus="($event.target as HTMLInputElement).select()"
          />
          <Button variant="secondary" size="icon" :aria-label="$t('profile.api_key_copy')" @click="copyApiKey">
            <MdiContentCopy />
          </Button>
        </div>

        <DialogFooter>
          <Button @click="dismissApiKeyResult">{{ $t("global.close") }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :dialog-id="DialogID.ChangePassword">
      <DialogContent>
        <DialogHeader>
          <DialogTitle> {{ $t("profile.change_password") }} </DialogTitle>
        </DialogHeader>

        <form id="change-password-form" name="change-password" method="post" @submit.prevent="changePassword">
          <FormPassword
            id="current-password"
            v-model="passwordChange.current"
            :label="$t('profile.current_password')"
            name="current-password"
            autocomplete="current-password"
            placeholder=""
            :required="true"
            class="mb-2"
          />
          <FormPassword
            id="new-password"
            v-model="passwordChange.new"
            :label="$t('profile.new_password')"
            name="new-password"
            autocomplete="new-password"
            placeholder=""
            :min-length="PASSWORD_MIN_LENGTH"
            :passwordrules="PASSWORD_RULES"
            :required="true"
          />
          <PasswordScore v-model:valid="passwordChange.isValid" :password="passwordChange.new" />

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

      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiKeyVariant class="-mt-1 mr-2" />
            <span>{{ $t("profile.api_keys") }}</span>
            <template #description>{{ $t("profile.api_keys_sub") }}</template>
          </BaseSectionHeader>
        </template>

        <div class="px-4 pb-4">
          <div class="mx-1 divide-y rounded-md border">
            <p v-if="!apiKeysLoading && apiKeys.length === 0" class="p-2 text-center text-sm">
              {{ $t("profile.no_api_keys") }}
            </p>
            <article v-for="k in apiKeys" :key="k.id" class="p-2">
              <div class="flex flex-wrap items-center gap-2">
                <p class="mr-auto text-lg">{{ k.name }}</p>
                <Button variant="destructive" size="icon" :aria-label="$t('global.delete')" @click="revokeApiKey(k)">
                  <MdiDelete />
                </Button>
              </div>
              <div class="flex flex-wrap justify-between gap-x-4 gap-y-1 py-1 text-sm text-muted-foreground">
                <p>
                  {{ $t("profile.api_key_created") }}:
                  <DateTime format="relative" datetime-type="time" :date="k.createdAt" />
                </p>
                <p>
                  {{ $t("profile.api_key_last_used") }}:
                  <template v-if="k.lastUsedAt">
                    <DateTime format="relative" datetime-type="time" :date="k.lastUsedAt" />
                  </template>
                  <template v-else>{{ $t("profile.api_key_never_used") }}</template>
                </p>
                <p>
                  {{ $t("profile.api_key_expires_at") }}:
                  <template v-if="k.expiresAt">
                    <DateTime format="human" datetime-type="date" :date="k.expiresAt" />
                  </template>
                  <template v-else>{{ $t("profile.api_key_no_expiration") }}</template>
                </p>
              </div>
            </article>
          </div>

          <div class="mt-4">
            <Button variant="secondary" size="sm" @click="openCreateApiKey">
              {{ $t("profile.api_key_create") }}
            </Button>
          </div>
        </div>
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
            <span> {{ $t("profile.integrations") }} </span>
            <template #description>
              {{ $t("profile.integrations_sub") }}
            </template>
          </BaseSectionHeader>
        </template>

        <div class="space-y-6 px-4 pb-4">
          <!-- Paperless Settings -->
          <div class="space-y-2">
            <h4 class="font-semibold">{{ $t("profile.paperless_settings") }}</h4>
            <FormTextField
              v-model="integrationSettings.paperlessUrl"
              :label="$t('profile.paperless_url')"
              :placeholder="$t('profile.paperless_url_placeholder')"
              type="url"
              class="mb-2"
            />
            <FormTextField
              v-model="integrationSettings.paperlessToken"
              :label="$t('profile.paperless_token')"
              :placeholder="$t('profile.paperless_token_placeholder')"
              type="password"
            />
          </div>

          <!-- Immich Settings -->
          <div class="space-y-2 border-t pt-4">
            <h4 class="font-semibold">{{ $t("profile.immich_settings") }}</h4>
            <FormTextField
              v-model="integrationSettings.immichUrl"
              :label="$t('profile.immich_url')"
              :placeholder="$t('profile.immich_url_placeholder')"
              type="url"
              class="mb-2"
            />
            <FormTextField
              v-model="integrationSettings.immichToken"
              :label="$t('profile.immich_token')"
              :placeholder="$t('profile.immich_token_placeholder')"
              type="password"
            />
          </div>

          <div class="border-t pt-4">
            <Button
              :disabled="integrationSettings.saving || integrationSettings.loading"
              @click="saveIntegrationSettings"
            >
              <MdiLoading v-if="integrationSettings.saving" class="animate-spin" />
              {{ $t("global.save") }}
            </Button>
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
