<template>
  <Dialog :dialog-id="DialogID.PermissionGroup">
    <DialogScrollContent>
      <DialogHeader>
        <DialogTitle>
          {{
            editingId ? $t("collection.permission_groups.edit_title") : $t("collection.permission_groups.create_title")
          }}
        </DialogTitle>
      </DialogHeader>

      <form class="flex flex-col gap-4" @submit.prevent="save">
        <FormTextField
          v-model="form.name"
          :autofocus="true"
          :label="$t('collection.permission_groups.name')"
          :max-length="255"
          :min-length="1"
        />
        <FormTextArea
          v-model="form.description"
          :label="$t('collection.permission_groups.description')"
          :max-length="1000"
        />

        <div class="flex flex-col gap-2">
          <Label>{{ $t("permissions.permissions") }}</Label>
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
                    :model-value="selectedPermissions.includes(def.key)"
                    @update:model-value="togglePermission(def.key, $event === true)"
                  />
                  {{ $t(`permissions.action.${def.action}`) }}
                </label>
              </div>
            </div>
          </div>
        </div>

        <div class="flex flex-col gap-2">
          <Label>{{ $t("collection.permission_groups.members") }}</Label>
          <div class="flex max-h-48 flex-col gap-1.5 overflow-y-auto rounded-md border p-3">
            <p v-if="!members.length" class="text-sm text-muted-foreground">
              {{ $t("collection.members.empty") }}
            </p>
            <label v-for="member in members" :key="member.id" class="flex cursor-pointer items-center gap-1.5 text-sm">
              <Checkbox
                :model-value="selectedMemberIds.includes(member.id)"
                @update:model-value="toggleMember(member.id, $event === true)"
              />
              {{ member.name }}
              <span class="text-muted-foreground">{{ member.email }}</span>
            </label>
          </div>
        </div>

        <DialogFooter>
          <Button type="submit" :disabled="saving">
            {{ editingId ? $t("global.update") : $t("global.create") }}
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
  import { Checkbox } from "@/components/ui/checkbox";
  import { Label } from "@/components/ui/label";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import { toast } from "@/components/ui/sonner";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import type { PermissionDefinition, UserSummary } from "~~/lib/api/types/data-contracts";

  const emit = defineEmits<{
    refresh: [];
  }>();

  const { t } = useI18n();
  const api = useUserApi();
  const { closeDialog, registerOpenDialogCallback } = useDialog();

  const editingId = ref<string | null>(null);
  const form = reactive({
    name: "",
    description: "",
  });
  const selectedPermissions = ref<string[]>([]);
  const selectedMemberIds = ref<string[]>([]);
  const saving = ref(false);

  const catalog = ref<PermissionDefinition[]>([]);
  const members = ref<UserSummary[]>([]);

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

  async function loadMembers() {
    const { data, error } = await api.group.getMembers();
    if (!error && data) {
      members.value = data;
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

  function toggleMember(id: string, checked: boolean) {
    if (checked) {
      if (!selectedMemberIds.value.includes(id)) {
        selectedMemberIds.value = [...selectedMemberIds.value, id];
      }
    } else {
      selectedMemberIds.value = selectedMemberIds.value.filter(m => m !== id);
    }
  }

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.PermissionGroup, params => {
      const group = params?.group;
      editingId.value = group?.id ?? null;
      form.name = group?.name ?? "";
      form.description = group?.description ?? "";
      const storedPermissions = group?.permissions ?? [];
      selectedPermissions.value = [...storedPermissions];
      selectedMemberIds.value = group?.members?.map(m => m.id) ?? [];
      saving.value = false;

      // Stored lists may contain wildcards ("*"); display concrete keys once
      // the catalog is available.
      void loadCatalog().then(() => {
        selectedPermissions.value = expandPermissions(storedPermissions, catalog.value);
      });
      void loadMembers();
    });

    onUnmounted(cleanup);
  });

  async function save() {
    if (saving.value) return;

    if (!form.name.trim()) {
      toast.error(t("collection.permission_groups.toast.name_required"));
      return;
    }

    saving.value = true;

    try {
      const payload = {
        name: form.name.trim(),
        description: form.description,
        // Full selection persists as ["*"] so it covers future permissions.
        permissions: normalizePermissions(selectedPermissions.value, catalog.value),
      };

      const res = editingId.value
        ? await api.permissions.updatePermissionGroup(editingId.value, payload)
        : await api.permissions.createPermissionGroup(payload);

      if (res.error || !res.data) {
        toast.error(t("collection.permission_groups.toast.save_failed"));
        return;
      }

      const membersRes = await api.permissions.setPermissionGroupMembers(res.data.id, selectedMemberIds.value);
      if (membersRes.error) {
        toast.error(t("collection.permission_groups.toast.members_failed"));
        return;
      }

      toast.success(
        editingId.value
          ? t("collection.permission_groups.toast.updated")
          : t("collection.permission_groups.toast.created")
      );
      closeDialog(DialogID.PermissionGroup);
      emit("refresh");
    } finally {
      saving.value = false;
    }
  }
</script>
