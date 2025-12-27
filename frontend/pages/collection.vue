<script setup lang="ts">
  import { useI18n } from "vue-i18n";

  import { api, type User as MockUser, type Invite as MockInvite } from "~/mock/collections";

  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
  // Popover removed from invite UI; no longer importing
  import { Button, ButtonGroup } from "@/components/ui/button";
  import { Label } from "@/components/ui/label";
  import { Input } from "@/components/ui/input";
  import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select";
  import { Card } from "@/components/ui/card"; // Assuming you have a Card component
  import { Badge } from "@/components/ui/badge"; // Assuming you have a Badge component
  import { PlusCircle, Trash } from "lucide-vue-next"; // Icons
  import { format } from "date-fns";
  import CopyText from "@/components/global/CopyText.vue";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import { useDialog } from "~/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  const { openDialog } = useDialog();

  const { t } = useI18n();

  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "HomeBox | " + t("menu.maintenance"),
  });

  // Use centralized mock data / fake API
  const users = ref<MockUser[]>(api.getUsers());
  const invites = ref<MockInvite[]>(api.getInvites());

  // Current collection context (this page shows a single collection)
  // For now use the first mock collection as the active collection
  const currentCollectionId = api.getCollections()[0]?.id ?? "";

  // New invite email input
  // (declared below with invite inputs)

  // Settings state
  const collectionName = ref<string>("Personal Inventory");
  const saved = ref(false);

  // invite inputs (moved to dialog)

  const page = ref(1);

  const roles = ["owner", "admin", "editor", "viewer"];

  function inviteUrl(code: string) {
    if (typeof window === "undefined") return "";
    return `${window.location.origin}?token=${code}`;
  }

  function getMembershipRole(user: MockUser) {
    const mem = user.collections.find(c => c.id === currentCollectionId);
    return mem?.role ?? "viewer";
  }

  function roleVariant(role: string) {
    return role === "owner" ? "default" : role === "admin" ? "secondary" : "outline";
  }

  function handleRoleChange(userId: string, newRole: unknown) {
    // Update the role for this user specific to the current collection
    const roleStr = String(newRole || "viewer");
    api.addUserToCollection(userId, currentCollectionId, roleStr as MockUser["collections"][number]["role"]);
    users.value = api.getUsers();
  }

  function handleRemoveUser(userId: string) {
    api.deleteUser(userId);
    users.value = api.getUsers();
  }

  // Invite creation now handled by dialog component; keep helper removed.

  function deleteInvite(inviteId: string) {
    api.deleteInvite(inviteId);
    invites.value = api.getInvites();
  }

  function saveSettings() {
    // Stub: persist settings to API when implemented
    console.log("Saving collection settings", collectionName.value);
    saved.value = true;
    setTimeout(() => (saved.value = false), 2000);
  }
</script>

<template>
  <div>
    <BaseContainer class="flex flex-col gap-4">
      <BaseSectionHeader> Collection Settings </BaseSectionHeader>
      <ButtonGroup>
        <Button size="sm" :variant="page == 1 ? 'default' : 'outline'" @click="page = 1"> Users </Button>
        <Button size="sm" :variant="page == 2 ? 'default' : 'outline'" @click="page = 2"> Invites </Button>
        <Button size="sm" :variant="page == 3 ? 'default' : 'outline'" @click="page = 3"> Settings </Button>
      </ButtonGroup>

      <Card v-if="page == 1" class="">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Joined</TableHead>
              <TableHead></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="user in users" :key="user.id">
              <TableCell class="font-medium">
                {{ user.name }}
              </TableCell>
              <TableCell>
                <Select
                  :model-value="getMembershipRole(user)"
                  @update:model-value="newRole => handleRoleChange(user.id, newRole)"
                >
                  <SelectTrigger>
                    <span class="flex items-center">
                      <Badge class="whitespace-nowrap" :variant="roleVariant(getMembershipRole(user))">{{
                        getMembershipRole(user)
                      }}</Badge>
                    </span>
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="role in roles" :key="role" :value="role">
                      <div class="flex w-full items-center justify-between">
                        <Badge
                          class="whitespace-nowrap"
                          :variant="role === 'owner' ? 'default' : role === 'admin' ? 'secondary' : 'outline'"
                        >
                          {{ role }}
                        </Badge>
                      </div>
                    </SelectItem>
                  </SelectContent>
                </Select>
              </TableCell>
              <TableCell>
                {{ (user as any).created_at ? format(new Date((user as any).created_at), "PPP") : "-" }}
              </TableCell>
              <TableCell class="text-right">
                <div class="flex w-full items-center justify-end gap-2">
                  <Button variant="destructive" size="icon" @click="handleRemoveUser(user.id)">
                    <Trash class="size-4" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </Card>

      <Card v-if="page == 2" class="p-4">
        <div class="flex flex-col gap-4">
          <h3 class="text-lg font-semibold">Existing Invites</h3>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Code</TableHead>
                <TableHead>Expires</TableHead>
                <TableHead>Max Uses</TableHead>
                <TableHead>Uses</TableHead>
                <TableHead class="text-right"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="invite in invites" :key="invite.id">
                <TableCell class="font-medium">{{ invite.id }}</TableCell>
                <TableCell>{{ invite.expires_at ? format(new Date(invite.expires_at), "PPP") : "Never" }}</TableCell>
                <TableCell>{{ invite.max_uses ?? "∞" }}</TableCell>
                <TableCell>{{ invite.uses ?? 0 }}</TableCell>
                <TableCell class="w-max">
                  <div class="flex items-center justify-end gap-2">
                    <CopyText :text="inviteUrl(invite.id)" />
                    <Button variant="destructive" size="icon" @click="deleteInvite(invite.id)">
                      <Trash class="size-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>

          <hr class="my-4" />

          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">Create New Invite</h3>
            <div class="w-56">
              <Button
                class="w-full"
                @click="openDialog(DialogID.CreateInvite, { onClose: () => (invites.value = api.getInvites()) })"
              >
                <PlusCircle class="mr-2 size-4" /> Generate Invite
              </Button>
            </div>
          </div>
        </div>
      </Card>

      <Card v-if="page == 3" class="p-4">
        <h3 class="text-lg font-semibold">Collection Settings</h3>

        <div class="mt-4 grid items-end gap-4 md:grid-cols-2">
          <div class="flex flex-col gap-2">
            <Label for="collection-name">Name</Label>
            <Input id="collection-name" v-model="collectionName" placeholder="Collection name" />
          </div>

          <div class="flex items-end">
            <Button class="w-full" @click="saveSettings">Save</Button>
          </div>
        </div>

        <p v-if="saved" class="mt-3 text-sm text-green-600">Saved</p>
      </Card>
    </BaseContainer>
  </div>
</template>
