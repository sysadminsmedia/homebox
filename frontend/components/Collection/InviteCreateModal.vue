<template>
  <BaseModal :dialog-id="DialogID.CreateGroupInvite" :title="$t('collection.create_invite')" :hide-footer="true">
    <form class="flex min-w-0 flex-col gap-4" @submit.prevent="create">
      <FormTextField v-model="form.uses" :label="$t('collection.uses')" type="number" :required="true" />

      <div class="flex w-full flex-col gap-1.5">
        <Label class="cursor-pointer">{{ $t("collection.expires_at") }}</Label>
        <VueDatePicker
          v-model="form.expiresAt"
          :enable-time-picker="true"
          clearable
          :dark="isDark"
          :format="formatDateTime"
        />
      </div>

      <div class="flex w-full flex-col gap-1.5">
        <button
          type="button"
          class="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
          @click="showPermissions = !showPermissions"
        >
          <MdiChevronRight class="size-4 transition-transform" :class="{ 'rotate-90': showPermissions }" />
          {{ $t("components.collection.invite_create_modal.permissions") }}
        </button>

        <div v-if="showPermissions" class="flex flex-col gap-3 rounded-md border p-3">
          <p class="text-xs text-muted-foreground">
            {{ $t("components.collection.invite_create_modal.permissions_hint") }}
          </p>
          <div v-for="group in catalogByResource" :key="group.resource" class="flex flex-col gap-1.5">
            <span class="text-sm font-medium capitalize">{{ $t(`permissions.resource.${group.resource}`) }}</span>
            <div class="flex flex-wrap gap-x-4 gap-y-1.5">
              <label
                v-for="def in group.definitions"
                :key="def.key"
                class="flex cursor-pointer items-center gap-1.5 text-sm"
              >
                <Checkbox
                  :model-value="selectedPermissions.includes(def.key)"
                  @update:model-value="togglePermission(def.key, $event === true)"
                />
                {{ $t(`permissions.action.${def.action}`) }}
              </label>
            </div>
          </div>
        </div>
      </div>

      <div class="mt-4 flex flex-row-reverse">
        <ButtonGroup>
          <Button :disabled="loading" type="submit">
            {{ $t("global.create") }}
          </Button>
        </ButtonGroup>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import { normalizePermissions } from "~~/lib/api/permissions-utils";
  import { useI18n } from "vue-i18n";
  import VueDatePicker from "@vuepic/vue-datepicker";
  import "@vuepic/vue-datepicker/dist/main.css";
  import MdiChevronRight from "~icons/mdi/chevron-right";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { useDialog } from "~/components/ui/dialog-provider";
  import BaseModal from "@/components/App/CreateModal.vue";
  import FormTextField from "~/components/Form/TextField.vue";
  import { Button, ButtonGroup } from "~/components/ui/button";
  import { Checkbox } from "~/components/ui/checkbox";
  import { Label } from "~/components/ui/label";
  import { toast } from "@/components/ui/sonner";
  import { useUserApi } from "~/composables/use-api";
  import { darkThemes } from "~/lib/data/themes";
  import type { PermissionDefinition } from "~~/lib/api/types/data-contracts";

  const { t } = useI18n();
  const { activeDialog, closeDialog } = useDialog();
  const api = useUserApi();

  const loading = ref(false);
  const form = reactive<{ uses: number; expiresAt: Date | null }>({
    uses: 1,
    expiresAt: defaultExpiry(),
  });

  const isDark = useIsThemeInList(darkThemes);

  const formatDateTime = (date: Date | string | number) => fmtDate(date, "human", "datetime");

  function defaultExpiry(): Date {
    return new Date(Date.now() + 7 * 24 * 60 * 60 * 1000);
  }

  const showPermissions = ref(false);
  const catalog = ref<PermissionDefinition[]>([]);
  const selectedPermissions = ref<string[]>([]);

  const catalogByResource = computed(() => {
    const grouped = new Map<string, PermissionDefinition[]>();
    for (const def of catalog.value) {
      const list = grouped.get(def.resource) ?? [];
      list.push(def);
      grouped.set(def.resource, list);
    }
    return [...grouped.entries()].map(([resource, definitions]) => ({ resource, definitions }));
  });

  async function loadCatalog() {
    if (catalog.value.length) return;
    const { data, error } = await api.permissions.getCatalog();
    if (!error && data) {
      catalog.value = data;
    }
  }

  function togglePermission(key: string, checked: boolean) {
    if (checked) {
      if (!selectedPermissions.value.includes(key)) {
        selectedPermissions.value = [...selectedPermissions.value, key];
      }
    } else {
      selectedPermissions.value = selectedPermissions.value.filter(p => p !== key);
    }
  }

  watch(
    () => activeDialog.value,
    active => {
      if (active && active === DialogID.CreateGroupInvite) {
        form.uses = 1;
        form.expiresAt = defaultExpiry();
        loading.value = false;
        showPermissions.value = false;
        // Default to full access: all catalog permissions checked.
        void loadCatalog().then(() => {
          selectedPermissions.value = catalog.value.map(def => def.key);
        });
      }
    }
  );

  async function create() {
    if (loading.value) {
      return;
    }

    const parsedUses = Number(form.uses ?? 0);
    if (!Number.isFinite(parsedUses) || parsedUses < 1 || parsedUses > 100) {
      toast.error(t("components.collection.invite_create_modal.toast.invalid_uses"));
      return;
    }
    const uses = parsedUses;

    if (!form.expiresAt) {
      toast.error(t("components.collection.invite_create_modal.toast.invalid_expiry_missing"));
      return;
    }

    const now = new Date();
    const exp = new Date(form.expiresAt);
    if (exp.getTime() <= now.getTime()) {
      toast.error(t("components.collection.invite_create_modal.toast.invalid_expiry_past"));
      return;
    }

    const expiresAtToSend: Date = exp;

    loading.value = true;

    try {
      const res = await api.group.createInvitation({
        expiresAt: expiresAtToSend,
        uses,
        // Full selection persists as ["*"] so it covers future permissions.
        permissions: normalizePermissions(selectedPermissions.value, catalog.value),
      });

      if (res.error) {
        const msg = t("errors.api_failure") + String(res.error);
        toast.error(msg);
        loading.value = false;
        return;
      }

      const data = res.data ?? undefined;
      closeDialog(DialogID.CreateGroupInvite, data);
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      toast.error(msg);
    } finally {
      loading.value = false;
    }
  }
</script>
