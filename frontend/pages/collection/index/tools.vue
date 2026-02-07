<template>
  <div>
    <AppImportDialog />
    <BaseContainer class="m-0 flex flex-col gap-4 px-0">
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiFileChart class="mr-2" />
            <span> {{ $t("tools.reports") }} </span>
            <template #description> {{ $t("tools.reports_sub") }} </template>
          </BaseSectionHeader>
        </template>
        <div class="divide-y border-t p-4">
          <DetailAction to="/reports/label-generator">
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
        <div class="divide-y border-t px-6 pb-3">
          <DetailAction @action="openDialog(DialogID.Import)">
            <template #title> {{ $t("tools.import_export_set.import") }} </template>
            <!-- eslint-disable-next-line vue/no-v-html -->
            <div v-html="DOMPurify.sanitize($t('tools.import_export_set.import_sub'))" />
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
              <div v-html="DOMPurify.sanitize($t('tools.actions_sub'))" />
            </template>
          </BaseSectionHeader>
        </template>
        <div class="divide-y border-t px-6 pb-3">
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
            <div v-html="DOMPurify.sanitize($t('tools.actions_set.zero_datetimes_sub'))" />
            <template #button> {{ $t("tools.actions_set.zero_datetimes_button") }} </template>
          </DetailAction>
          <DetailAction @action="setPrimaryPhotos">
            <template #title> {{ $t("tools.actions_set.set_primary_photo") }} </template>
            <!-- eslint-disable-next-line vue/no-v-html -->
            <div v-html="DOMPurify.sanitize($t('tools.actions_set.set_primary_photo_sub'))" />
            <template #button> {{ $t("tools.actions_set.set_primary_photo_button") }} </template>
          </DetailAction>
          <DetailAction @action="createMissingThumbnails">
            <template #title> {{ $t("tools.actions_set.create_missing_thumbnails") }} </template>
            <!-- eslint-disable-next-line vue/no-v-html -->
            <div v-html="DOMPurify.sanitize($t('tools.actions_set.create_missing_thumbnails_sub'))" />
            <template #button> {{ $t("tools.actions_set.create_missing_thumbnails_button") }} </template>
          </DetailAction>
          <DetailAction @action="wipeInventory">
            <template #title> {{ $t("tools.actions_set.wipe_inventory") }} </template>
            <!-- eslint-disable-next-line vue/no-v-html -->
            <div v-html="DOMPurify.sanitize($t('tools.actions_set.wipe_inventory_sub'))" />
            <template #button> {{ $t("tools.actions_set.wipe_inventory_button") }} </template>
          </DetailAction>
        </div>
      </BaseCard>
    </BaseContainer>
  </div>
</template>

<script setup lang="ts">
  import DOMPurify from "dompurify";
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiFileChart from "~icons/mdi/file-chart";
  import MdiArrowRight from "~icons/mdi/arrow-right";
  import MdiDatabase from "~icons/mdi/database";
  import MdiAlert from "~icons/mdi/alert";
  import { useDialog } from "~/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import AppImportDialog from "@/components/App/ImportDialog.vue";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseCard from "@/components/Base/Card.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import DetailAction from "@/components/DetailAction.vue";

  const { t } = useI18n();

  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "HomeBox | " + t("collection.tabs.tools"),
  });

  const { openDialog } = useDialog();

  const api = useUserApi();
  const confirm = useConfirm();
  const pubApi = usePublicApi();
  const { data: status } = useAsyncData(async () => {
    const { data } = await pubApi.status();
    return data;
  });

  const getBillOfMaterials = () => {
    const url = api.reports.billOfMaterialsURL();
    window.open(url, "_blank");
  };

  const getExportCSV = () => {
    const url = api.items.exportURL();
    window.open(url, "_blank");
  };

  const ensureAssetIDs = async () => {
    const { isCanceled } = await confirm.open(t("tools.actions_set.ensure_ids_confirm"));

    if (isCanceled) {
      return;
    }

    const result = await api.actions.ensureAssetIDs();

    if (result.error) {
      toast.error(t("tools.toast.failed_ensure_ids"));
      return;
    }

    toast.success(t("tools.toast.asset_success", { results: result.data.completed }));
  };

  const createMissingThumbnails = async () => {
    const { isCanceled } = await confirm.open(t("tools.actions_set.create_missing_thumbnails_confirm"));

    if (isCanceled) {
      return;
    }

    const result = await api.actions.createMissingThumbnails();

    if (result.error) {
      toast.error(t("tools.toast.failed_create_missing_thumbnails"));
      return;
    }

    toast.success(t("tools.toast.asset_success", { results: result.data.completed }));
  };

  const ensureImportRefs = async () => {
    const { isCanceled } = await confirm.open(t("tools.import_export_set.import_ref_confirm"));

    if (isCanceled) {
      return;
    }

    const result = await api.actions.ensureImportRefs();

    if (result.error) {
      toast.error(t("tools.toast.failed_ensure_import_refs"));
      return;
    }

    toast.success(t("tools.toast.asset_success", { results: result.data.completed }));
  };

  const resetItemDateTimes = async () => {
    const { isCanceled } = await confirm.open(t("tools.actions_set.zero_datetimes_confirm"));

    if (isCanceled) {
      return;
    }

    const result = await api.actions.resetItemDateTimes();

    if (result.error) {
      toast.error(t("tools.toast.failed_zero_datetimes"));
      return;
    }

    toast.success(t("tools.toast.asset_success", { results: result.data.completed }));
  };

  const setPrimaryPhotos = async () => {
    const { isCanceled } = await confirm.open(t("tools.actions_set.set_primary_photo_confirm"));

    if (isCanceled) {
      return;
    }

    const result = await api.actions.setPrimaryPhotos();

    if (result.error) {
      toast.error(t("tools.toast.failed_set_primary_photos"));
      return;
    }

    toast.success(t("tools.toast.asset_success", { results: result.data.completed }));
  };

  const wipeInventory = async () => {
    if (status.value?.demo) {
      await confirm.open(t("tools.demo_mode_error.wipe_inventory"));
      return;
    }

    openDialog(DialogID.WipeInventory, {
      onClose: async result => {
        if (!result) {
          return;
        }

        const apiResult = await api.actions.wipeInventory({
          wipeTags: result.wipeTags,
          wipeLocations: result.wipeLocations,
          wipeMaintenance: result.wipeMaintenance,
        });

        if (apiResult.error) {
          toast.error(t("tools.toast.failed_wipe_inventory"));
          return;
        }

        toast.success(t("tools.toast.wipe_inventory_success", { results: apiResult.data.completed }));
      },
    });
  };
</script>
