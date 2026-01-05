<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import BaseContainer from "@/components/Base/Container.vue";
  import { Card } from "@/components/ui/card";
  import { Button, ButtonGroup } from "@/components/ui/button";

  import MdiAccountMultiple from "~icons/mdi/account-multiple";
  import MdiEmailPlus from "~icons/mdi/email-plus";
  import MdiBell from "~icons/mdi/bell";
  import MdiCog from "~icons/mdi/cog";
  import MdiWrench from "~icons/mdi/wrench";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

  const route = useRoute();

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

  const { selectedCollection } = useCollections();
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
              <span>{{ t(tab.label) }}</span>
            </NuxtLink>
          </Button>
        </ButtonGroup>

        <div id="collection-header-actions" class="ml-auto flex items-center gap-2" />
      </div>
    </section>

    <section>
      <div class="space-y-6">
        <NuxtPage />
      </div>
    </section>
  </BaseContainer>
</template>
