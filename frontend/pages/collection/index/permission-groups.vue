<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
  import { Badge } from "@/components/ui/badge";
  import { Button } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import MdiDelete from "~icons/mdi/delete";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiPlus from "~icons/mdi/plus";
  import { toast } from "@/components/ui/sonner";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import CollectionPermissionGroupDialog from "~/components/Collection/PermissionGroupDialog.vue";
  import type { PermissionGroupOut } from "~~/lib/api/types/data-contracts";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

  useHead({ title: `HomeBox | ${t("collection.tabs.permission_groups")}` });

  const api = useUserApi();
  const confirm = useConfirm();
  const { openDialog } = useDialog();

  const loading = ref(true);
  const groups = ref<PermissionGroupOut[]>([]);
  const deleting = ref<Record<string, boolean>>({});

  const loadGroups = async () => {
    loading.value = true;

    try {
      const res = await api.permissions.getPermissionGroups();
      if (res.error) {
        const msg = t("errors.api_failure") + String(res.error);
        groups.value = [];
        toast.error(msg);
      } else {
        groups.value = res.data ?? [];
      }
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      groups.value = [];
      toast.error(msg);
    } finally {
      loading.value = false;
    }
  };

  const handleCreate = () => {
    openDialog(DialogID.PermissionGroup);
  };

  const handleEdit = (group: PermissionGroupOut) => {
    openDialog(DialogID.PermissionGroup, { params: { group } });
  };

  const handleDelete = async (group: PermissionGroupOut) => {
    const result = await confirm.open(t("collection.permission_groups.delete_confirm"));
    if (result.isCanceled) {
      return;
    }

    deleting.value = { ...deleting.value, [group.id]: true };

    try {
      const res = await api.permissions.deletePermissionGroup(group.id);
      if (res.error) {
        const msg = t("errors.api_failure") + String(res.error);
        toast.error(msg);
      } else {
        groups.value = groups.value.filter(g => g.id !== group.id);
        toast.success(t("collection.permission_groups.toast.deleted"));
      }
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      toast.error(msg);
    } finally {
      deleting.value = { ...deleting.value, [group.id]: false };
    }
  };

  onMounted(() => {
    loadGroups();
  });
</script>

<template>
  <div class="space-y-4">
    <CollectionPermissionGroupDialog @refresh="loadGroups" />

    <div class="flex justify-end">
      <Button size="sm" @click="handleCreate">
        <MdiPlus class="mr-1 size-4" />
        {{ $t("global.create") }}
      </Button>
    </div>

    <div v-if="loading" class="rounded-md border bg-card p-4 text-sm text-muted-foreground">
      {{ $t("global.loading") }}
    </div>

    <div v-else>
      <div v-if="!groups.length" class="rounded-md border bg-card p-4 text-sm text-muted-foreground">
        {{ $t("collection.permission_groups.empty") }}
      </div>

      <div v-else class="scroll-bg overflow-x-auto rounded-md border bg-card">
        <Table class="min-w-[480px]">
          <TableHeader>
            <TableRow>
              <TableHead>{{ $t("collection.permission_groups.name") }}</TableHead>
              <TableHead>{{ $t("collection.permission_groups.members") }}</TableHead>
              <TableHead>{{ $t("permissions.permissions") }}</TableHead>
              <TableHead class="w-32 text-right"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="group in groups" :key="group.id">
              <TableCell>
                <p class="font-medium">{{ group.name }}</p>
                <p v-if="group.description" class="text-xs text-muted-foreground">{{ group.description }}</p>
              </TableCell>
              <TableCell>{{ group.members.length }}</TableCell>
              <TableCell>
                <div class="flex max-w-md flex-wrap gap-1">
                  <Badge v-for="key in group.permissions" :key="key" variant="secondary" class="text-xs">
                    {{ key }}
                  </Badge>
                  <span v-if="!group.permissions.length" class="text-xs text-muted-foreground">
                    {{ $t("permissions.no_permissions") }}
                  </span>
                </div>
              </TableCell>
              <TableCell>
                <div class="ml-auto flex justify-end gap-1">
                  <TooltipProvider :delay-duration="0">
                    <Tooltip>
                      <TooltipTrigger as-child>
                        <Button
                          variant="ghost"
                          size="icon"
                          class="size-8"
                          :aria-label="$t('global.edit')"
                          @click="handleEdit(group)"
                        >
                          <MdiPencil class="size-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        {{ $t("global.edit") }}
                      </TooltipContent>
                    </Tooltip>
                    <Tooltip>
                      <TooltipTrigger as-child>
                        <Button
                          variant="destructive"
                          size="icon"
                          class="size-8"
                          :aria-label="$t('global.delete')"
                          :disabled="deleting[group.id]"
                          @click="handleDelete(group)"
                        >
                          <MdiDelete class="size-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        {{ $t("global.delete") }}
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </div>
  </div>
</template>
