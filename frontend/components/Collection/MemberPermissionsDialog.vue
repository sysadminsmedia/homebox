<template>
  <Dialog :dialog-id="DialogID.MemberPermissions">
    <DialogScrollContent>
      <DialogHeader>
        <DialogTitle>{{ $t("collection.members.permissions_title") }}</DialogTitle>
      </DialogHeader>

      <div v-if="loading" class="text-sm text-muted-foreground">
        {{ $t("global.loading") }}
      </div>

      <form v-else-if="member" class="flex flex-col gap-4" @submit.prevent="save">
        <div class="flex items-center gap-2">
          <Label>{{ $t("collection.members.role") }}</Label>
          <Badge variant="secondary" class="capitalize">{{ member.role }}</Badge>
        </div>

        <div class="flex flex-col gap-2">
          <Label>{{ $t("permissions.direct_permissions") }}</Label>
          <div class="flex flex-col gap-3 rounded-md border p-3">
            <div v-for="group in catalogByResource" :key="group.resource" class="flex flex-col gap-1.5">
              <span class="text-sm font-medium capitalize">{{ $t(`permissions.resource.${group.resource}`) }}</span>
              <div class="flex flex-wrap gap-x-4 gap-y-1.5">
                <label
                  v-for="def in group.definitions"
                  :key="def.key"
                  class="flex cursor-pointer items-center gap-1.5 text-sm"
                >
                  <Checkbox
                    :model-value="direct.includes(def.key)"
                    @update:model-value="toggleDirect(def.key, $event === true)"
                  />
                  {{ $t(`permissions.action.${def.action}`) }}
                </label>
              </div>
            </div>
          </div>
        </div>

        <div class="flex flex-col gap-2">
          <Label>{{ $t("permissions.permission_groups") }}</Label>
          <div class="flex flex-wrap gap-1.5">
            <Badge v-for="group in member.permissionGroups" :key="group.id" variant="secondary">
              {{ group.name }}
            </Badge>
            <span v-if="!member.permissionGroups.length" class="text-sm text-muted-foreground">
              {{ $t("permissions.no_permission_groups") }}
            </span>
          </div>
        </div>

        <div class="flex flex-col gap-2">
          <Label>{{ $t("permissions.effective_permissions") }}</Label>
          <div class="flex flex-wrap gap-1.5">
            <Badge v-for="key in member.effective" :key="key" variant="outline">
              {{ key }}
            </Badge>
            <span v-if="!member.effective.length" class="text-sm text-muted-foreground">
              {{ $t("permissions.no_permissions") }}
            </span>
          </div>
        </div>

        <DialogFooter>
          <Button type="submit" :disabled="saving">
            {{ $t("global.save") }}
          </Button>
        </DialogFooter>
      </form>
    </DialogScrollContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { expandPermissions, normalizePermissions } from "~~/lib/api/permissions-utils";
  import { useI18n } from "vue-i18n";
  import { Dialog, DialogFooter, DialogHeader, DialogScrollContent, DialogTitle } from "@/components/ui/dialog";
  import { Button } from "@/components/ui/button";
  import { Badge } from "@/components/ui/badge";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Label } from "@/components/ui/label";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import { toast } from "@/components/ui/sonner";
  import type { MemberPermissions, PermissionDefinition } from "~~/lib/api/types/data-contracts";

  const emit = defineEmits<{
    refresh: [];
  }>();

  const { t } = useI18n();
  const api = useUserApi();
  const { closeDialog, registerOpenDialogCallback } = useDialog();
  const { refresh: refreshSelfPermissions } = usePermissions();

  const loading = ref(false);
  const saving = ref(false);
  const member = ref<MemberPermissions | null>(null);
  const direct = ref<string[]>([]);
  const catalog = ref<PermissionDefinition[]>([]);

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

  async function load(userId: string) {
    loading.value = true;
    member.value = null;
    direct.value = [];

    try {
      const [memberRes] = await Promise.all([api.permissions.getMemberPermissions(userId), loadCatalog()]);
      if (memberRes.error || !memberRes.data) {
        toast.error(t("permissions.toast.load_failed"));
        closeDialog(DialogID.MemberPermissions);
        return;
      }

      member.value = memberRes.data;
      // Stored lists may contain wildcards ("*"); display concrete keys.
      direct.value = expandPermissions(memberRes.data.direct, catalog.value);
    } finally {
      loading.value = false;
    }
  }

  function toggleDirect(key: string, checked: boolean) {
    if (checked) {
      if (!direct.value.includes(key)) {
        direct.value = [...direct.value, key];
      }
    } else {
      direct.value = direct.value.filter(p => p !== key);
    }
  }

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.MemberPermissions, params => {
      saving.value = false;
      void load(params.userId);
    });

    onUnmounted(cleanup);
  });

  async function save() {
    if (saving.value || !member.value) return;

    saving.value = true;

    try {
      const { error } = await api.permissions.setMemberPermissions(
        member.value.userId,
        // Full selection persists as ["*"] so it covers future permissions.
        normalizePermissions(direct.value, catalog.value)
      );
      if (error) {
        toast.error(t("permissions.toast.save_failed"));
        return;
      }

      toast.success(t("permissions.toast.saved"));
      // The caller may have edited their own permissions; refresh the cache.
      void refreshSelfPermissions();
      closeDialog(DialogID.MemberPermissions);
      emit("refresh");
    } finally {
      saving.value = false;
    }
  }
</script>
