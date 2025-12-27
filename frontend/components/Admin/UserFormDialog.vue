<script setup lang="ts">
  import { reactive, ref, onMounted, onUnmounted } from "vue";
  import type { User as MockUser, Collection as MockCollection } from "~/mock/collections";
  import { api } from "~/mock/collections";
  import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogFooter,
    DialogTitle,
    DialogDescription,
  } from "@/components/ui/dialog";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { Button } from "@/components/ui/button";
  import { Input } from "@/components/ui/input";
  import { Badge } from "@/components/ui/badge";
  import { Select, SelectTrigger, SelectContent, SelectItem, SelectValue } from "@/components/ui/select";
  import MdiClose from "~icons/mdi/close";

  // dialog provider
  const { closeDialog, registerOpenDialogCallback } = useDialog();

  // local collections snapshot used for checkbox list
  const availableCollections = ref<MockCollection[]>(api.getCollections() as MockCollection[]);
  const isNew = ref(true);

  const localEditing = reactive<MockUser>({
    id: String(Date.now()),
    name: "",
    email: "",
    role: "user",
    password_set: false,
    collections: [],
  });

  const localCollectionIds = ref<string[]>([]);
  const localNewPassword = ref("");
  const newAddCollectionId = ref<string>("");

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.EditUser, params => {
      // refresh available collections each time
      availableCollections.value = api.getCollections() as MockCollection[];

      if (params && (params as { userId?: string }).userId) {
        const u = api.getUser(params.userId!);
        if (u) {
          Object.assign(localEditing, u as MockUser);
          localCollectionIds.value = (u.collections ?? []).map(c => c.id);
          isNew.value = false;
        } else {
          reset();
          isNew.value = true;
        }
      } else {
        // new user
        reset();
        isNew.value = true;
      }
      localNewPassword.value = "";
    });

    onUnmounted(cleanup);
  });

  type Membership = { id: string; role: "owner" | "admin" | "editor" | "viewer" };

  function getCollectionName(id: string) {
    const found = availableCollections.value.find(c => c.id === id);
    return found ? found.name : id;
  }

  // localEditing will be set when dialog opens via registerOpenDialogCallback

  function close() {
    reset();
    closeDialog(DialogID.EditUser);
  }

  function onSave() {
    if (!localEditing.name.trim() || !localEditing.email.trim()) {
      alert("Name and email are required");
      return;
    }

    if (localNewPassword.value && localEditing) localEditing.password_set = true;

    const existing = api.getUser(localEditing.id);
    if (existing) {
      const updated = {
        ...existing,
        name: localEditing.name,
        email: localEditing.email,
        role: localEditing.role,
        password_set: localEditing.password_set,
      } as MockUser;
      api.updateUser(updated);
    } else {
      const toCreate = { ...localEditing, collections: [] } as MockUser;
      const created = api.addUser(toCreate);
      localCollectionIds.value.forEach(id => api.addUserToCollection(created.id, id, "viewer"));
    }

    // close and signal caller to refresh
    closeDialog(DialogID.EditUser, true);
    reset();
  }

  function reset() {
    localEditing.id = String(Date.now());
    localEditing.name = "";
    localEditing.email = "";
    localEditing.role = "user";
    localEditing.password_set = false;
    localEditing.collections = [];
    localCollectionIds.value = [];
    localNewPassword.value = "";
  }

  function removeMembership(id: string) {
    const existing = (localEditing.collections ?? []) as Membership[];
    const found = existing.find((x: Membership) => x.id === id);
    if (found?.role === "owner") {
      const ok = confirm(
        `This user is the owner of this collection.\nRemoving the owner will delete the collection. Continue?`
      );
      if (!ok) return;
    }

    const existsInApi = !!api.getUser(localEditing.id);
    if (existsInApi) {
      const ok = api.removeUserFromCollection(localEditing.id, id);
      if (ok) {
        localCollectionIds.value = localCollectionIds.value.filter(x => x !== id);
        localEditing.collections = (localEditing.collections ?? []).filter((x: Membership) => x.id !== id);
        const refreshed = api.getUser(localEditing.id);
        if (refreshed) Object.assign(localEditing, refreshed as MockUser);
      }
      return;
    }

    // not in API yet — local only
    localCollectionIds.value = localCollectionIds.value.filter(x => x !== id);
    localEditing.collections = (localEditing.collections ?? []).filter((x: Membership) => x.id !== id);
  }

  function addMembership(id: string) {
    if (!id) return;
    const existsInApi = !!api.getUser(localEditing.id);
    if (existsInApi) {
      const mem = api.addUserToCollection(localEditing.id, id, "viewer");
      if (mem) {
        if (!localCollectionIds.value.includes(id)) localCollectionIds.value.push(id);
        localEditing.collections = (localEditing.collections ?? []).filter((x: Membership) => x.id !== id);
        localEditing.collections.push(mem as Membership);
        const refreshed = api.getUser(localEditing.id);
        if (refreshed) Object.assign(localEditing, refreshed as MockUser);
      }
      return;
    }

    // new user — add locally
    if (!localCollectionIds.value.includes(id)) localCollectionIds.value.push(id);
    localEditing.collections = (localEditing.collections ?? []).filter((x: Membership) => x.id !== id);
    localEditing.collections.push({ id, role: "viewer" });
  }

  function updateMembershipRole(id: string, role: Membership["role"]) {
    const existsInApi = !!api.getUser(localEditing.id);
    if (existsInApi) {
      // best-effort: remove then re-add with new role if API doesn't expose direct update
      api.removeUserFromCollection(localEditing.id, id);
      const mem = api.addUserToCollection(localEditing.id, id, role);
      if (mem) {
        const refreshed = api.getUser(localEditing.id);
        if (refreshed) Object.assign(localEditing, refreshed as MockUser);
        localCollectionIds.value = (localEditing.collections ?? []).map((c: Membership) => c.id);
      }
      return;
    }

    // local-only
    const existing = (localEditing.collections ?? []) as Membership[];
    const found = existing.find(x => x.id === id);
    if (found) found.role = role;
  }
</script>

<template>
  <Dialog :dialog-id="DialogID.EditUser">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ isNew ? "Add User" : "Edit User" }}</DialogTitle>
        <DialogDescription>Manage user details and collection memberships.</DialogDescription>
      </DialogHeader>

      <form class="flex flex-col gap-3" @submit.prevent="onSave">
        <label class="block">
          <div class="mb-1 text-sm">Name</div>
          <Input v-model="localEditing.name" />
        </label>

        <label class="block">
          <div class="mb-1 text-sm">Email</div>
          <Input v-model="localEditing.email" />
        </label>

        <label class="block">
          <div class="mb-1 text-sm">Password</div>
          <Input v-model="localNewPassword" type="password" placeholder="Leave blank to keep" />
        </label>

        <div>
          <div class="mb-1 text-sm">Collections</div>
          <div class="flex flex-col gap-3">
            <div
              v-for="m in localEditing.collections ?? []"
              :key="m.id"
              class="flex items-center justify-between rounded-lg border py-1 pl-3 pr-1"
            >
              <div class="text-lg font-medium">
                <Badge>
                  {{ getCollectionName(m.id) }}
                </Badge>
              </div>
              <div class="flex items-center gap-3">
                <Select v-model="m.role" @update:model-value="val => updateMembershipRole(m.id, val)">
                  <SelectTrigger class="w-40">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="owner">Owner</SelectItem>
                    <SelectItem value="admin">Admin</SelectItem>
                    <SelectItem value="editor">Editor</SelectItem>
                    <SelectItem value="viewer">Viewer</SelectItem>
                  </SelectContent>
                </Select>
                <Button
                  variant="destructive"
                  size="icon"
                  class="ml-2"
                  :title="$t ? $t('global.remove') : 'Remove'"
                  @click.prevent="removeMembership(m.id)"
                >
                  <MdiClose class="size-4" />
                </Button>
              </div>
            </div>

            <div class="mt-2 flex items-center gap-2">
              <Select v-model="newAddCollectionId">
                <SelectTrigger class="flex-1">
                  <SelectValue placeholder="Select collection to add" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="c in availableCollections.filter(c => !localCollectionIds.includes(c.id))"
                    :key="c.id"
                    :value="c.id"
                  >
                    {{ c.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <Button
                type="button"
                class="ml-2 w-10 px-0"
                variant="default"
                size="lg"
                :disabled="!newAddCollectionId"
                @click="addMembership(newAddCollectionId)"
              >
                +
              </Button>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" type="button" @click="close">Cancel</Button>
          <Button type="submit">Save</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
