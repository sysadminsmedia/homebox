<script setup lang="ts">
  import { Button } from "@/components/ui/button";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseCard from "@/components/Base/Card.vue";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "@/components/ui/dialog-provider/utils";

  definePageMeta({
    middleware: ["auth"],
  });

  const { openDialog } = useDialog();
</script>

<template>
  <BaseContainer class="flex justify-center">
    <div class="w-full">
      <BaseCard>
        <template #title>
          <h2 class="text-center text-xl font-semibold tracking-tight">
            {{ $t("collection.no_collections.title") }}
          </h2>
        </template>

        <div class="mx-4 mb-4 flex flex-col gap-4 text-sm text-muted-foreground">
          <p class="text-center text-base text-foreground">
            {{ $t("collection.no_collections.message") }}
          </p>

          <div class="flex flex-col gap-2 sm:flex-row sm:justify-center">
            <Button
              class="w-full sm:w-auto"
              @click="
                openDialog(DialogID.CreateCollection, {
                  params: {
                    redirectTo: '/home',
                  },
                })
              "
            >
              {{ $t("collection.no_collections.create") }}
            </Button>
            <Button
              class="w-full sm:w-auto"
              variant="secondary"
              @click="
                openDialog(DialogID.JoinCollection, {
                  params: {
                    redirectTo: '/home',
                  },
                })
              "
            >
              {{ $t("collection.no_collections.join") }}
            </Button>
          </div>
        </div>
      </BaseCard>
    </div>
  </BaseContainer>
</template>
