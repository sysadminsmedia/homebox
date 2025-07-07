<template>
  <Dialog dialog-id="product-import">  
    <DialogContent  :class="'w-full md:max-w-xl lg:max-w-4xl'" >
      <DialogHeader>
        <DialogTitle>{{ $t("components.item.product_import.title") }}</DialogTitle>
      </DialogHeader>

      <div class="flex items-center space-x-4">
        <FormTextField :disabled=searching class="w-[30%]" :modelValue="barcode" :label="$t('components.item.product_import.barcode')" />
        <Button :variant="searching ? 'destructive' : 'default'" @click="retrieveProductInfo(barcode)" style="margin-top: auto"> 
          <div class="relative mx-2">
            <div class="absolute inset-0 flex items-center justify-center">
              <MdiBarcode class="size-5 group-hover:hidden" />
            </div>
          </div>
          {{ searching ? "Cancel" : "Search product" }}
        </Button>
      </div>
          
      <div class="divide-y border-t" />

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
                    'justify-end': h.align === 'left',
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
              class='cursor-pointer'
              :class="{ selected: selectedRow === index }" 
              @click="selectProduct(index)">
                <TableCell v-for="h in headers"
                  :class="{
                  'text-center': h.align === 'center',
                  'text-left': h.align === 'left',
                }">

                  <template v-if="h.type === 'name'">
                      <div class="flex items-center space-x-4">
                        <img :src="p.imageBase64" class="w-16 rounded object-fill shadow-sm" alt="Product's photo" />
                        <span class="text-sm font-medium">
                          {{ p.item.name }}
                        </span>
                      </div>
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
  import { Button, ButtonGroup } from "~/components/ui/button";
  import type { BarcodeProduct } from "~~/lib/api/types/data-contracts";
  import { useDialog } from "~/components/ui/dialog-provider";
  import MdiBarcode from "~icons/mdi/barcode";
  import type { TableData } from "~/components/Item/View/Table.types";
  const { openDialog, activeDialog, closeDialog } = useDialog();

  const searching = ref(false);
  const barcode = ref<string>("");
  const products = ref<BarcodeProduct[] | null>(null);
  const selectedRow = ref(-1);

  import type { ItemSummary } from "~~/lib/api/types/data-contracts";

  type BarcodeTableHeader = {
    text: string;
    value: string;
    align?: "left" | "center" | "right";
    type?: "name";
  };

  const defaultHeaders = [
    {
      text: "items.name",
      value: "name",
      align: "left",
      type: "name",
    },
    { text: "items.manufacturer", value: "manufacturer", align: "center"},
    { text: "items.model_number", value: "modelNumber", align: "center"},
    { text: "DB source", value: "search_engine_name", align: "center"},
  ] satisfies BarcodeTableHeader[];

  // Need for later filtering
  const headers = defaultHeaders;

  watch(
    () => activeDialog.value,
    active => {
      if (active && active.id === "product-import") {
        selectedRow.value = -1;

        if(active.params)
        {
          // Reset if the barcode is different
          if(active.params != barcode.value)
          {
            barcode.value = active.params;

            retrieveProductInfo(barcode.value).then(() =>
            {
              console.log("Processing finished");
            });
          }
        }
        else
        {
          barcode.value = "";
          products.value = null;
        }
      }
    }
  );

  const api = useUserApi();

  async function createItem(close = true) {
    if (products !== null)
    {
      var p = products.value![selectedRow.value];
      openDialog("create-item", p);  
    }
  }

  async function retrieveProductInfo(barcode: string) {
    products.value = null;
    searching.value = true;

    if (!barcode || barcode.trim().length === 0) {
      console.error('Invalid barcode provided');
      return;
    }

    try {
      const result = await api.products.searchFromBarcode(barcode.trim());
      if(result.error)
      {
        console.error('API Error:', result.error);
        return;
      }
      else
      {
        products.value = result.data;
      }
    } catch (error) {
      console.error('Failed to retrieve product info:', error);
    } finally {
      searching.value = false;
    }
  }

  function extractValue(data: TableData, value: string) {
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
    if(selectedRow.value == index)
    {
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