<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiLoading from "~icons/mdi/loading";
  import { Button } from "@/components/ui/button";
  import FormTextField from "~/components/Form/TextField.vue";
  import BaseCard from "@/components/Base/Card.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";

  type IntegrationSettingsState = Record<string, string | boolean> & {
    loading: boolean;
    saving: boolean;
  };

  const { t } = useI18n();
  const api = useUserApi();

  const settings = reactive<IntegrationSettingsState>({
    loading: false,
    saving: false,
  });

  function setting(key: string): string {
    const value = settings[key];
    return typeof value === "string" ? value : "";
  }

  const paperlessUrl = computed({
    get: () => setting("paperless_url"),
    set: value => {
      settings.paperless_url = value;
    },
  });

  const paperlessToken = computed({
    get: () => setting("paperless_token"),
    set: value => {
      settings.paperless_token = value;
    },
  });

  async function loadSettings() {
    settings.loading = true;
    const { data, error } = await api.user.getSettings();
    settings.loading = false;

    if (error || !data?.item) {
      toast.error(t("errors.api_failure"));
      return;
    }

    const item = data.item as Record<string, unknown>;
    settings.paperless_url = typeof item.paperless_url === "string" ? item.paperless_url : "";
    settings.paperless_token = typeof item.paperless_token === "string" ? item.paperless_token : "";
  }

  async function saveSettings() {
    settings.saving = true;
    const { data: current, error: currentError } = await api.user.getSettings();
    if (currentError) {
      settings.saving = false;
      toast.error(t("profile.toast.failed_settings_save"));
      return;
    }

    const { error } = await api.user.setSettings({
      ...(current?.item ?? {}),
      paperless_url: setting("paperless_url"),
      paperless_token: setting("paperless_token"),
    });
    settings.saving = false;

    if (error) {
      toast.error(t("profile.toast.failed_settings_save"));
      return;
    }

    toast.success(t("profile.toast.settings_saved"));
  }

  onMounted(() => {
    void loadSettings();
  });
</script>

<template>
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
      <div class="space-y-2">
        <h4 class="font-semibold">{{ $t("profile.paperless_settings") }}</h4>
        <FormTextField
          v-model="paperlessUrl"
          :label="$t('profile.paperless_url')"
          :placeholder="$t('profile.paperless_url_placeholder')"
          type="url"
          class="mb-2"
        />
        <FormTextField
          v-model="paperlessToken"
          :label="$t('profile.paperless_token')"
          :placeholder="$t('profile.paperless_token_placeholder')"
          type="password"
        />
      </div>

      <div class="border-t pt-4">
        <Button :disabled="settings.saving || settings.loading" @click="saveSettings">
          <MdiLoading v-if="settings.saving" class="animate-spin" />
          {{ $t("global.save") }}
        </Button>
      </div>
    </div>
  </BaseCard>
</template>
