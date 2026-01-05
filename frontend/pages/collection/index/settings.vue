<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import { Button } from "@/components/ui/button";
  import { Label } from "@/components/ui/label";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import MdiLoading from "~icons/mdi/loading";
  import FormTextField from "~/components/Form/TextField.vue";
  import type { CurrenciesCurrency, Group } from "~~/lib/api/types/data-contracts";
  import { fmtCurrencyAsync } from "~/composables/utils";
  import { getLocaleCode } from "~/composables/use-formatters";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();
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
      const res = await api.group.update(
        {
          name: name.value,
          currency: currencyCode.value,
        },
        selectedCollection.value.id
      );

      if (res.error || !res.data) {
        const msg = t("profile.toast.failed_update_group");
        error.value = msg;
        toast.error(msg);
        return;
      }

      group.value = res.data;
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
    </div>
  </div>
</template>
