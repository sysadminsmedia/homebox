<script setup lang="ts">
  import type { AnyDetail, Detail, Details } from "~~/components/global/DetailsSection/types";
  import { filterZeroValues } from "~~/components/global/DetailsSection/types";
  import type { ItemAttachment } from "~~/lib/api/types/data-contracts";
  import MdiClose from "~icons/mdi/close";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPlus from "~icons/mdi/plus";
  import MdiMinus from "~icons/mdi/minus";
  import MdiDownload from "~icons/mdi/download";

  definePageMeta({
    middleware: ["auth"],
  });

  const route = useRoute();
  const api = useUserApi();
  const toast = useNotifier();

  const itemId = computed<string>(() => route.params.id as string);
  const preferences = useViewPreferences();

  const hasNested = computed<boolean>(() => {
    return route.fullPath.split("/").at(-1) !== itemId.value;
  });

  const { data: item, refresh } = useAsyncData(itemId.value, async () => {
    const { data, error } = await api.items.get(itemId.value);
    if (error) {
      toast.error("Failed to load item");
      navigateTo("/home");
      return;
    }
    return data;
  });
  onMounted(() => {
    refresh();
  });

  const lastRoute = ref(route.fullPath);
  watchEffect(() => {
    if (lastRoute.value.endsWith("edit")) {
      refresh();
    }

    lastRoute.value = route.fullPath;
  });

  async function adjustQuantity(amount: number) {
    if (!item.value) {
      return;
    }

    const newQuantity = item.value.quantity + amount;
    if (newQuantity < 0) {
      toast.error("Quantity cannot be negative");
      return;
    }

    const resp = await api.items.patch(item.value.id, {
      id: item.value.id,
      quantity: newQuantity,
    });

    if (resp.error) {
      toast.error("Failed to adjust quantity");
      return;
    }

    item.value.quantity = newQuantity;
  }

  type FilteredAttachments = {
    attachments: ItemAttachment[];
    warranty: ItemAttachment[];
    manuals: ItemAttachment[];
    receipts: ItemAttachment[];
  };

  type Photo = {
    src: string;
  };

  const photos = computed<Photo[]>(() => {
    return (
      item.value?.attachments.reduce((acc, cur) => {
        if (cur.type === "photo") {
          acc.push({
            // @ts-expect-error - it's impossible for this to be null at this point
            src: api.authURL(`/items/${item.value.id}/attachments/${cur.id}`),
          });
        }
        return acc;
      }, [] as Photo[]) || []
    );
  });

  const attachments = computed<FilteredAttachments>(() => {
    if (!item.value) {
      return {
        attachments: [],
        manuals: [],
        warranty: [],
        receipts: [],
      };
    }

    return item.value.attachments.reduce(
      (acc, attachment) => {
        if (attachment.type === "photo") {
          return acc;
        }
        if (attachment.type === "warranty") {
          acc.warranty.push(attachment);
        } else if (attachment.type === "manual") {
          acc.manuals.push(attachment);
        } else if (attachment.type === "receipt") {
          acc.receipts.push(attachment);
        } else {
          acc.attachments.push(attachment);
        }
        return acc;
      },
      {
        attachments: [] as ItemAttachment[],
        warranty: [] as ItemAttachment[],
        manuals: [] as ItemAttachment[],
        receipts: [] as ItemAttachment[],
      }
    );
  });

  const assetID = computed<Details>(() => {
    if (!item.value) {
      return [];
    }

    if (item.value?.assetId === "000-000") {
      return [];
    }

    return [
      {
        name: "items.asset_id",
        text: item.value?.assetId,
      },
    ];
  });

  const itemDetails = computed<Details>(() => {
    if (!item.value) {
      return [];
    }

    const ret: Details = [
      {
        name: "items.quantity",
        text: item.value?.quantity,
        slot: "quantity",
      },
      {
        name: "items.serial_number",
        text: item.value?.serialNumber,
        copyable: true,
      },
      {
        name: "items.model_number",
        text: item.value?.modelNumber,
        copyable: true,
      },
      {
        name: "items.manufacturer",
        text: item.value?.manufacturer,
        copyable: true,
      },
      {
        name: "items.insured",
        text: item.value?.insured ? "Yes" : "No",
      },
      {
        name: "items.archived",
        text: item.value?.archived ? "Yes" : "No",
      },
      {
        name: "items.notes",
        type: "markdown",
        text: item.value?.notes,
      },
      ...assetID.value,
      ...item.value.fields.map(field => {
        /**
         * Support Special URL Syntax
         */
        const url = maybeUrl(field.textValue);
        if (url.isUrl) {
          return {
            type: "link",
            name: field.name,
            text: url.text,
            href: url.url,
          } as AnyDetail;
        }

        return {
          name: field.name,
          text: field.textValue,
        };
      }),
    ];

    if (!preferences.value.showEmpty) {
      return filterZeroValues(ret);
    }

    return ret;
  });

  const showAttachments = computed(() => {
    if (preferences.value?.showEmpty) {
      return true;
    }

    return (
      attachments.value.attachments.length > 0 ||
      attachments.value.warranty.length > 0 ||
      attachments.value.manuals.length > 0 ||
      attachments.value.receipts.length > 0
    );
  });

  const attachmentDetails = computed(() => {
    const details: Detail[] = [];

    const push = (name: string, slot: string) => {
      details.push({
        name,
        text: "",
        slot,
      });
    };

    if (attachments.value.attachments.length > 0) {
      push("items.attachments", "attachments");
    }

    if (attachments.value.warranty.length > 0) {
      push("items.warranty", "warranty");
    }

    if (attachments.value.manuals.length > 0) {
      push("items.manuals", "manuals");
    }

    if (attachments.value.receipts.length > 0) {
      push("items.receipts", "receipts");
    }

    return details;
  });

  const showWarranty = computed(() => {
    if (preferences.value.showEmpty) {
      return true;
    }
    return validDate(item.value?.warrantyExpires);
  });

  const warrantyDetails = computed(() => {
    const details: Details = [
      {
        name: "items.lifetime_warranty",
        text: item.value?.lifetimeWarranty ? "Yes" : "No",
      },
    ];

    if (item.value?.lifetimeWarranty) {
      details.push({
        name: "items.warranty_expires",
        text: "N/A",
      });
    } else {
      details.push({
        name: "items.warranty_expires",
        text: item.value?.warrantyExpires || "",
        type: "date",
        date: true,
      });
    }

    details.push({
      name: "items.warranty_details",
      type: "markdown",
      text: item.value?.warrantyDetails || "",
    });

    if (!preferences.value.showEmpty) {
      return filterZeroValues(details);
    }

    return details;
  });

  const showPurchase = computed(() => {
    if (preferences.value.showEmpty) {
      return true;
    }
    return item.value?.purchaseFrom || item.value?.purchasePrice !== 0;
  });

  const purchaseDetails = computed<Details>(() => {
    const v: Details = [
      {
        name: "items.purchased_from",
        text: item.value?.purchaseFrom || "",
      },
      {
        name: "items.purchase_price",
        text: String(item.value?.purchasePrice) || "",
        type: "currency",
      },
      {
        name: "items.purchase_date",
        text: item.value?.purchaseTime || "",
        type: "date",
        date: true,
      },
    ];

    if (!preferences.value.showEmpty) {
      return filterZeroValues(v);
    }

    return v;
  });

  const showSold = computed(() => {
    if (preferences.value.showEmpty) {
      return true;
    }
    return item.value?.soldTo || item.value?.soldPrice !== 0;
  });

  const soldDetails = computed<Details>(() => {
    const v: Details = [
      {
        name: "items.sold_to",
        text: item.value?.soldTo || "",
      },
      {
        name: "items.sold_price",
        text: String(item.value?.soldPrice) || "",
        type: "currency",
      },
      {
        name: "items.sold_at",
        text: item.value?.soldTime || "",
        type: "date",
        date: true,
      },
    ];

    if (!preferences.value.showEmpty) {
      return filterZeroValues(v);
    }

    return v;
  });

  const refDialog = ref<HTMLDialogElement>();
  const dialoged = reactive({
    src: "",
  });

  function openDialog(img: Photo) {
    // @ts-ignore - I don't know why this is happening
    refDialog.value?.showModal();
    dialoged.src = img.src;
  }

  function closeDialog() {
    // @ts-ignore - I don't know why this is happening
    refDialog.value?.close();
  }

  const refDialogBody = ref<HTMLDivElement>();
  onClickOutside(refDialogBody, () => {
    closeDialog();
  });

  const currentUrl = computed(() => {
    return window.location.href;
  });

  const currentPath = computed(() => {
    return route.path;
  });

  const tabs = computed(() => {
    return [
      {
        id: "details",
        name: "global.details",
        to: `/item/${itemId.value}`,
      },
      {
        id: "log",
        name: "global.maintenance",
        to: `/item/${itemId.value}/maintenance`,
      },
      {
        id: "edit",
        name: "global.edit",
        to: `/item/${itemId.value}/edit`,
      },
    ];
  });

  const fullpath = computedAsync(async () => {
    if (!item.value) {
      return [];
    }

    const resp = await api.items.fullpath(item.value.id);
    if (resp.error) {
      toast.error("Failed to load item");
      return [];
    }

    return resp.data;
  });

  const items = computedAsync(async () => {
    if (!item.value) {
      return [];
    }

    const resp = await api.items.getAll({
      parentIds: [item.value.id],
    });

    if (resp.error) {
      toast.error("Failed to load items");
      return [];
    }

    return resp.data.items;
  });
</script>

<template>
  <BaseContainer v-if="item" class="pb-8">
    <Title>{{ item.name }}</Title>
    <dialog ref="refDialog" class="fixed z-[999] overflow-visible bg-transparent">
      <div ref="refDialogBody" class="relative">
        <div class="absolute right-0 -mr-3 -mt-3 space-x-1 sm:-mr-4 sm:-mt-4">
          <a class="btn btn-circle btn-primary btn-sm sm:btn-md" :href="dialoged.src" download>
            <MdiDownload class="size-5" />
          </a>
          <button class="btn btn-circle btn-primary btn-sm sm:btn-md" @click="closeDialog()">
            <MdiClose class="size-5" />
          </button>
        </div>

        <img class="max-h-[80vh] max-w-[80vw]" :src="dialoged.src" />
      </div>
    </dialog>

    <section>
      <div class="rounded bg-base-100 p-3">
        <header class="mb-2">
          <div class="flex flex-wrap items-end gap-2">
            <div class="avatar placeholder mb-auto">
              <div class="w-12 rounded-full bg-neutral-focus text-neutral-content">
                <MdiPackageVariant class="size-7" />
              </div>
            </div>
            <div>
              <div v-if="fullpath && fullpath.length > 0" class="breadcrumbs py-0 text-sm">
                <ul class="text-base-content/70">
                  <li v-for="part in fullpath" :key="part.id" class="text-wrap">
                    <NuxtLink :to="`/${part.type}/${part.id}`"> {{ part.name }}</NuxtLink>
                  </li>
                </ul>
              </div>
              <h1 class="text-wrap pb-1 text-2xl">
                {{ item ? item.name : "" }}
              </h1>
              <div class="flex flex-wrap gap-2 pb-1">
                <LabelChip v-for="label in item?.labels || []" :key="label.id" :label="label" size="sm" />
              </div>
              <div class="flex flex-wrap gap-1 text-wrap text-xs">
                <div>
                  Created
                  <DateTime :date="item?.createdAt" />
                </div>
                -
                <div>
                  Updated
                  <DateTime :date="item?.updatedAt" />
                </div>
              </div>
            </div>
          </div>
        </header>
        <div class="divider my-0 mb-1"></div>
        <div class="prose max-w-full p-1">
          <Markdown v-if="item && item.description" class="text-base" :source="item.description"> </Markdown>
        </div>
      </div>

      <div class="mb-6 mt-3 flex flex-wrap items-center justify-between">
        <div class="btn-group">
          <NuxtLink
            v-for="t in tabs"
            :key="t.id"
            :to="t.to"
            class="btn btn-sm"
            :class="`${t.to === currentPath ? 'btn-active' : ''}`"
          >
            {{ $t(t.name) }}
          </NuxtLink>
        </div>

        <LabelMaker v-if="typeof item.assetId === 'string' && item.assetId != ''" :id="item.assetId" type="asset" />
        <LabelMaker v-else :id="item.id" type="item" />
      </div>
    </section>

    <section>
      <div class="space-y-6">
        <BaseCard v-if="!hasNested" collapsable>
          <template #title> {{ $t("items.details") }} </template>
          <template #title-actions>
            <div class="mt-2 flex flex-wrap items-center justify-between gap-4">
              <label class="label cursor-pointer">
                <input v-model="preferences.showEmpty" type="checkbox" class="toggle toggle-primary" />
                <span class="label-text ml-4"> Show Empty </span>
              </label>
              <div class="space-x-1">
                <CopyText :text="currentUrl" :icon-size="16" class="btn btn-circle btn-ghost btn-xs" />
                <PageQRCode />
              </div>
            </div>
          </template>
          <DetailsSection :details="itemDetails">
            <template #quantity="{ detail }">
              {{ detail.text }}
              <span
                class="my-0 ml-4 inline-flex gap-2 opacity-0 transition-opacity duration-75 group-hover:opacity-100"
              >
                <button class="btn btn-circle btn-xs" @click="adjustQuantity(-1)">
                  <MdiMinus class="size-3" />
                </button>
                <button class="btn btn-circle btn-xs" @click="adjustQuantity(1)">
                  <MdiPlus class="size-3" />
                </button>
              </span>
            </template>
          </DetailsSection>
        </BaseCard>

        <NuxtPage :item="item" :page-key="itemId" />
        <template v-if="!hasNested">
          <BaseCard v-if="photos && photos.length > 0">
            <template #title> {{ $t("items.photos") }} </template>
            <div
              class="scroll-bg container mx-auto flex max-h-[500px] flex-wrap gap-2 overflow-y-scroll border-t border-gray-300 p-4"
            >
              <button v-for="(img, i) in photos" :key="i" @click="openDialog(img)">
                <img class="max-h-[200px] rounded" :src="img.src" />
              </button>
            </div>
          </BaseCard>

          <BaseCard v-if="showAttachments" collapsable>
            <template #title> {{ $t("items.attachments") }} </template>
            <DetailsSection v-if="attachmentDetails.length > 0" :details="attachmentDetails">
              <template #manuals>
                <ItemAttachmentsList
                  v-if="attachments.manuals.length > 0"
                  :attachments="attachments.manuals"
                  :item-id="item.id"
                />
              </template>
              <template #attachments>
                <ItemAttachmentsList
                  v-if="attachments.attachments.length > 0"
                  :attachments="attachments.attachments"
                  :item-id="item.id"
                />
              </template>
              <template #warranty>
                <ItemAttachmentsList
                  v-if="attachments.warranty.length > 0"
                  :attachments="attachments.warranty"
                  :item-id="item.id"
                />
              </template>
              <template #receipts>
                <ItemAttachmentsList
                  v-if="attachments.receipts.length > 0"
                  :attachments="attachments.receipts"
                  :item-id="item.id"
                />
              </template>
            </DetailsSection>
            <div v-else>
              <p class="px-6 pb-4 text-base-content/70">No attachments found</p>
            </div>
          </BaseCard>

          <BaseCard v-if="showPurchase" collapsable>
            <template #title> {{ $t("items.purchase_details") }} </template>
            <DetailsSection :details="purchaseDetails" />
          </BaseCard>

          <BaseCard v-if="showWarranty" collapsable>
            <template #title> {{ $t("items.warranty_details") }} </template>
            <DetailsSection :details="warrantyDetails" />
          </BaseCard>

          <BaseCard v-if="showSold" collapsable>
            <template #title> {{ $t("items.sold_details") }} </template>
            <DetailsSection :details="soldDetails" />
          </BaseCard>
        </template>
      </div>
    </section>

    <section v-if="items && items.length > 0" class="my-6">
      <ItemViewSelectable :items="items" />
    </section>
  </BaseContainer>
</template>

<style lang="css" scoped>
  /* Style dialog background */
  dialog::backdrop {
    background: rgba(0, 0, 0, 0.5);
  }
</style>
