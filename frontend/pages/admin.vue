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
  // import MdiOpenInNew from "~icons/mdi/open-in-new";
  // Badge component for collections display
  import { Badge } from "@/components/ui/badge";
  import UserFormDialog from "@/components/Admin/UserFormDialog.vue";

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

  const editing = ref<User | null>(null);
  const showForm = ref(false);
  const newPassword = ref("");
  const editingCollectionIds = ref<string[]>([]);
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
    editing.value = { id: String(Date.now()), name: "", email: "", role: "user", password_set: false, collections: [] };
    newPassword.value = "";
    editingCollectionIds.value = [];
    showForm.value = true;
  }

  function openEdit(u: User) {
    editing.value = { ...u };
    editingCollectionIds.value = (u.collections ?? []).map(c => c.id);
    newPassword.value = "";
    showForm.value = true;
  }

  function saveUser() {
    if (!editing.value) return;
    // basic validation
    if (!editing.value.name.trim() || !editing.value.email.trim()) {
      // keep UX simple: alert for now
      // Replace with a nicer notification component when available
      alert("Name and email are required");
      return;
    }

    // apply password flag if new password was set locally
    if (newPassword.value && editing.value) editing.value.password_set = true;

    const existing = api.getUser(editing.value.id);
    if (existing) {
      // update only scalar fields; collections are managed via the add/remove API
      const updated = {
        ...existing,
        name: editing.value.name,
        email: editing.value.email,
        role: editing.value.role,
        password_set: editing.value.password_set,
      } as User;
      api.updateUser(updated);
    } else {
      // create user without collections first, then add memberships
      const toCreate = { ...editing.value, collections: [] } as User;
      const created = api.addUser(toCreate);
      editingCollectionIds.value.forEach(id => api.addUserToCollection(created.id, id, "viewer"));
    }

    // refresh local cache
    users.value = api.getUsers();

    editing.value = null;
    showForm.value = false;
    // TODO: call backend API to persist changes when available
  }

  function cancelForm() {
    editing.value = null;
    showForm.value = false;
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

  function onUpdateEditing(val: User | null) {
    editing.value = val;
  }

  function onUpdateEditingCollectionIds(val: string[]) {
    editingCollectionIds.value = val;
  }

  function onUpdateNewPassword(val: string) {
    newPassword.value = val;
  }
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
            <TableHead>{{ t("global.name") }}</TableHead>
            <TableHead>{{ t("global.email") }}</TableHead>
            <TableHead>Is Admin</TableHead>
            <TableHead>Collections</TableHead>
            <TableHead class="w-32 text-center">Auth</TableHead>
            <TableHead class="w-40 text-center">{{ t("global.details") }}</TableHead>
          </TableRow>
        </TableHeader>

        <TableBody>
          <template v-if="filtered.length">
            <TableRow v-for="u in filtered" :key="u.id">
              <TableCell>{{ u.name }}</TableCell>
              <TableCell>{{ u.email }}</TableCell>
              <TableCell class="text-center">
                <span class="font-medium">{{ u.role === "admin" ? "Yes" : "No" }}</span>
              </TableCell>
              <TableCell>
                <div class="flex flex-wrap items-center gap-2">
                  <template v-if="u.collections && u.collections.length">
                    <Badge v-for="c in u.collections" :key="c.id" class="whitespace-nowrap"
                      >{{ collectionName(c.id) }}<span class="ml-1 text-xs opacity-60">({{ c.role }})</span></Badge
                    >
                  </template>
                  <span v-else class="text-muted-foreground">-</span>
                </div>
              </TableCell>
              <TableCell class="text-center">
                <span>{{ authType(u) }}</span>
              </TableCell>
              <TableCell class="text-right">
                <div class="flex justify-end gap-2">
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
    <UserFormDialog
      v-model="showForm"
      :editing="editing"
      :collections="collections"
      :editing-collection-ids="editingCollectionIds"
      :new-password="newPassword"
      @update:editing="onUpdateEditing"
      @update:editing-collection-ids="onUpdateEditingCollectionIds"
      @update:new-password="onUpdateNewPassword"
      @save="saveUser"
      @cancel="cancelForm"
      @collections-changed="() => (collections = api.getCollections())"
    />
  </BaseContainer>
</template>
