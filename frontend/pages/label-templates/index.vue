<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import MdiPlus from "~icons/mdi/plus";
  import MdiPrinter from "~icons/mdi/printer";
  import { Button } from "@/components/ui/button";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import LabelTemplateCard from "~/components/LabelTemplate/Card.vue";
  import LabelTemplateCreateModal from "~/components/LabelTemplate/CreateModal.vue";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

  useHead({
    title: computed(() => `HomeBox | ${t("pages.label_templates.title")}`),
  });

  const { templates, refresh, pending } = useLabelTemplates();
  const { openDialog } = useDialog();

  const handleRefresh = () => refresh();
</script>

<template>
  <BaseContainer>
    <div class="mb-4 flex justify-between">
      <BaseSectionHeader>{{ $t("pages.label_templates.title") }}</BaseSectionHeader>
      <div class="flex gap-2">
        <Button variant="outline" as-child>
          <NuxtLink to="/label-templates/batch">
            <MdiPrinter class="mr-2" />
            {{ $t("pages.label_templates.batch.title") }}
          </NuxtLink>
        </Button>
        <Button @click="openDialog(DialogID.CreateLabelTemplate)">
          <MdiPlus class="mr-2" />
          {{ $t("global.create") }}
        </Button>
      </div>
    </div>

    <LabelTemplateCreateModal @created="handleRefresh" />

    <div v-if="pending" class="flex items-center justify-center py-12">
      <div class="text-muted-foreground">{{ $t("global.loading") }}</div>
    </div>

    <div v-else-if="templates && templates.length > 0" class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <LabelTemplateCard
        v-for="tpl in templates"
        :key="tpl.id"
        :template="tpl"
        @deleted="handleRefresh"
        @duplicated="handleRefresh"
      />
    </div>

    <div v-else class="flex flex-col items-center justify-center py-12 text-center">
      <p class="mb-4 text-muted-foreground">{{ $t("pages.label_templates.no_templates") }}</p>
      <Button @click="openDialog(DialogID.CreateLabelTemplate)">
        <MdiPlus class="mr-2" />
        {{ $t("components.label_template.create_modal.title") }}
      </Button>
    </div>
  </BaseContainer>
</template>
