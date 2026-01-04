<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import BaseContainer from "@/components/Base/Container.vue";
  import { Card } from "@/components/ui/card";
  import { Button, ButtonGroup } from "@/components/ui/button";

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
    },
    {
      id: "invites",
      label: "collection.tabs.invites",
      to: "/collection/invites",
    },
    {
      id: "settings",
      label: "collection.tabs.settings",
      to: "/collection/settings",
    },
    {
      id: "tools",
      label: "collection.tabs.tools",
      to: "/collection/tools",
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

      <div class="my-3 flex flex-wrap items-center justify-between">
        <ButtonGroup>
          <Button
            v-for="tab in tabs"
            :key="tab.id"
            as-child
            :variant="tab.to === currentPath ? 'default' : 'outline'"
            size="sm"
          >
            <NuxtLink :to="tab.to">
              {{ t(tab.label) }}
            </NuxtLink>
          </Button>
        </ButtonGroup>
      </div>
    </section>

    <section>
      <div class="space-y-6">
        <NuxtPage />
      </div>
    </section>
  </BaseContainer>
</template>
