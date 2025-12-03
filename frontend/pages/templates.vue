<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import MdiPlus from "~icons/mdi/plus";
  import { Button } from "@/components/ui/button";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import TemplateCard from "~/components/Template/Card.vue";
  import TemplateCreateModal from "~/components/Template/CreateModal.vue";

  definePageMeta({
    middleware: ["auth"],
  });

  useHead({
    title: "HomeBox | Templates",
  });

  const api = useUserApi();
  const { openDialog } = useDialog();

  const { data: templates, refresh } = useAsyncData("templates", async () => {
    const { data, error } = await api.templates.getAll();
    if (error) {
      toast.error("Failed to load templates");
      return [];
    }
    return data;
  });
</script>

<template>
  <BaseContainer>
    <div class="mb-4 flex justify-between">
      <BaseSectionHeader>Templates</BaseSectionHeader>
      <Button @click="openDialog(DialogID.CreateTemplate)">
        <MdiPlus class="mr-2" />
        {{ $t("global.create") }}
      </Button>
    </div>

    <TemplateCreateModal @created="refresh" />

    <div v-if="templates && templates.length > 0" class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <TemplateCard v-for="tpl in templates" :key="tpl.id" :template="tpl" @deleted="refresh" @duplicated="refresh" />
    </div>

    <div v-else class="flex flex-col items-center justify-center py-12 text-center">
      <p class="mb-4 text-muted-foreground">No templates yet.</p>
      <Button @click="openDialog(DialogID.CreateTemplate)">
        <MdiPlus class="mr-2" />
        Create Template
      </Button>
    </div>
  </BaseContainer>
</template>
