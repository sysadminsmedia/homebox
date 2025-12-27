<!-- TODO:
  - make collection on hover show role and colour based on role 
-->

<script setup lang="ts">
  import { ref, computed } from "vue";
  import { useI18n } from "vue-i18n";
  import { useConfirm } from "~/composables/use-confirm";
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow, TableEmpty } from "~/components/ui/table";
  import { Button } from "~/components/ui/button";
  import { Card } from "@/components/ui/card";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import MdiCheck from "~icons/mdi/check";
  import MdiClose from "~icons/mdi/close";
  // import MdiOpenInNew from "~icons/mdi/open-in-new";
  // Badge component for collections display
  import { Badge } from "@/components/ui/badge";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import UserFormDialog from "@/components/Admin/UserFormDialog.vue";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "@/components/ui/dialog-provider/utils";

  import { api, type Collection as MockCollection, type User } from "~/mock/collections";

  // api.getCollections returns collections augmented with `count` and the current user's `role`
  type CollectionSummary = MockCollection & { count: number; role: User["collections"][number]["role"] };

  const collections = ref<CollectionSummary[]>(api.getCollections() as CollectionSummary[]);
  const users = ref<User[]>(api.getUsers());

  const query = ref("");
  const filtered = computed(() => {
    const q = query.value.trim().toLowerCase();
    if (!q) return users.value;
    return users.value.filter(u => {
      return `${u.name} ${u.email} ${u.role}`.toLowerCase().includes(q);
    });
  });

  const { openDialog } = useDialog();
  const confirm = useConfirm();
  const { t } = useI18n();

  // editing state handled in dialog component; role toggle logic applied on save

  // helper to compute auth type for display
  // authType removed — not used in the template

  function authType(u: User) {
    const parts: string[] = [];
    if (u.password_set) parts.push("Password");
    if (u.oidc_subject) parts.push("OIDC");
    return parts.length ? parts.join(" & ") : "None";
  }

  function openAdd() {
    openDialog(DialogID.EditUser, {
      onClose: result => {
        if (result) {
          users.value = api.getUsers();
          collections.value = api.getCollections();
        }
      },
    });
  }

  function openEdit(u: User) {
    openDialog(DialogID.EditUser, {
      params: { userId: u.id },
      onClose: result => {
        if (result) {
          users.value = api.getUsers();
          collections.value = api.getCollections();
        }
      },
    });
  }

  async function confirmDelete(u: User) {
    const { isCanceled } = await confirm.open({
      message: t("global.delete_confirm") + " " + `${u.name} (${u.email})?`,
    });
    if (isCanceled) return;

    api.deleteUser(u.id);
    users.value = api.getUsers();
    // TODO: call backend API to delete user when available
  }

  // no more toggleActive; active is not used

  function collectionName(id: string) {
    const col = collections.value.find(c => c.id === id);
    return col ? col.name : id;
  }

  function roleVariant(role: string | undefined) {
    if (role === "owner") return "default";
    if (role === "admin") return "secondary";
    return "outline";
  }

  // dialog handles editing state now via dialog provider
</script>

<template>
  <BaseContainer class="flex flex-col gap-4">
    <BaseSectionHeader>
      <span>User Management</span>
      <div class="ml-auto">
        <Button @click="openAdd">Add User</Button>
      </div>
    </BaseSectionHeader>

    <Card class="p-0">
      <Table class="w-full">
        <TableHeader>
          <TableRow>
            <TableHead class="min-w-[160px]">{{ t("global.name") }}</TableHead>
            <TableHead class="min-w-[220px]">{{ t("global.email") }}</TableHead>
            <TableHead class="min-w-[96px] text-center">Is Admin</TableHead>
            <TableHead class="min-w-[220px]">Collections</TableHead>
            <TableHead class="min-w-[96px] text-center">{{ t("global.details") }}</TableHead>
            <TableHead class="w-40 text-center"></TableHead>
          </TableRow>
        </TableHeader>

        <TableBody>
          <template v-if="filtered.length">
            <TableRow v-for="u in filtered" :key="u.id">
              <TableCell>{{ u.name }}</TableCell>
              <TableCell>{{ u.email }}</TableCell>
              <TableCell class="text-center align-middle">
                <div class="flex size-full items-center justify-center font-medium">
                  <MdiCheck v-if="u.role === 'admin'" class="text-primary" />
                  <MdiClose v-else class="text-destructive" />
                </div>
              </TableCell>
              <TableCell>
                <div class="flex flex-wrap items-center gap-2">
                  <template v-if="u.collections && u.collections.length">
                    <TooltipProvider :delay-duration="0">
                      <template v-for="c in u.collections" :key="c.id">
                        <Tooltip>
                          <TooltipTrigger as-child>
                            <Badge class="whitespace-nowrap" :variant="roleVariant(c.role)">{{
                              collectionName(c.id)
                            }}</Badge>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p class="text-sm">{{ c.role }}</p>
                          </TooltipContent>
                        </Tooltip>
                      </template>
                    </TooltipProvider>
                  </template>
                  <span v-else class="text-muted-foreground">-</span>
                </div>
              </TableCell>
              <TableCell class="text-center align-middle">
                <div class="flex size-full items-center justify-center">
                  <span>{{ authType(u) }}</span>
                </div>
              </TableCell>
              <TableCell class="text-right align-middle">
                <div class="flex size-full items-center justify-end gap-2">
                  <Button size="icon" variant="outline" class="size-8" :title="t('global.edit')" @click="openEdit(u)">
                    <MdiPencil class="size-4" />
                  </Button>
                  <Button
                    size="icon"
                    variant="destructive"
                    class="size-8"
                    :title="t('global.delete')"
                    @click="confirmDelete(u)"
                  >
                    <MdiDelete class="size-4" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </template>

          <template v-else>
            <TableEmpty :colspan="6">
              <p>{{ $t("items.selector.no_results") }}</p>
            </TableEmpty>
          </template>
        </TableBody>
      </Table>
    </Card>

    <!-- Add / Edit form modal (moved to component) -->
    <UserFormDialog />
  </BaseContainer>
</template>
