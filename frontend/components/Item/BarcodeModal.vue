<template>
  <Dialog dialog-id="product-import">  
    <DialogContent  :class="'w-full md:max-w-xl lg:max-w-4xl'" >
      <DialogHeader>
        <DialogTitle>{{ $t("components.item.product_import.title") }}</DialogTitle>
      </DialogHeader>

      <div class="flex items-center space-x-4">
        <FormTextField :disabled=searching class="w-[30%]" :modelValue="barcode" disabled :label="$t('components.item.product_import.barcode')" />
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
                @click="sortBy(h.value)"
              >
                <div
                  class="flex items-center gap-1"
                  :class="{
                    'justify-center': h.align === 'center',
                    'justify-start': h.align === 'right',
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
              @click="selectProduct(index, p)">
                <TableCell v-for="h in headers"
                  :class="{
                  'text-center': h.align === 'center',
                  'text-right': h.align === 'right',
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

      <form class="flex flex-col gap-4" @submit.prevent="submitCsvFile">
        <DialogFooter>
          <Button type="import" :disabled="selectedRow === -1" @click=createItem> Import selected </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { Button, ButtonGroup } from "~/components/ui/button";
  import type { BarcodeProduct } from "~~/lib/api/types/data-contracts";
  import { useDialog } from "~/components/ui/dialog-provider";
  import MdiBarcode from "~icons/mdi/barcode";

  const { openDialog, activeDialog, closeDialog } = useDialog();

  const searching = ref(false);
  const barcode = ref<string | null>(null);
  const products = ref<BarcodeProduct[] | null>(null);
  const selectedRow = ref(-1);

  const defaultHeaders = [
    {
      text: "items.name",
      value: "name",
      enabled: true,
      type: "name",
    },
    { text: "items.manufacturer", value: "manufacturer", align: "center", enabled: true },
    { text: "items.model_number", value: "modelNumber", align: "center", enabled: true },
    { text: "DB source", value: "search_engine_name", align: "center", enabled: true },
  ] satisfies TableHeaderType[];

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
          barcode.value = null;
          products.value = null;
        }
      }
    }
  );

  const api = useUserApi();

  async function createItem(close = true) {
    var p = products.value[selectedRow.value];
    closeDialog("product-import");
    openDialog("create-item", p);
  }

  async function retrieveProductInfo(barcode: string) {
    products.value = null;
    searching.value = true;
    const result = await api.actions.getEAN(barcode);
    searching.value = false;

    if(result.error)
      return

    products.value = result.data;
  }

  function extractValue(data: TableData, value: string) {
    const parts = value.split(".");
    let current = data;
    for (const part of parts) {
      current = current[part];
    }
    return current;
  }

  function cell(h: TableHeaderType) {
    return `cell-${h.value.replace(".", "_")}`;
  }

  function selectProduct(index) {
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