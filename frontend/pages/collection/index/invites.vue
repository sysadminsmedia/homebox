<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
  import { Button } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import MdiPlus from "~icons/mdi/plus";
  import MdiDelete from "~icons/mdi/delete";
  import { toast } from "@/components/ui/sonner";
  import { useUserApi } from "~/composables/use-api";
  import { useDialog } from "~/components/ui/dialog-provider";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import CopyText from "@/components/global/CopyText.vue";
  import type { Group, GroupInvitation } from "~~/lib/api/types/data-contracts";

  definePageMeta({
    middleware: ["auth"],
  });

  type Invitation = GroupInvitation & { id: string; group?: Group };
  type InvitationResult = GroupInvitation & { id: string; group?: Group };

  const { t } = useI18n();

  useHead({ title: `HomeBox | ${t("collection.tabs.invites")}` });

  const api = useUserApi();
  const { openDialog } = useDialog();
  const confirm = useConfirm();
  const localInvites = ref<Invitation[]>([]);
  const baseUrl = `${window.location.protocol}//${window.location.host}`;
  const loading = ref(true);
  const invites = ref<Invitation[]>([]);
  const error = ref<string | null>(null);
  const removing = ref<Record<string, boolean>>({});

  const allInvites = computed<Invitation[]>(() => {
    const now = Date.now();

    const isActive = (inv: Invitation) => {
      if (!inv.expiresAt) return true;

      const expiresAtTime = new Date(inv.expiresAt).getTime();
      return Number.isFinite(expiresAtTime) && expiresAtTime > now;
    };

    return [...localInvites.value.filter(isActive), ...invites.value.filter(isActive)];
  });

  const loadInvites = async () => {
    loading.value = true;
    error.value = null;

    try {
      const res = await api.group.getInvitations();

      if (res.error) {
        const msg = t("errors.api_failure") + String(res.error);
        error.value = msg;
        invites.value = [];
        toast.error(msg);
      } else {
        invites.value = (res.data || []) as Invitation[];
        console.log("Loaded invites:", invites.value);
      }
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      error.value = msg;
      invites.value = [];
      toast.error(msg);
    } finally {
      loading.value = false;
    }
  };

  const handleOpenCreate = () => {
    openDialog(DialogID.CreateGroupInvite, {
      onClose: (result?: InvitationResult) => {
        if (!result) return;

        console.log("Created invite:", result);

        const localInvite: Invitation = {
          ...(result as InvitationResult),
        };

        localInvites.value = [localInvite, ...localInvites.value];
      },
    });
  };

  const handleDelete = async (inv: Invitation) => {
    if (!inv?.id) return;

    const result = await confirm.open(t("collection.invites_delete_confirm"));
    if (result.isCanceled) {
      return;
    }

    removing.value = { ...removing.value, [inv.id]: true };

    try {
      const res = await api.group.deleteInvitation(inv.id);

      if (res.error) {
        const msg = t("errors.api_failure") + String(res.error);
        toast.error(msg);
      } else {
        invites.value = invites.value.filter(i => i.id !== inv.id);
        localInvites.value = localInvites.value.filter(i => i.id !== inv.id && i.token !== inv.token);
      }
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      toast.error(msg);
    } finally {
      removing.value = { ...removing.value, [inv.id]: false };
    }
  };

  onMounted(() => {
    loadInvites();
  });
</script>

<template>
  <div class="space-y-4">
    <Teleport to="#collection-header-actions" defer>
      <TooltipProvider :delay-duration="0">
        <Tooltip>
          <TooltipTrigger as-child>
            <Button
              class="size-8"
              variant="outline"
              size="icon"
              :aria-label="$t('collection.create_invite')"
              :disabled="loading"
              @click="handleOpenCreate"
            >
              <MdiPlus class="size-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            {{ $t("collection.create_invite") }}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </Teleport>

    <div v-if="loading" class="rounded-md border bg-card p-4 text-sm text-muted-foreground">
      {{ $t("global.loading") }}
    </div>

    <div v-else class="space-y-3">
      <div v-if="localInvites.length" class="rounded-md bg-secondary p-3 text-xs font-bold text-secondary-foreground">
        {{ $t("collection.invite_token_warning") }}
      </div>

      <div v-if="!allInvites.length" class="rounded-md border bg-card p-4 text-sm text-muted-foreground">
        {{ $t("collection.no_invites") }}
      </div>

      <div v-else class="scroll-bg overflow-x-auto rounded-md border bg-card">
        <Table class="min-w-[640px]">
          <TableHeader>
            <TableRow>
              <TableHead>{{ $t("collection.invite_token") }}</TableHead>
              <TableHead>{{ $t("collection.expires_at") }}</TableHead>
              <TableHead>{{ $t("collection.remaining_uses") }}</TableHead>
              <TableHead class="w-40 text-right"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="inv in allInvites" :key="inv.id || inv.token">
              <TableCell>
                <span class="break-all font-mono text-xs">
                  <template v-if="inv.token">
                    {{ inv.token }}
                  </template>
                  <span v-else class="font-sans text-muted-foreground">
                    {{ $t("collection.invite_token_hidden") }}
                  </span>
                </span>
              </TableCell>
              <TableCell>
                {{ new Date(inv.expiresAt).toLocaleString() }}
              </TableCell>
              <TableCell>
                {{ inv.uses }}
              </TableCell>
              <TableCell>
                <div class="ml-auto flex items-center gap-2">
                  <CopyText
                    v-if="inv.token"
                    :text="`${baseUrl}?token=${inv.token}`"
                    :icon-size="16"
                    :tooltip="$t('collection.copy_invite')"
                  />
                  <TooltipProvider :delay-duration="0">
                    <Tooltip>
                      <TooltipTrigger as-child>
                        <Button
                          variant="destructive"
                          size="icon"
                          :aria-label="$t('global.delete')"
                          :disabled="removing[inv.id]"
                          @click="handleDelete(inv)"
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
