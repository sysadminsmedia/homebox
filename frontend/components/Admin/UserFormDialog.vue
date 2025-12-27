<script setup lang="ts">
  import { reactive, watch, ref } from "vue";
  import type { User as MockUser, Collection as MockCollection } from "~/mock/collections";
  import { api } from "~/mock/collections";
  import { DialogContent, DialogHeader, DialogFooter, DialogTitle, DialogDescription } from "@/components/ui/dialog";
  import { Button } from "@/components/ui/button";
  import { Input } from "@/components/ui/input";
  import { Badge } from "@/components/ui/badge";

  const props = defineProps<{
    modelValue: boolean;
    // editing matches the mock User shape: collections are {id, role} tuples
    editing: MockUser | null;
    collections: MockCollection[];
    editingCollectionIds: string[];
    newPassword: string;
  }>();

  const emit = defineEmits([
    "update:modelValue",
    "update:editing",
    "update:editingCollectionIds",
    "update:newPassword",
    "save",
    "cancel",
    "collections-changed",
  ] as const);

  const localEditing = reactive<MockUser>(
    props.editing
      ? { ...props.editing }
      : { id: String(Date.now()), name: "", email: "", role: "user", password_set: false, collections: [] }
  );

  const localCollectionIds = ref<string[]>([...(props.editingCollectionIds ?? [])]);
  const localNewPassword = ref(props.newPassword ?? "");

  type Membership = { id: string; role: "owner" | "admin" | "editor" | "viewer" };

  watch(
    () => props.editing,
    v => {
      if (v) Object.assign(localEditing, v);
      else {
        localEditing.id = String(Date.now());
        localEditing.name = "";
        localEditing.email = "";
        localEditing.role = "user";
        localEditing.password_set = false;
        localEditing.collections = [];
      }
    },
    { immediate: true }
  );

  watch(
    () => props.editingCollectionIds,
    v => {
      localCollectionIds.value = [...(v ?? [])];
    },
    { immediate: true }
  );

  watch(
    () => props.newPassword,
    v => (localNewPassword.value = v ?? ""),
    { immediate: true }
  );

  function close() {
    emit("update:modelValue", false);
    emit("cancel");
  }

  function onSave() {
    // propagate changes back to parent
    emit("update:editing", { ...localEditing });
    emit("update:editingCollectionIds", [...localCollectionIds.value]);
    emit("update:newPassword", localNewPassword.value);
    emit("save");
    emit("update:modelValue", false);
  }

  function onCheckboxChange(e: Event, id: string) {
    const checked = (e.target as HTMLInputElement).checked;
    // if this user exists in the api, call the api to add/remove membership immediately
    const existsInApi = !!api.getUser(localEditing.id);
    if (!checked) {
      // unchecking
      if (existsInApi) {
        // will confirm inside removeMembership
        const ok = api.removeUserFromCollection(localEditing.id, id);
        if (ok) {
          // update local state to reflect api change
          localCollectionIds.value = localCollectionIds.value.filter(x => x !== id);
          localEditing.collections = (localEditing.collections ?? []).filter((x: Membership) => x.id !== id);
        }
        return;
      }
      // not in API yet (new user) — just update local state
      localCollectionIds.value = localCollectionIds.value.filter(x => x !== id);
      localEditing.collections = (localEditing.collections ?? []).filter((x: Membership) => x.id !== id);
      return;
    }

    // checking
    if (existsInApi) {
      const mem = api.addUserToCollection(localEditing.id, id, "viewer");
      if (mem) {
        if (!localCollectionIds.value.includes(id)) localCollectionIds.value.push(id);
        localEditing.collections = (localEditing.collections ?? []).filter((x: Membership) => x.id !== id);
        localEditing.collections.push(mem as Membership);
        emit("update:editing", { ...api.getUser(localEditing.id) });
        emit("update:editingCollectionIds", [...localCollectionIds.value]);
        emit("collections-changed");
      }
      return;
    }

    // new user — just add locally
    if (!localCollectionIds.value.includes(id)) localCollectionIds.value.push(id);
    localEditing.collections = (localEditing.collections ?? []).filter((x: Membership) => x.id !== id);
    localEditing.collections.push({ id, role: "viewer" });
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
        emit("update:editing", { ...api.getUser(localEditing.id) });
        emit("update:editingCollectionIds", [...localCollectionIds.value]);
        emit("collections-changed");
      }
      return;
    }

    // not in API yet — local only
    localCollectionIds.value = localCollectionIds.value.filter(x => x !== id);
    localEditing.collections = (localEditing.collections ?? []).filter((x: Membership) => x.id !== id);
  }
</script>

<template>
  <DialogContent v-if="props.modelValue">
    <DialogHeader>
      <DialogTitle>{{ props.editing ? "Edit User" : "Add User" }}</DialogTitle>
      <DialogDescription> Manage user details and collection memberships. </DialogDescription>
    </DialogHeader>

    <div class="grid gap-3">
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
        <div class="flex flex-wrap gap-2">
          <label v-for="c in props.collections" :key="c.id" class="inline-flex items-center gap-2">
            <input
              type="checkbox"
              :value="c.id"
              :checked="localCollectionIds.includes(c.id)"
              @change="onCheckboxChange($event, c.id)"
            />
            <Badge class="whitespace-nowrap">{{ c.name }}</Badge>
            <button type="button" class="ml-2 text-destructive" @click.prevent="removeMembership(c.id)">×</button>
          </label>
        </div>
      </div>
    </div>

    <DialogFooter>
      <Button variant="outline" @click="close">Cancel</Button>
      <Button @click="onSave">Save</Button>
    </DialogFooter>
  </DialogContent>
</template>
