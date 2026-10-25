<template>
  <div class="flex flex-col gap-4">
    <div v-if="loading" class="text-sm text-muted-foreground">
      {{ $t("global.loading") }}
    </div>

    <div v-else class="flex flex-col gap-2">
      <p v-if="!grants.length" class="text-sm text-muted-foreground">
        {{ $t("items.sharing.empty") }}
      </p>

      <div v-for="grant in grants" :key="grant.id" class="flex items-center gap-2 rounded-md border p-2">
        <component
          :is="grant.targetType === 'user' ? MdiAccount : MdiAccountGroup"
          class="size-4 shrink-0 text-muted-foreground"
        />
        <div class="mr-auto min-w-0">
          <p class="truncate text-sm font-medium">{{ grant.targetName }}</p>
          <div class="flex flex-wrap gap-1">
            <Badge v-for="action in grant.actions" :key="action" variant="secondary" class="text-xs">
              {{ $t(`items.sharing.actions.${action}`) }}
            </Badge>
          </div>
        </div>
        <Button
          variant="ghost"
          size="icon"
          class="size-8 shrink-0 text-destructive"
          :aria-label="$t('global.delete')"
          :disabled="deleting[grant.id]"
          @click="removeGrant(grant)"
        >
          <MdiDelete class="size-4" />
        </Button>
      </div>
    </div>

    <form class="flex flex-col gap-3 rounded-md border p-3" @submit.prevent="addGrant">
      <p class="text-sm font-medium">{{ $t("items.sharing.add_grant") }}</p>

      <div class="grid gap-3 sm:grid-cols-2">
        <div class="flex flex-col gap-1.5">
          <Label>{{ $t("items.sharing.target_type") }}</Label>
          <Select v-model="targetType">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="user">{{ $t("items.sharing.user") }}</SelectItem>
              <SelectItem value="permissionGroup">{{ $t("items.sharing.permission_group") }}</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div class="flex flex-col gap-1.5">
          <Label>{{ $t("items.sharing.target") }}</Label>
          <Select v-model="targetId">
            <SelectTrigger>
              <SelectValue :placeholder="$t('global.select')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="option in targetOptions" :key="option.id" :value="option.id">
                {{ option.name }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div class="flex flex-col gap-1.5">
        <Label>{{ $t("items.sharing.grant_actions") }}</Label>
        <div class="flex flex-wrap gap-x-4 gap-y-1.5">
          <label v-for="action in grantActions" :key="action" class="flex cursor-pointer items-center gap-1.5 text-sm">
            <Checkbox
              :model-value="selectedActions.includes(action)"
              :disabled="action === 'read'"
              @update:model-value="toggleAction(action, $event === true)"
            />
            {{ $t(`items.sharing.actions.${action}`) }}
          </label>
        </div>
        <p class="text-xs text-muted-foreground">{{ $t("items.sharing.read_implied") }}</p>
      </div>

      <div class="flex justify-end">
        <Button type="submit" size="sm" :disabled="creating || !targetId">
          {{ $t("global.add") }}
        </Button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import MdiAccount from "~icons/mdi/account";
  import MdiAccountGroup from "~icons/mdi/account-group";
  import MdiDelete from "~icons/mdi/delete";
  import { Badge } from "@/components/ui/badge";
  import { Button } from "@/components/ui/button";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Label } from "@/components/ui/label";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { toast } from "@/components/ui/sonner";
  import type { AccessGrantOut, PermissionGroupOut, UserSummary } from "~~/lib/api/types/data-contracts";

  const props = defineProps<{
    entityId: string;
  }>();

  const { t } = useI18n();
  const api = useUserApi();
  const confirm = useConfirm();

  const grantActions = ["read", "update", "delete", "attachments"] as const;

  const loading = ref(true);
  const creating = ref(false);
  const deleting = ref<Record<string, boolean>>({});
  const grants = ref<AccessGrantOut[]>([]);

  const members = ref<UserSummary[]>([]);
  const permissionGroups = ref<PermissionGroupOut[]>([]);

  const targetType = ref<"user" | "permissionGroup">("user");
  const targetId = ref<string>("");
  const selectedActions = ref<string[]>(["read"]);

  const targetOptions = computed(() => {
    return targetType.value === "user"
      ? members.value.map(m => ({ id: m.id, name: m.name }))
      : permissionGroups.value.map(g => ({ id: g.id, name: g.name }));
  });

  watch(targetType, () => {
    targetId.value = "";
  });

  async function loadGrants() {
    const { data, error } = await api.permissions.getEntityGrants(props.entityId);
    if (error) {
      toast.error(t("items.sharing.toast.load_failed"));
      grants.value = [];
      return;
    }
    grants.value = data ?? [];
  }

  async function loadTargets() {
    const [membersRes, groupsRes] = await Promise.all([api.group.getMembers(), api.permissions.getPermissionGroups()]);
    members.value = membersRes.error ? [] : (membersRes.data ?? []);
    permissionGroups.value = groupsRes.error ? [] : (groupsRes.data ?? []);
  }

  onMounted(async () => {
    loading.value = true;
    try {
      await Promise.all([loadGrants(), loadTargets()]);
    } finally {
      loading.value = false;
    }
  });

  watch(
    () => props.entityId,
    async () => {
      loading.value = true;
      try {
        await loadGrants();
      } finally {
        loading.value = false;
      }
    }
  );

  function toggleAction(action: string, checked: boolean) {
    if (action === "read") {
      return;
    }
    if (checked) {
      if (!selectedActions.value.includes(action)) {
        selectedActions.value = [...selectedActions.value, action];
      }
    } else {
      selectedActions.value = selectedActions.value.filter(a => a !== action);
    }
  }

  async function addGrant() {
    if (creating.value || !targetId.value) {
      return;
    }

    creating.value = true;

    try {
      const { error } = await api.permissions.createEntityGrant(props.entityId, {
        targetType: targetType.value,
        targetId: targetId.value,
        actions: selectedActions.value,
      });

      if (error) {
        toast.error(t("items.sharing.toast.create_failed"));
        return;
      }

      toast.success(t("items.sharing.toast.created"));
      targetId.value = "";
      selectedActions.value = ["read"];
      await loadGrants();
    } finally {
      creating.value = false;
    }
  }

  async function removeGrant(grant: AccessGrantOut) {
    const { isCanceled } = await confirm.open(t("items.sharing.delete_confirm"));
    if (isCanceled) {
      return;
    }

    deleting.value = { ...deleting.value, [grant.id]: true };

    try {
      const { error } = await api.permissions.deleteEntityGrant(props.entityId, grant.id);
      if (error) {
        toast.error(t("items.sharing.toast.delete_failed"));
        return;
      }

      grants.value = grants.value.filter(g => g.id !== grant.id);
      toast.success(t("items.sharing.toast.deleted"));
    } finally {
      deleting.value = { ...deleting.value, [grant.id]: false };
    }
  }
</script>
