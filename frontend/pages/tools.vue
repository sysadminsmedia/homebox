<template>
  <div>
    <AppImportDialog v-model="modals.import" />
    <BaseContainer class="flex flex-col gap-4 mb-6">
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiFileChart class="mr-2" />
            <span> Reports </span>
            <template #description> Generate different reports for your inventory. </template>
          </BaseSectionHeader>
        </template>
        <div class="border-t px-6 pb-3 border-gray-300 divide-gray-300 divide-y">
          <DetailAction @action="navigateTo('/reports/label-generator')">
            <template #title>Asset ID Labels</template>
            Generates a printable PDF of labels for a range of Asset ID. These are not specific to your inventory so you
            are able to print labels ahead of time and apply them to your inventory when you receive them.
            <template #button>
              Label Generator
              <MdiArrowRight class="ml-2" />
            </template>
          </DetailAction>
          <DetailAction @action="getBillOfMaterials()">
            <template #title>Bill of Materials</template>
            Generates a TSV (Tab Separated Values) file that can be imported into a spreadsheet program. This is a
            summary of your inventory with basic item and pricing information.
            <template #button> Generate BOM </template>
          </DetailAction>
        </div>
      </BaseCard>
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiDatabase class="mr-2" />
            <span> Import / Export </span>
            <template #description>
              Import and export your inventory to and from a CSV file. This is useful for migrating your inventory to a
              new instance of Homebox.
            </template>
          </BaseSectionHeader>
        </template>
        <div class="border-t px-6 pb-3 border-gray-300 divide-gray-300 divide-y">
          <DetailAction @action="modals.import = true">
            <template #title>Import Inventory</template>
            Imports the standard CSV format for Homebox. This will <b>not</b> overwrite any existing items in your
            inventory. It will only add new items.
          </DetailAction>
          <DetailAction @action="getExportTSV()">
            <template #title>Export Inventory</template>
            Exports the standard CSV format for Homebox. This will export all items in your inventory.
          </DetailAction>
        </div>
      </BaseCard>
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiAlert class="mr-2" />
            <span> Inventory Actions </span>
            <template #description>
              Apply Actions to your inventory in bulk. These are irreversible actions. <b>Be careful.</b>
            </template>
          </BaseSectionHeader>
        </template>
        <div class="border-t px-6 pb-3 border-gray-300 divide-gray-300 divide-y">
          <DetailAction @action="ensureAssetIDs">
            <template #title>Ensure Asset IDs</template>
            Ensures that all items in your inventory have a valid asset_id field. This is done by finding the highest
            current asset_id field in the database and applying the next value to each item that has an unset asset_id
            field. This is done in order of the created_at field.
          </DetailAction>
          <DetailAction @action="ensureImportRefs">
            <template #title>Ensures Import Refs</template>
            Ensures that all items in your inventory have a valid import_ref field. This is done by randomly generating
            a 8 character string for each item that has an unset import_ref field.
          </DetailAction>
          <DetailAction @action="resetItemDateTimes">
            <template #title> Zero Item Date Times</template>
            Resets the time value for all date time fields in your inventory to the beginning of the date. This is to
            fix a bug that was introduced early on in the development of the site that caused the time value to be
            stored with the time which caused issues with date fields displaying accurate values.
            <a class="link" href="https://github.com/hay-kot/homebox/issues/236" target="_blank">
              See Github Issue #236 for more details.
            </a>
          </DetailAction>
          <DetailAction @action="setPrimaryPhotos">
            <template #title> Set Primary Photos </template>
            In version v0.10.0 of Homebox, the primary image field was added to attachments of type photo. This action
            will set the primary image field to the first image in the attachments array in the database, if it is not
            already set. <a class="link" href="https://github.com/hay-kot/homebox/pull/576">See GitHub PR #576</a>
          </DetailAction>
        </div>
      </BaseCard>
    </BaseContainer>
  </div>
</template>

<script setup lang="ts">
  import MdiFileChart from "~icons/mdi/file-chart";
  import MdiArrowRight from "~icons/mdi/arrow-right";
  import MdiDatabase from "~icons/mdi/database";
  import MdiAlert from "~icons/mdi/alert";

  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "Homebox | Profile",
  });

  const modals = ref({
    item: false,
    location: false,
    label: false,
    import: false,
  });

  const api = useUserApi();
  const confirm = useConfirm();
  const notify = useNotifier();

  function getBillOfMaterials() {
    const url = api.reports.billOfMaterialsURL();
    window.open(url, "_blank");
  }

  function getExportTSV() {
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
      notify.error("Failed to ensure asset IDs.");
      return;
    }

    notify.success(`${result.data.completed} assets have been updated.`);
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
      notify.error("Failed to ensure import refs.");
      return;
    }

    notify.success(`${result.data.completed} assets have been updated.`);
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
      notify.error("Failed to reset date and time values.");
      return;
    }

    notify.success(`${result.data.completed} assets have been updated.`);
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
      notify.error("Failed to set primary photos.");
      return;
    }

    notify.success(`${result.data.completed} assets have been updated.`);
  }
</script>

<style scoped></style>
