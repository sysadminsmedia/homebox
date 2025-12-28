<template>
  <Dialog :dialog-id="DialogID.ProductImport">
    <DialogContent :class="'w-full md:max-w-xl lg:max-w-4xl'">
      <DialogHeader>
        <DialogTitle>{{ $t("components.item.product_import.title") }}</DialogTitle>
      </DialogHeader>

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
          :tag="$t('components.item.product_import.barcode')"
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

      <Separator />

      <BaseCard>
        <Table class="w-full">
          <TableHeader>
            <TableRow>
              <TableHead
                v-for="h in headers"
                :key="h.value"
                class="text-no-transform bg-secondary text-sm text-secondary-foreground hover:bg-secondary/90"
              >
                <div
                  class="flex items-center gap-1"
                  :class="{
                    'justify-center': h.align === 'center',
                  }"
                >
                  <template v-if="typeof h === 'string'">{{ h }}</template>
                  <template v-else>{{ $t(h.text) }}</template>
                </div>
              </TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            <TableRow
              v-for="(p, index) in products"
              :key="index"
              class="cursor-pointer"
              :class="{ selected: selectedRow === index }"
              @click="selectProduct(index)"
            >
              <TableCell
                v-for="h in headers"
                :key="h.value"
                :class="{
                  'text-center': h.align === 'center',
                }"
              >
                <template v-if="h.type === 'name'">
                  <div class="flex items-center space-x-4">
                    <img :src="p.imageBase64" class="w-16 rounded object-fill shadow-sm" alt="Product's photo" />
                    <span class="text-sm font-medium">
                      {{ p.item.name }}
                    </span>
                  </div>
                </template>
                <template v-else-if="h.type === 'url'">
                  <NuxtLink class="underline" :to="'https://' + extractValue(p, h.value)" target="_blank">{{
                    extractValue(p, h.value)
                  }}</NuxtLink>
                </template>

                <slot v-else :name="cell(h)">
                  {{ extractValue(p, h.value) }}
                </slot>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </BaseCard>

      <DialogFooter>
        <Button type="import" :disabled="selectedRow === -1" @click="createItem"> Import selected </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { Button } from "~/components/ui/button";
  import type { BarcodeProduct } from "~~/lib/api/types/data-contracts";
  import { useDialog } from "~/components/ui/dialog-provider";
  import MdiAlertCircleOutline from "~icons/mdi/alert-circle-outline";
  import MdiBarcode from "~icons/mdi/barcode";
  import MdiLoading from "~icons/mdi/loading";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
  import { Separator } from "@/components/ui/separator";
  import BaseCard from "@/components/Base/Card.vue";
  import FormTextField from "@/components/Form/TextField.vue";

  const { openDialog, registerOpenDialogCallback } = useDialog();
  const { t } = useI18n();

  const searching = ref(false);
  const barcode = ref<string>("");
  const products = ref<BarcodeProduct[] | null>(null);
  const selectedRow = ref(-1);
  const errorMessage = ref<string | null>(null);

  type BarcodeTableHeader = {
    text: string;
    value: string;
    align?: "left" | "center" | "right";
    type?: "name" | "url";
  };

  const defaultHeaders = [
    {
      text: "items.name",
      value: "name",
      align: "center",
      type: "name",
    },
    { text: "items.manufacturer", value: "manufacturer", align: "center" },
    { text: "items.model_number", value: "modelNumber", align: "center" },
    { text: "components.item.product_import.db_source", value: "search_engine_name", align: "center", type: "url" },
  ] satisfies BarcodeTableHeader[];

  // Need for later filtering
  const headers = defaultHeaders;

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.ProductImport, params => {
      selectedRow.value = -1;
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
        products.value = null;
      }
    });

    onUnmounted(cleanup);
  });

  const api = useUserApi();

  function createItem() {
    if (
      products.value !== null &&
      products.value.length > 0 &&
      selectedRow.value >= 0 &&
      selectedRow.value < products.value.length
    ) {
      const p = products.value![selectedRow.value];
      openDialog(DialogID.CreateItem, {
        params: { product: p },
      });
    }
  }

  async function retrieveProductInfo(barcode: string) {
    errorMessage.value = null;

    if (!barcode || barcode.trim().length === 0 || !/^[0-9]+$/.test(barcode)) {
      errorMessage.value = t("components.item.product_import.error_invalid_barcode");
      console.error(errorMessage.value);
      return;
    }

    products.value = null;
    searching.value = true;

    try {
      const result = await api.products.searchFromBarcode(barcode.trim());
      if (result.error) {
        errorMessage.value = t("errors.api_failure") + result.error;
        console.error(errorMessage.value);
      } else {
        if (result.data === undefined || result.data.length === undefined || result.data.length === 0) {
          errorMessage.value = t("components.item.product_import.error_not_found");
        }

        products.value = result.data;
      }
    } catch (error) {
      errorMessage.value = t("components.item.product_import.error_exception") + error;
      console.error(errorMessage.value);
    } finally {
      searching.value = false;
    }
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function extractValue(data: Record<string, any>, value: string) {
    const parts = value.split(".");
    let current = data;
    for (const part of parts) {
      current = current[part];
    }
    return current;
  }

  function cell(h: BarcodeTableHeader) {
    return `cell-${h.value.replace(".", "_")}`;
  }

  function selectProduct(index: number) {
    // Unselect if already selected
    if (selectedRow.value === index) {
      selectedRow.value = -1;
      return;
    }

    selectedRow.value = index;
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
