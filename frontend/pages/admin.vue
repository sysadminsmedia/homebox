<script setup lang="ts">
  import { ref, computed } from "vue";
  import { useI18n } from "vue-i18n";
  import { useConfirm } from "~/composables/use-confirm";
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow, TableEmpty } from "~/components/ui/table";
  import { Button } from "~/components/ui/button";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import MdiAccountMultiple from "~icons/mdi/account-multiple";
  import MdiOpenInNew from "~icons/mdi/open-in-new";
  import MdiCheck from "~icons/mdi/check";

  type Group = { id: string; name: string; ownerName?: string };

  type User = {
    id: string;
    name: string;
    email: string;
    role: "admin" | "user" | string;
    // password_set indicates whether the user has a local password
    password_set?: boolean;
    group?: Group | null;
    oidc_subject?: string | null;
    oidc_issuer?: string | null;
  };

  // Mock groups (group.name is the owner's name per your request)
  const groups = ref<Group[]>([
    { id: "g1", name: "Alice Admin" },
    { id: "g2", name: "Owner Two" },
  ]);

  const users = ref<User[]>([
    {
      id: "1",
      name: "Alice Admin",
      email: "alice@example.com",
      role: "admin",
      password_set: true,
      group: groups.value[0],
    },
    {
      id: "2",
      name: "Bob User",
      email: "bob@example.com",
      role: "user",
      password_set: true,
      group: groups.value[0],
      oidc_subject: "bob-sub",
      oidc_issuer: "https://oidc.example.com",
    },
    { id: "3", name: "Charlie", email: "charlie@example.com", role: "user", password_set: false, group: null },
  ]);

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
  const editingGroupId = ref<string | null>(null);
  const confirm = useConfirm();
  const { t } = useI18n();

  const isEditingExisting = computed(() => editing.value !== null && users.value.some(u => u.id === editing.value!.id));

  const editingIsAdmin = computed({
    get: () => editing.value?.role === "admin",
    set: (v: boolean) => {
      if (!editing.value) return;
      editing.value.role = v ? "admin" : "user";
    },
  });

  // helper to compute auth type for display
  // authType removed — not used in the template

  function openAdd() {
    editing.value = { id: String(Date.now()), name: "", email: "", role: "user", password_set: false, group: null };
    newPassword.value = "";
    editingGroupId.value = null;
    showForm.value = true;
  }

  function openEdit(u: User) {
    editing.value = { ...u };
    editingGroupId.value = u.group?.id ?? null;
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

    const idx = users.value.findIndex(x => x.id === editing.value!.id);
    if (idx >= 0) {
      // apply password flag if new password was set locally
      if (newPassword.value && editing.value) editing.value.password_set = true;
      // apply group selection object
      if (editing.value) {
        editing.value.group = groups.value.find(g => g.id === editingGroupId.value) ?? null;
      }
      users.value.splice(idx, 1, { ...editing.value });
    } else {
      if (newPassword.value && editing.value) editing.value.password_set = true;
      if (editing.value) editing.value.group = groups.value.find(g => g.id === editingGroupId.value) ?? null;
      users.value.unshift({ ...editing.value });
    }

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

    users.value = users.value.filter(x => x.id !== u.id);
    // TODO: call backend API to delete user when available
  }

  // no more toggleActive; active is not used
</script>

<template>
  <div class="mx-auto max-w-6xl p-6">
    <header class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-semibold">{{ t("global.details") }} - Administration</h1>
      <div class="flex items-center gap-3">
        <input v-model="query" :placeholder="t('global.search')" class="rounded border px-3 py-2" />
        <Button @click="openAdd">{{ t("global.add") }}</Button>
      </div>
    </header>

    <section>
      <Table class="w-full">
        <TableHeader>
          <TableRow>
            <TableHead>{{ t("global.name") }}</TableHead>
            <TableHead>{{ t("global.email") }}</TableHead>
            <TableHead>Role</TableHead>
            <TableHead>Group</TableHead>
            <TableHead class="w-32 text-center">Auth</TableHead>
            <TableHead class="w-40 text-center">{{ t("global.details") }}</TableHead>
          </TableRow>
        </TableHeader>

        <TableBody>
          <template v-if="filtered.length">
            <TableRow v-for="u in filtered" :key="u.id">
              <TableCell>{{ u.name }}</TableCell>
              <TableCell>{{ u.email }}</TableCell>
              <TableCell class="flex items-center gap-2">
                <MdiCheck v-if="u.role === 'admin'" class="size-4 text-green-600" />
                <span v-if="u.role === 'admin'">admin</span>
                <span v-else>-</span>
              </TableCell>
              <TableCell>
                <div class="flex items-center gap-2">
                  <MdiAccountMultiple class="size-4" />
                  <span>{{ u.group?.name ?? "-" }}</span>
                </div>
              </TableCell>
              <TableCell class="text-center">
                <span v-if="u.oidc_subject" :title="u.oidc_issuer || u.oidc_subject">
                  <MdiOpenInNew class="inline-block size-4" />
                </span>
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
    </section>

    <!-- Add / Edit form modal (simple) -->
    <div v-if="showForm" class="fixed inset-0 z-40 flex items-center justify-center bg-black/40">
      <div class="w-full max-w-md rounded bg-white p-6 shadow-lg">
        <h2 class="mb-4 text-lg font-medium">{{ isEditingExisting ? t("global.edit") : t("global.add") }}</h2>
        <div class="space-y-3">
          <label class="block">
            <div class="mb-1 text-sm">{{ t("global.name") }}</div>
            <input v-model="editing!.name" class="w-full rounded border px-3 py-2" />
          </label>
          <label class="block">
            <div class="mb-1 text-sm">{{ t("global.email") }}</div>
            <input v-model="editing!.email" class="w-full rounded border px-3 py-2" />
          </label>
          <label class="flex items-center gap-2">
            <input v-model="editingIsAdmin" type="checkbox" />
            <span class="text-sm">Admin</span>
          </label>
          <label class="block">
            <div class="mb-1 text-sm">Password</div>
            <input
              v-model="newPassword"
              type="password"
              placeholder="Leave blank to keep"
              class="w-full rounded border px-3 py-2"
            />
          </label>
          <label class="block">
            <div class="mb-1 text-sm">Group</div>
            <select v-model="editingGroupId" class="w-full rounded border px-3 py-2">
              <option :value="null">-</option>
              <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </label>
        </div>

        <div class="mt-4 flex justify-end gap-2">
          <Button variant="outline" @click="cancelForm">{{ t("global.cancel") }}</Button>
          <Button @click="saveUser">{{ t("global.save") }}</Button>
        </div>
      </div>
    </div>
  </div>
</template>
