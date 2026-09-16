<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import { Button } from "@/components/ui/button";
  import { Label } from "@/components/ui/label";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Switch } from "@/components/ui/switch";
  import MdiLoading from "~icons/mdi/loading";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import type { CurrenciesCurrency, Group, GroupUpdate } from "~~/lib/api/types/data-contracts";
  import { fmtCurrencyAsync } from "~/composables/utils";
  import { utf8Length } from "@/lib/utils";

  // The generated GroupUpdate type marks every field (including
  // foundContactEnabled/foundContactMessage) as required, but the backend applies pointer
  // semantics: omitted fields are left unchanged. Redeclared here as partial to match reality.
  type GroupUpdatePayload = Partial<GroupUpdate>;

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

  useHead({ title: `HomeBox | ${t("collection.tabs.settings")}` });

  const api = useUserApi();
  const { selectedCollection, load: reloadCollections } = useCollections();

  const loading = ref(true);
  const saving = ref(false);
  const error = ref<string | null>(null);

  const group = ref<Group | null>(null);
  const currencies = ref<CurrenciesCurrency[]>([]);
  const name = ref("");
  const currencyCode = ref("USD");
  const currencyExample = ref("$1,000.00");

  const foundContactEnabled = ref(false);
  const foundContactMessage = ref("");
  const savingFoundContact = ref(false);

  const loadSettings = async () => {
    if (!selectedCollection.value) {
      loading.value = false;
      return;
    }

    loading.value = true;
    error.value = null;

    try {
      if (!currencies.value.length) {
        const respCurrencies = await api.group.currencies();
        if (respCurrencies.error) {
          toast.error(t("profile.toast.failed_get_currencies"));
        } else if (respCurrencies.data) {
          currencies.value = respCurrencies.data;
        }
      }

      const res = await api.group.get(selectedCollection.value.id);
      if (res.error || !res.data) {
        const msg = t("errors.api_failure") + String(res.error ?? "");
        error.value = msg;
        toast.error(msg);
        return;
      }

      group.value = res.data;
      name.value = res.data.name;
      currencyCode.value = res.data.currency;
      foundContactEnabled.value = res.data.foundContactEnabled;
      foundContactMessage.value = res.data.foundContactMessage ?? "";
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      error.value = msg;
      toast.error(msg);
    } finally {
      loading.value = false;
    }
  };

  watch(
    () => selectedCollection.value?.id,
    () => {
      void loadSettings();
    },
    { immediate: true }
  );

  watch(
    currencyCode,
    async () => {
      if (!currencyCode.value) return;
      try {
        currencyExample.value = await fmtCurrencyAsync(1000, currencyCode.value, getLocaleCode());
      } catch {
        currencyExample.value = `${currencyCode.value} 1000`;
      }
    },
    { immediate: true }
  );

  const save = async () => {
    if (!selectedCollection.value) return;

    saving.value = true;
    error.value = null;

    try {
      const payload: GroupUpdatePayload = {
        name: name.value,
        currency: currencyCode.value,
      };
      const res = await api.group.update(payload as GroupUpdate, selectedCollection.value.id);

      if (res.error || !res.data) {
        const msg = t("profile.toast.failed_update_group");
        error.value = msg;
        toast.error(msg);
        return;
      }

      group.value = res.data;
      setCurrency(res.data.currency);
      toast.success(t("profile.toast.group_updated"));

      await reloadCollections();
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      error.value = msg;
      toast.error(msg);
    } finally {
      saving.value = false;
    }
  };

  const saveFoundContact = async () => {
    if (!selectedCollection.value || !group.value) return;

    savingFoundContact.value = true;
    error.value = null;

    try {
      const payload: GroupUpdatePayload = {
        name: group.value.name,
        currency: group.value.currency,
        foundContactEnabled: foundContactEnabled.value,
        foundContactMessage: foundContactMessage.value,
      };
      const res = await api.group.update(payload as GroupUpdate, selectedCollection.value.id);

      if (res.error || !res.data) {
        const msg = t("profile.toast.failed_update_group");
        error.value = msg;
        toast.error(msg);
        return;
      }

      group.value = res.data;
      foundContactEnabled.value = res.data.foundContactEnabled;
      foundContactMessage.value = res.data.foundContactMessage ?? "";
      toast.success(t("found.settings.saved"));
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      error.value = msg;
      toast.error(msg);
    } finally {
      savingFoundContact.value = false;
    }
  };
</script>

<template>
  <div class="space-y-4">
    <div v-if="loading" class="rounded-md border bg-card p-4 text-sm text-muted-foreground">
      {{ $t("global.loading") }}
    </div>

    <div v-else>
      <div v-if="!selectedCollection" class="rounded-md border bg-card p-4 text-sm text-muted-foreground">
        {{ $t("components.collection.selector.select_collection") }}
      </div>

      <div v-else class="space-y-4 rounded-md border bg-card p-4">
        <FormTextField v-model="name" :label="$t('global.name')" />

        <div>
          <Label for="currency"> {{ $t("profile.currency_format") }} </Label>
          <Select
            id="currency"
            :model-value="currencyCode"
            @update:model-value="val => (currencyCode = String(val || ''))"
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="c in currencies" :key="c.code" :value="c.code">
                {{ c.name }}
              </SelectItem>
            </SelectContent>
          </Select>
          <p class="m-2 text-sm">{{ $t("profile.example") }}: {{ currencyExample }}</p>
        </div>

        <div class="mt-4">
          <Button variant="secondary" size="sm" :disabled="saving" @click="save">
            <MdiLoading v-if="saving" class="mr-2 inline-block animate-spin" />
            <span>{{ $t("profile.update_group") }}</span>
          </Button>
        </div>
      </div>

      <div v-if="selectedCollection" class="mt-4 space-y-4 rounded-md border bg-card p-4">
        <div>
          <h2 class="text-lg font-medium">{{ $t("found.settings.title") }}</h2>
          <p class="text-sm text-muted-foreground">{{ $t("found.settings.description") }}</p>
        </div>

        <div class="flex items-center gap-2">
          <Switch id="found-contact-enabled" v-model="foundContactEnabled" />
          <Label for="found-contact-enabled">{{ $t("found.settings.enable") }}</Label>
        </div>

        <FormTextArea
          v-model="foundContactMessage"
          :label="$t('found.settings.message_label')"
          :placeholder="$t('found.settings.message_placeholder')"
          :max-length="500"
        />

        <div class="rounded-md border border-accent-foreground bg-accent p-4 text-accent-foreground">
          <p class="text-sm">{{ $t("found.settings.no_smtp_warning") }}</p>
        </div>

        <div class="mt-4">
          <Button
            variant="secondary"
            size="sm"
            :disabled="savingFoundContact || utf8Length(foundContactMessage) > 500"
            @click="saveFoundContact"
          >
            <MdiLoading v-if="savingFoundContact" class="mr-2 inline-block animate-spin" />
            <span>{{ $t("global.save") }}</span>
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
