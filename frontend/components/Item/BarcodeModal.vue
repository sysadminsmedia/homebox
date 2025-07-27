<template>
  <BaseModal
    :dialog-id="DialogID.ProductImport"
    :title="$t('components.item.product_import.title')"
    >
    <div
      v-if="errorMessage"
      class="flex items-center gap-2 rounded-md border border-destructive bg-destructive/10 p-4 text-destructive"
      role="alert"
    >
      <MdiAlertCircleOutline class="text-destructive" />
      <span class="text-sm font-medium">{{ errorMessage }}</span>
    </div>

    <div class="flex items-center gap-3">
      <FormTextField
        v-model="barcode"
        :disabled="searching"
        class="w-[30%]"
        :label="$t('components.item.product_import.barcode')"
        @keyup.enter="retrieveProductInfo(barcode)"
      />
      <Button
        :variant="searching ? 'destructive' : 'default'"
        class="mt-auto h-10"
        @click="retrieveProductInfo(barcode)"
      >
        <MdiLoading v-if="searching" class="animate-spin" />
        <div v-if="!searching" class="relative mx-2">
          <div class="absolute inset-0 flex items-center justify-center">
            <MdiBarcode class="size-5 group-hover:hidden" />
          </div>
        </div>
        {{ searching ? $t("global.cancel") : $t("components.item.product_import.search_item") }}
      </Button>
    </div>

    <Separator class="mt-4 sm:mt-0" />

    <ItemViewSelectable
      :items="products"
      :item-type="'barcodeproduct'"
      :selection-mode="true"
      @update:selected-item="onSelectedItemChange"
    />

    <DialogFooter>
      <Button type="import" :disabled="selectedItem === null" @click="createItem"> Import selected </Button>
    </DialogFooter>
  </BaseModal>
</template>

<script setup lang="ts">
  import { ref } from "vue";
  import { useI18n } from "vue-i18n";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import BaseModal from "@/components/App/CreateModal.vue";
  import { Button } from "~/components/ui/button";
  import type { BarcodeProduct, ItemSummary } from "~~/lib/api/types/data-contracts";
  import { useDialog } from "~/components/ui/dialog-provider";
  import MdiAlertCircleOutline from "~icons/mdi/alert-circle-outline";
  import MdiBarcode from "~icons/mdi/barcode";
  import MdiLoading from "~icons/mdi/loading";

  const { openDialog, registerOpenDialogCallback } = useDialog();
  const { t } = useI18n();

  const searching = ref(false);
  const barcode = ref<string>("");
  const products = ref<BarcodeProduct[]>([]);
  const errorMessage = ref<string | null>(null);
  const selectedItem = ref<BarcodeProduct | null>(null);

  function onSelectedItemChange(item: BarcodeProduct | ItemSummary | null) {
    if (item === null) selectedItem.value = null;
    else selectedItem.value = item as BarcodeProduct;
  }

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.ProductImport, params => {
      searching.value = false;
      errorMessage.value = null;

      if (params?.barcode) {
        // Reset if the barcode is different
        if (params.barcode !== barcode.value) {
          barcode.value = params.barcode;

          retrieveProductInfo(barcode.value).then(() => {
            console.log("Processing finished");
          });
        }
      } else {
        barcode.value = "";
        products.value = [];
      }
    });

    onUnmounted(cleanup);
  });

  const api = useUserApi();

  function createItem() {
    if (selectedItem.value === null) return;

    openDialog(DialogID.CreateItem, {
        params: { product: p },
    });
  }

  async function retrieveProductInfo(barcode: string) {
    errorMessage.value = null;

    if (!barcode || barcode.trim().length === 0 || !/^[0-9]+$/.test(barcode)) {
      errorMessage.value = t("components.item.product_import.error_invalid_barcode");
      console.error(errorMessage.value);
      return;
    }

    products.value = [];
    searching.value = true;

    try {
      const result = await api.products.searchFromBarcode(barcode.trim());
      if (result.error) {
        errorMessage.value = t("errors.api_failure") + result.error;
        console.error(errorMessage.value);
      } else if (result.data === undefined || result.data.length === undefined || result.data.length === 0) {
        errorMessage.value = t("components.item.product_import.error_not_found");
        products.value = [];
      } else {
        products.value = result.data;
      }
    } catch (error) {
      errorMessage.value = t("components.item.product_import.error_exception") + error;
      console.error(errorMessage.value);
    } finally {
      searching.value = false;
    }
  }
</script>

<style>
  tr.selected {
    background-color: hsl(var(--primary));
    color: hsl(var(--background));
  }

  tr:hover.selected {
    background-color: hsl(var(--primary));
  }
</style>
