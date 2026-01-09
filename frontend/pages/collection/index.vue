<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import BaseContainer from "@/components/Base/Container.vue";
  import { Card } from "@/components/ui/card";
  import { Button, ButtonGroup } from "@/components/ui/button";
  import { toast } from "@/components/ui/sonner";

  import MdiAccountMultiple from "~icons/mdi/account-multiple";
  import MdiEmailPlus from "~icons/mdi/email-plus";
  import MdiBell from "~icons/mdi/bell";
  import MdiCog from "~icons/mdi/cog";
  import MdiWrench from "~icons/mdi/wrench";
  import MdiLogout from "~icons/mdi/logout";
  import MdiDelete from "~icons/mdi/delete";
  import type { UserUpdate } from "~/lib/api/types/data-contracts";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

  const route = useRoute();
  const api = useUserApi();
  const auth = useAuthContext();
  const confirm = useConfirm();

  const currentPath = computed(() => route.path);

  const tabs = computed(() => [
    {
      id: "members",
      label: "collection.tabs.members",
      to: "/collection/members",
      icon: MdiAccountMultiple,
    },
    {
      id: "invites",
      label: "collection.tabs.invites",
      to: "/collection/invites",
      icon: MdiEmailPlus,
    },
    {
      id: "notifiers",
      label: "collection.tabs.notifiers",
      to: "/collection/notifiers",
      icon: MdiBell,
    },
    {
      id: "settings",
      label: "collection.tabs.settings",
      to: "/collection/settings",
      icon: MdiCog,
    },
    {
      id: "tools",
      label: "collection.tabs.tools",
      to: "/collection/tools",
      icon: MdiWrench,
    },
  ]);

  const { selectedCollection, load: reloadCollections } = useCollections();

  const members = ref<Array<UserUpdate>>([]);
  const membersLoading = ref(false);
  const actionLoading = ref(false);

  const currentUserId = computed(() => auth.user?.id ?? "");

  const membersCount = computed(() => members.value.length);

  const isOnlyMember = computed(() => membersCount.value === 1 && members.value[0]?.id === currentUserId.value);
  const isActionDisabled = computed(() => !selectedCollection.value || membersLoading.value || actionLoading.value);

  const loadMembers = async () => {
    if (!selectedCollection.value) {
      members.value = [];
      return;
    }

    membersLoading.value = true;
    try {
      const res = await api.group.getMembers();
      if (res.error) {
        const msg = t("errors.api_failure") + String(res.error);
        toast.error(msg);
        members.value = [];
      } else {
        members.value = Array.isArray(res.data) ? (res.data as Array<{ id: string }>) : [];
      }
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      toast.error(msg);
      members.value = [];
    } finally {
      membersLoading.value = false;
    }
  };

  watch(
    () => selectedCollection.value?.id,
    () => {
      void loadMembers();
    },
    { immediate: true }
  );

  const handleLeaveCollection = async () => {
    if (!selectedCollection.value) return;

    const result = await confirm.open(t("collection.leave_confirm"));
    if (result.isCanceled) {
      return;
    }

    actionLoading.value = true;

    try {
      let userId = currentUserId.value;
      if (!userId) {
        const { data } = await api.user.self();
        userId = data?.item.id ?? "";
      }

      if (!userId) {
        const msg = t("errors.api_failure") + "Missing user id";
        toast.error(msg);
        return;
      }

      const res = await api.group.removeMember(userId);
      if (res.error) {
        const msg = t("errors.api_failure") + String(res.error);
        toast.error(msg);
        return;
      }

      toast.success(t("collection.left_collection"));
      await reloadCollections();
      window.location.reload();
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      toast.error(msg);
    } finally {
      actionLoading.value = false;
    }
  };

  const handleDeleteCollection = async () => {
    if (!selectedCollection.value) return;

    const result = await confirm.open(t("collection.delete_confirm"));
    if (result.isCanceled) {
      return;
    }

    actionLoading.value = true;

    try {
      const res = await api.group.delete(selectedCollection.value.id);
      if (res.error) {
        const msg = t("errors.api_failure") + String(res.error);
        toast.error(msg);
        return;
      }

      toast.success(t("collection.deleted_collection"));
      await reloadCollections();
      window.location.reload();
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      toast.error(msg);
    } finally {
      actionLoading.value = false;
    }
  };

  const handleCollectionPrimaryAction = async () => {
    if (!selectedCollection.value) return;

    if (isOnlyMember.value) {
      await handleDeleteCollection();
    } else {
      await handleLeaveCollection();
    }
  };
</script>

<template>
  <BaseContainer>
    <Title>{{ t("menu.collection_options") }}</Title>

    <section>
      <Card class="p-3">
        <header>
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div>
              <h1 class="text-2xl">
                {{
                  t("collection.manage_collection") + " - " + selectedCollection?.name ||
                  t("components.collection.selector.select_collection")
                }}
              </h1>
            </div>
          </div>
        </header>
      </Card>

      <div class="my-3 flex flex-wrap items-center justify-between gap-2">
        <ButtonGroup>
          <Button
            v-for="tab in tabs"
            :key="tab.id"
            as-child
            :variant="tab.to === currentPath ? 'default' : 'outline'"
            size="sm"
          >
            <NuxtLink :to="tab.to" class="flex items-center gap-2">
              <component :is="tab.icon" v-if="tab.icon" class="size-4" />
              <span class="hidden sm:block">{{ t(tab.label) }}</span>
            </NuxtLink>
          </Button>
        </ButtonGroup>

        <div id="collection-header-actions" class="ml-auto flex items-center gap-1">
          <Button
            variant="outline"
            size="icon"
            class="size-8"
            :aria-label="$t(isOnlyMember ? 'collection.delete_collection' : 'collection.leave_collection')"
            :disabled="isActionDisabled"
            @click="handleCollectionPrimaryAction"
          >
            <component :is="isOnlyMember ? MdiDelete : MdiLogout" class="size-4" />
          </Button>
        </div>
      </div>
    </section>

    <section>
      <div class="space-y-6">
        <NuxtPage />
      </div>
    </section>
  </BaseContainer>
</template>
