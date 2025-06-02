<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { v4 as uuidv4 } from 'uuid'; // For generating unique invite IDs

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Card } from '@/components/ui/card'; // Assuming you have a Card component
import { Badge } from '@/components/ui/badge'; // Assuming you have a Badge component
import {
  Calendar as CalendarIcon,
  PlusCircle,
  Trash,
} from 'lucide-vue-next'; // Icons
import { Calendar } from '@/components/ui/calendar';
import { format } from 'date-fns';

const { t } = useI18n();

definePageMeta({
  middleware: ['auth'],
});
useHead({
  title: 'HomeBox | ' + t('menu.maintenance'),
});

interface User {
  username: string;
  id: string;
  role: 'owner' | 'admin' | 'editor' | 'viewer';
  lastActive: string;
  added: string;
}

interface Invite {
  id: string;
  code: string;
  expiresAt: Date | null;
  maxUses: number | null;
  uses: number;
}

const users = ref<User[]>([
  {
    username: 'tonya',
    id: '1',
    role: 'owner',
    lastActive: '12 hours ago',
    added: '13 hours ago',
  },
  {
    username: 'steve',
    id: '2',
    role: 'admin',
    lastActive: '1 day ago',
    added: '2 days ago',
  },
  {
    username: 'bob',
    id: '3',
    role: 'editor',
    lastActive: '30 minutes ago',
    added: '5 hours ago',
  },
  {
    username: 'john',
    id: '4',
    role: 'viewer',
    lastActive: '2 hours ago',
    added: '1 day ago',
  },
]);

const invites = ref<Invite[]>([
  {
    id: uuidv4(),
    code: 'ABCDEF',
    expiresAt: null,
    maxUses: null,
    uses: 0,
  },
  {
    id: uuidv4(),
    code: 'GHIJKL',
    expiresAt: new Date(new Date().setDate(new Date().getDate() + 7)), // Expires in 7 days
    maxUses: 5,
    uses: 2,
  },
]);

const newInviteExpiresAt = ref<Date | null>(null);
const newInviteMaxUses = ref<number | null>(null);

const page = ref(1);

const roles = ['owner', 'admin', 'editor', 'viewer'];

function handleRoleChange(userId: string, newRole: string) {
  const userIndex = users.value.findIndex((user) => user.id === userId);
  if (userIndex !== -1) {
    users.value[userIndex].role = newRole as
      | 'owner'
      | 'admin'
      | 'editor'
      | 'viewer';
  }
}

function handleRemoveUser(userId: string) {
  users.value = users.value.filter((user) => user.id !== userId);
}

function generateInviteCode() {
  return Math.random().toString(36).substring(2, 8).toUpperCase();
}

function createNewInvite() {
  const newInvite: Invite = {
    id: uuidv4(),
    code: generateInviteCode(),
    expiresAt: newInviteExpiresAt.value,
    maxUses: newInviteMaxUses.value,
    uses: 0,
  };
  invites.value.push(newInvite);
  newInviteExpiresAt.value = null;
  newInviteMaxUses.value = null;
}

function deleteInvite(inviteId: string) {
  invites.value = invites.value.filter((invite) => invite.id !== inviteId);
}
</script>

<template>
  <div>
    <BaseContainer class="flex flex-col gap-4">
      <BaseSectionHeader> Collection Settings </BaseSectionHeader>
      <ButtonGroup>
        <Button
          size="sm"
          :variant="page == 1 ? 'default' : 'outline'"
          @click="page = 1"
        >
          Users
        </Button>
        <Button
          size="sm"
          :variant="page == 2 ? 'default' : 'outline'"
          @click="page = 2"
        >
          Invites
        </Button>
        <Button
          size="sm"
          :variant="page == 3 ? 'default' : 'outline'"
          @click="page = 3"
        >
          Settings
        </Button>
      </ButtonGroup>

      <Card v-if="page == 1" class="p-4 m-4">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Username</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Last Active</TableHead>
              <TableHead>Added</TableHead>
              <TableHead class="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="user in users" :key="user.id">
              <TableCell class="font-medium">
                {{ user.username }}
              </TableCell>
              <TableCell>
                <Badge
                  :variant="
                    user.role === 'owner'
                      ? 'default'
                      : user.role === 'admin'
                        ? 'secondary'
                        : 'outline'
                  "
                >
                  {{ user.role }}
                </Badge>
              </TableCell>
              <TableCell>{{ user.lastActive }}</TableCell>
              <TableCell>{{ user.added }}</TableCell>
              <TableCell class="text-right">
                <Popover>
                  <PopoverTrigger as-child>
                    <Button size="sm" variant="outline"> Edit </Button>
                  </PopoverTrigger>
                  <PopoverContent class="w-48">
                    <div class="grid gap-4">
                      <div class="space-y-2">
                        <h4 class="font-medium leading-none">Edit User</h4>
                        <p class="text-sm text-muted-foreground">
                          {{ user.username }}
                        </p>
                      </div>
                      <div class="grid gap-2">
                        <Label for="role">Role</Label>
                        <Select
                          :model-value="user.role"
                          @update:model-value="
                            (newRole) => handleRoleChange(user.id, newRole)
                          "
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Select a role" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem
                              v-for="role in roles"
                              :key="role"
                              :value="role"
                            >
                              {{ role }}
                            </SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                      <Button
                        variant="destructive"
                        size="sm"
                        @click="handleRemoveUser(user.id)"
                      >
                        Remove User
                      </Button>
                    </div>
                  </PopoverContent>
                </Popover>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </Card>

      <Card v-if="page == 2" class="p-4 m-4">
        <div class="flex flex-col gap-4">
          <h3 class="text-lg font-semibold">Existing Invites</h3>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Code</TableHead>
                <TableHead>Expires</TableHead>
                <TableHead>Max Uses</TableHead>
                <TableHead>Uses</TableHead>
                <TableHead class="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="invite in invites" :key="invite.id">
                <TableCell class="font-medium">
                  {{ invite.code }}
                </TableCell>
                <TableCell>
                  {{
                    invite.expiresAt
                      ? format(invite.expiresAt, 'PPP')
                      : 'Never'
                  }}
                </TableCell>
                <TableCell>
                  {{ invite.maxUses !== null ? invite.maxUses : 'Unlimited' }}
                </TableCell>
                <TableCell>{{ invite.uses }}</TableCell>
                <TableCell class="text-right">
                  <Button
                    variant="destructive"
                    size="icon"
                    @click="deleteInvite(invite.id)"
                  >
                    <Trash class="w-4 h-4" />
                  </Button>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>

          <hr class="my-4" />

          <h3 class="text-lg font-semibold">Create New Invite</h3>
          <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <div class="flex flex-col gap-2">
              <Label for="new-invite-max-uses">Max Uses (optional)</Label>
              <Input
                id="new-invite-max-uses"
                type="number"
                v-model.number="newInviteMaxUses"
                placeholder="Unlimited"
              />
            </div>
            <div class="flex flex-col gap-2">
              <Label for="new-invite-expires-at">Expires At (optional)</Label>
              <Popover>
                <PopoverTrigger as-child>
                  <Button
                    variant="outline"
                    class="w-full justify-start text-left font-normal"
                    :class="
                      !newInviteExpiresAt && 'text-muted-foreground'
                    "
                  >
                    <CalendarIcon class="mr-2 h-4 w-4" />
                    {{
                      newInviteExpiresAt
                        ? format(newInviteExpiresAt, 'PPP')
                        : 'Pick a date'
                    }}
                  </Button>
                </PopoverTrigger>
                <PopoverContent class="w-auto p-0">
                  <Calendar v-model:model-value="newInviteExpiresAt" />
                </PopoverContent>
              </Popover>
            </div>
            <div class="flex items-end">
              <Button @click="createNewInvite" class="w-full">
                <PlusCircle class="mr-2 w-4 h-4" /> Generate Invite
              </Button>
            </div>
          </div>
        </div>
      </Card>

      <Card v-if="page == 3" class="p-4 m-4">
        <h3 class="text-lg font-semibold">Collection Settings</h3>
        <p class="text-muted-foreground">
          This is where you would configure general collection settings.
        </p>
        <!-- Add your settings forms/components here -->
      </Card>
    </BaseContainer>
  </div>
</template>