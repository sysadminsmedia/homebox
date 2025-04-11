<template>
  <div>
    <AppImportDialog />
    <BaseContainer class="mb-6 flex flex-col gap-4">
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiFileChart class="mr-2" />
            <span> {{ $t("tools.reports") }} </span>
            <template #description> {{ $t("tools.reports_sub") }} </template>
          </BaseSectionHeader>
        </template>
        <div class="divide-y divide-gray-300 border-t border-gray-300 px-6 pb-3">
          <DetailAction href="/reports/label-generator">
            <template #title>{{ $t("tools.reports_set.asset_labels") }}</template>
            {{ $t("tools.reports_set.asset_labels_sub") }}
            <template #button>
              {{ $t("tools.reports_set.asset_labels_button") }}
              <MdiArrowRight class="ml-2" />
            </template>
          </DetailAction>
          <DetailAction @action="getBillOfMaterials()">
            <template #title>{{ $t("tools.reports_set.bill_of_materials") }}</template>
            {{ $t("tools.reports_set.bill_of_materials_sub") }}
            <template #button> {{ $t("tools.reports_set.bill_of_materials_button") }} </template>
          </DetailAction>
        </div>
      </BaseCard>
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiDatabase class="mr-2" />
            <span> {{ $t("tools.import_export") }} </span>
            <template #description>
              {{ $t("tools.import_export_sub") }}
            </template>
          </BaseSectionHeader>
        </template>
        <div class="divide-y divide-gray-300 border-t border-gray-300 px-6 pb-3">
          <DetailAction @action="openDialog('import')">
            <template #title> {{ $t("tools.import_export_set.import") }} </template>
            <!-- eslint-disable-next-line vue/no-v-html -->
            <div v-html="DOMPurify.sanitize($t('tools.import_export_set.import_sub'))"></div>
            <template #button> {{ $t("tools.import_export_set.import_button") }} </template>
          </DetailAction>
          <DetailAction @action="getExportCSV()">
            <template #title>{{ $t("tools.import_export_set.export") }}</template>
            {{ $t("tools.import_export_set.export_sub") }}
            <template #button> {{ $t("tools.import_export_set.export_button") }} </template>
          </DetailAction>
        </div>
      </BaseCard>
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiAlert class="mr-2" />
            <span> {{ $t("tools.actions") }} </span>
            <template #description>
              <!-- eslint-disable-next-line vue/no-v-html -->
              <div v-html="DOMPurify.sanitize($t('tools.actions_sub'))"></div>
            </template>
          </BaseSectionHeader>
        </template>
        <div class="divide-y divide-gray-300 border-t border-gray-300 px-6 pb-3">
          <DetailAction @action="ensureAssetIDs">
            <template #title>{{ $t("tools.actions_set.ensure_ids") }}</template>
            {{ $t("tools.actions_set.ensure_ids_sub") }}
            <template #button> {{ $t("tools.actions_set.ensure_ids_button") }} </template>
          </DetailAction>
          <DetailAction @action="ensureImportRefs">
            <template #title>{{ $t("tools.actions_set.ensure_import_refs") }}</template>
            {{ $t("tools.actions_set.ensure_import_refs_sub") }}
            <template #button> {{ $t("tools.actions_set.ensure_import_refs_button") }} </template>
          </DetailAction>
          <DetailAction @action="resetItemDateTimes">
            <template #title> {{ $t("tools.actions_set.zero_datetimes") }} </template>
            <!-- eslint-disable-next-line vue/no-v-html -->
            <div v-html="DOMPurify.sanitize($t('tools.actions_set.zero_datetimes_sub'))"></div>
            <template #button> {{ $t("tools.actions_set.zero_datetimes_button") }} </template>
          </DetailAction>
          <DetailAction @action="setPrimaryPhotos">
            <template #title> {{ $t("tools.actions_set.set_primary_photo") }} </template>
            <!-- eslint-disable-next-line vue/no-v-html -->
            <div v-html="DOMPurify.sanitize($t('tools.actions_set.set_primary_photo_sub'))"></div>
            <template #button> {{ $t("tools.actions_set.set_primary_photo_button") }} </template>
          </DetailAction>
        </div>
      </BaseCard>
    </BaseContainer>
  </div>
</template>

<script setup lang="ts">
  import DOMPurify from "dompurify";
  import { toast } from "@/components/ui/sonner";
  import MdiFileChart from "~icons/mdi/file-chart";
  import MdiArrowRight from "~icons/mdi/arrow-right";
  import MdiDatabase from "~icons/mdi/database";
  import MdiAlert from "~icons/mdi/alert";
  import { useDialog } from "~/components/ui/dialog-provider";

  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "Homebox | Tools",
  });

  const { openDialog } = useDialog();

  const api = useUserApi();
  const confirm = useConfirm();

  function getBillOfMaterials() {
    const url = api.reports.billOfMaterialsURL();
    window.open(url, "_blank");
  }

  function getExportCSV() {
    const url = api.items.exportURL();
    window.open(url, "_blank");
  }

  async function ensureAssetIDs() {
    const { isCanceled } = await confirm.open(
      "Are you sure you want to ensure all assets have an ID? This can take a while and cannot be undone."
    );

    if (isCanceled) {
      return;
    }

    const result = await api.actions.ensureAssetIDs();

    if (result.error) {
      toast.error("Failed to ensure asset IDs.");
      return;
    }

    toast.success(`${result.data.completed} assets have been updated.`);
  }

  async function ensureImportRefs() {
    const { isCanceled } = await confirm.open(
      "Are you sure you want to ensure all assets have an import_ref? This can take a while and cannot be undone."
    );

    if (isCanceled) {
      return;
    }

    const result = await api.actions.ensureImportRefs();

    if (result.error) {
      toast.error("Failed to ensure import refs.");
      return;
    }

    toast.success(`${result.data.completed} assets have been updated.`);
  }

  async function resetItemDateTimes() {
    const { isCanceled } = await confirm.open(
      "Are you sure you want to reset all date and time values? This can take a while and cannot be undone."
    );

    if (isCanceled) {
      return;
    }

    const result = await api.actions.resetItemDateTimes();

    if (result.error) {
      toast.error("Failed to reset date and time values.");
      return;
    }

    toast.success(`${result.data.completed} assets have been updated.`);
  }

  async function setPrimaryPhotos() {
    const { isCanceled } = await confirm.open(
      "Are you sure you want to set primary photos? This can take a while and cannot be undone."
    );

    if (isCanceled) {
      return;
    }

    const result = await api.actions.setPrimaryPhotos();

    if (result.error) {
      toast.error("Failed to set primary photos.");
      return;
    }

    toast.success(`${result.data.completed} assets have been updated.`);
  }
</script>

<style scoped></style>
