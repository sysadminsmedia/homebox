<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import type { AnyDetail, Details } from "~~/components/global/DetailsSection/types";
  import { filterZeroValues } from "~~/components/global/DetailsSection/types";
  import type { ItemAttachment } from "~~/lib/api/types/data-contracts";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPlus from "~icons/mdi/plus";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { Card } from "@/components/ui/card";
  import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbSeparator,
  } from "@/components/ui/breadcrumb";
  import { Button } from "@/components/ui/button";
  import { Badge } from "@/components/ui/badge";
  import { Separator } from "@/components/ui/separator";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Switch } from "@/components/ui/switch";
  import type { LocationItemDepth, LocationDescendantDepth, LocationItem } from "~~/lib/api/classes/items";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import BaseCard from "@/components/Base/Card.vue";
  import Currency from "~/components/global/Currency.vue";
  import DateTime from "~/components/global/DateTime.vue";
  import LabelMaker from "~/components/global/LabelMaker.vue";
  import Markdown from "~/components/global/Markdown.vue";
  import DetailsSection from "~/components/global/DetailsSection/DetailsSection.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import ItemViewSelectable from "~/components/Item/View/Selectable.vue";
  import { withScopedLocationPath } from "~/components/Item/location-badge";
  import { clampPageToTotal, isCurrentLocationPage } from "~/components/Item/View/pagination";
  import ItemAttachmentsList from "~/components/Item/AttachmentsList.vue";
  import ItemImageDialog from "~/components/Item/ImageDialog.vue";
  import LocationCard from "~/components/Location/Card.vue";
  import TagChip from "~/components/Tag/Chip.vue";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

  const { openDialog } = useDialog();

  const route = useRoute();
  const api = useUserApi();
  const preferences = useViewPreferences();

  const locationId = computed<string>(() => route.params.id as string);

  const { data: location } = useAsyncData(locationId.value, async () => {
    const { data, error } = await api.items.getLocation(locationId.value);
    if (error) {
      toast.error(t("locations.toast.failed_load_location"));
      navigateTo("/home");
      return;
    }

    return data;
  });

  const confirm = useConfirm();

  async function confirmDelete() {
    const { isCanceled } = await confirm.open(t("locations.location_items_delete_confirm"));
    if (isCanceled) {
      return;
    }

    const { error } = await api.items.deleteLocation(locationId.value);
    if (error) {
      toast.error(t("locations.toast.failed_delete_location"));
      return;
    }

    toast.success(t("locations.toast.location_deleted"));
    navigateTo("/locations");
  }

  function openCreateItem() {
    openDialog(DialogID.CreateEntity, {
      params: {
        baseType: "item",
      },
    });
  }

  function goToEdit() {
    navigateTo(`/location/${locationId.value}/edit`);
  }

  // Photos
  type Photo = {
    thumbnailSrc?: string;
    originalSrc: string;
    attachmentId: string;
    originalType?: string;
  };

  const photos = computed<Photo[]>(() => {
    if (!location.value?.attachments) {
      return [];
    }
    return location.value.attachments.reduce((acc, cur) => {
      if (cur.type === "photo") {
        const photo: Photo = {
          originalSrc: api.authURL(`/entities/${location.value!.id}/attachments/${cur.id}`),
          originalType: cur.mimeType,
          attachmentId: cur.id,
        };
        if (cur.thumbnail) {
          photo.thumbnailSrc = api.authURL(`/entities/${location.value!.id}/attachments/${cur.thumbnail.id}`);
        } else {
          photo.thumbnailSrc = photo.originalSrc;
        }
        acc.push(photo);
      }
      return acc;
    }, [] as Photo[]);
  });

  function openImageDialog(img: Photo, entityId: string) {
    openDialog(DialogID.ItemImage, {
      params: {
        type: "preloaded",
        originalSrc: img.originalSrc,
        originalType: img.originalType,
        thumbnailSrc: img.thumbnailSrc,
        attachmentId: img.attachmentId,
        itemId: entityId,
      },
      onClose: result => {
        if (result?.action === "delete") {
          location.value!.attachments = location.value!.attachments.filter(a => a.id !== result.id);
        }
      },
    });
  }

  // Attachments (non-photo)
  const nonPhotoAttachments = computed(() => {
    if (!location.value?.attachments) {
      return { attachments: [], warranty: [], manuals: [], receipts: [] };
    }
    return location.value.attachments.reduce(
      (acc, attachment) => {
        if (attachment.type === "photo") return acc;
        if (attachment.type === "warranty") acc.warranty.push(attachment);
        else if (attachment.type === "manual") acc.manuals.push(attachment);
        else if (attachment.type === "receipt") acc.receipts.push(attachment);
        else acc.attachments.push(attachment);
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

  const hasNonPhotoAttachments = computed(() => {
    const a = nonPhotoAttachments.value;
    return a.attachments.length > 0 || a.warranty.length > 0 || a.manuals.length > 0 || a.receipts.length > 0;
  });

  // Details
  const locationDetails = computed<Details>(() => {
    if (!location.value) {
      return [];
    }

    const ret: Details = [
      {
        name: "items.notes",
        type: "markdown",
        text: location.value.notes,
      },
      ...(location.value.fields || []).map(field => {
        return {
          name: field.name,
          text: field.textValue,
        } as AnyDetail;
      }),
    ];

    if (!preferences.value.showEmpty) {
      return filterZeroValues(ret);
    }

    return ret;
  });

  // Keep the original direct-only listing unchanged. Opting into a child depth
  // uses a dedicated, paginated backend query that follows physical placement
  // overrides and returns each item's storage path.
  const depth = computed<LocationItemDepth>(() => {
    const saved = preferences.value.locationItemDepth;
    return ["direct", "1", "2", "3", "all"].includes(saved) ? saved : "direct";
  });
  const locationPage = ref(1);
  const locationPageSize = computed(() => Math.min(96, Math.max(1, preferences.value.itemsPerTablePage || 24)));
  const descendantTotal = ref(0);

  function setDepth(next: unknown) {
    if (["direct", "1", "2", "3", "all"].includes(String(next))) {
      preferences.value.locationItemDepth = String(next) as LocationItemDepth;
    }
  }

  watch([locationId, depth, locationPageSize], () => {
    locationPage.value = 1;
  });

  const itemListKey = computed(
    () => locationId.value + "_item_list_" + depth.value + "_" + locationPage.value + "_" + locationPageSize.value
  );
  const {
    data: itemList,
    status: itemListStatus,
    refresh: refreshItemList,
  } = useAsyncData(
    () => itemListKey.value,
    async () => {
      // Snapshot the parameters associated with this request so stale
      // responses cannot be mistaken for the currently selected page.
      const requestKey = itemListKey.value;
      const requestedLocation = locationId.value;
      const requestedDepth = depth.value;
      const requestedPage = locationPage.value;
      const requestedPageSize = locationPageSize.value;
      if (!requestedLocation) {
        return {
          requestKey,
          items: [] as LocationItem[],
          total: 0,
          directTotal: 0,
        };
      }

      if (requestedDepth === "direct") {
        const resp = await api.items.getAll({ parentIds: [requestedLocation] });
        if (resp.error || !resp.data) {
          toast.error(t("items.toast.failed_load_items"));
          return {
            requestKey,
            items: [] as LocationItem[],
            total: 0,
            directTotal: 0,
          };
        }
        return {
          requestKey,
          items: resp.data.items as LocationItem[],
          total: resp.data.items.length,
          directTotal: resp.data.items.length,
        };
      }

      const response = await api.items.getLocationDescendants(
        requestedLocation,
        requestedDepth as LocationDescendantDepth,
        requestedPage,
        requestedPageSize
      );
      if (response.error || !response.data) {
        toast.error(t("items.toast.failed_load_items"));
        return {
          requestKey,
          items: [] as LocationItem[],
          total: 0,
          directTotal: 0,
        };
      }
      return { ...response.data, requestKey };
    },
    { watch: [locationId, depth, locationPage, locationPageSize] }
  );

  // Nuxt can retain the prior async-data result while a new key loads.
  // Only permit currently loaded rows/counts into a selectable location view.
  const activeItemList = computed(() =>
    isCurrentLocationPage(itemListStatus.value, itemListKey.value, itemList.value?.requestKey)
      ? itemList.value
      : undefined
  );

  const items = computed(() => {
    const rows = activeItemList.value?.items ?? [];
    if (depth.value === "direct") return rows;
    // Even with an omitted/empty path, the scoped marker prevents displaying
    // a logical parent from another room as this item's physical location.
    return rows.map(item => withScopedLocationPath(item, preferences.value.showLocationPaths));
  });

  watch(
    () => activeItemList.value?.total,
    total => {
      descendantTotal.value = total ?? 0;
      // A delete/move can shrink the final page. Jump to the last valid page
      // and let the page-sensitive async-data key fetch its contents.
      if (depth.value !== "direct" && total !== undefined) {
        const validPage = clampPageToTotal(locationPage.value, total, locationPageSize.value);
        if (validPage !== locationPage.value) locationPage.value = validPage;
      }
    },
    { immediate: true }
  );

  const directItemTotal = computed(() => activeItemList.value?.directTotal ?? 0);
  const nestedItemTotal = computed(() =>
    Math.max(0, (activeItemList.value?.total ?? 0) - (activeItemList.value?.directTotal ?? 0))
  );

  const descendantPagination = proxyRefs({
    page: computed({
      get: () => locationPage.value,
      set: (value: number) => {
        locationPage.value = value;
      },
    }),
    pageSize: locationPageSize,
    totalSize: descendantTotal,
    setPage: (newPage: number) => {
      locationPage.value = Math.max(1, newPage);
    },
  });
</script>

<template>
  <div>
    <ItemImageDialog />

    <div v-if="location">
      <!-- set page title -->
      <Title>{{ location.name }}</Title>

      <!-- Photo gallery -->
      <section v-if="photos.length > 0" class="mb-4">
        <div class="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4">
          <button
            v-for="(photo, i) in photos"
            :key="i"
            class="group relative aspect-1 h-32 overflow-hidden rounded-lg border bg-muted"
            @click="openImageDialog(photo, location.id)"
          >
            <img
              :src="photo.thumbnailSrc || photo.originalSrc"
              :alt="location.name"
              class="size-full object-cover transition-transform duration-200 group-hover:scale-105"
            />
          </button>
        </div>
      </section>

      <Card class="p-3">
        <header :class="{ 'mb-2': location?.description }">
          <div class="flex flex-wrap items-end gap-2">
            <div
              class="mb-auto flex size-12 items-center justify-center rounded-full bg-secondary text-secondary-foreground"
            >
              <MdiPackageVariant class="size-7" />
            </div>
            <div>
              <Breadcrumb v-if="location?.parent">
                <BreadcrumbList>
                  <BreadcrumbItem>
                    <BreadcrumbLink as-child class="text-foreground/70 hover:underline">
                      <NuxtLink :to="`/location/${location.parent.id}`">
                        {{ location.parent.name }}
                      </NuxtLink>
                    </BreadcrumbLink>
                  </BreadcrumbItem>
                  <BreadcrumbSeparator />
                  <BreadcrumbItem> {{ location.name }} </BreadcrumbItem>
                </BreadcrumbList>
              </Breadcrumb>
              <h1 class="flex items-center gap-3 pb-1 text-2xl">
                {{ location ? location.name : "" }}

                <Badge v-if="location && location.totalPrice" variant="secondary">
                  <Currency :amount="location.totalPrice" />
                </Badge>
              </h1>
              <div class="flex flex-wrap gap-1 text-xs">
                <div>
                  {{ $t("global.created") }}
                  <DateTime :date="location?.createdAt" />
                </div>
              </div>
              <div v-if="location.tags && location.tags.length > 0" class="mt-2 flex flex-wrap gap-1">
                <TagChip v-for="tag in location.tags" :key="tag.id" :tag="tag" size="sm" />
              </div>
            </div>
            <div class="ml-auto mt-2 flex flex-wrap items-center justify-between gap-2">
              <LabelMaker :id="location.id" type="location" />
              <Button class="w-9 md:w-auto" @click="openCreateItem">
                <MdiPlus name="mdi-plus" />
                <span class="hidden md:inline">
                  {{ $t("components.location.create_item") }}
                </span>
              </Button>
              <Button class="w-9 md:w-auto" @click="goToEdit">
                <MdiPencil name="mdi-pencil" />
                <span class="hidden md:inline">
                  {{ $t("global.edit") }}
                </span>
              </Button>
              <Button variant="destructive" class="w-9 md:w-auto" @click="confirmDelete()">
                <MdiDelete name="mdi-delete" />
                <span class="hidden md:inline">
                  {{ $t("global.delete") }}
                </span>
              </Button>
            </div>
          </div>
        </header>
        <Separator v-if="location && location.description" />
        <Markdown v-if="location && location.description" class="mt-3 text-base" :source="location.description" />
      </Card>

      <!-- Details (notes, custom fields) -->
      <BaseCard v-if="locationDetails.length > 0" class="mt-4">
        <template #title> {{ $t("global.details") }} </template>
        <DetailsSection :details="locationDetails" />
      </BaseCard>

      <!-- Attachments (non-photo) -->
      <BaseCard v-if="hasNonPhotoAttachments" class="mt-4">
        <template #title> {{ $t("items.attachments") }} </template>
        <div class="border-t px-4 py-2">
          <ItemAttachmentsList
            v-if="nonPhotoAttachments.attachments.length > 0"
            :attachments="nonPhotoAttachments.attachments"
            :item-id="location.id"
          />
          <ItemAttachmentsList
            v-if="nonPhotoAttachments.warranty.length > 0"
            :attachments="nonPhotoAttachments.warranty"
            :item-id="location.id"
          />
          <ItemAttachmentsList
            v-if="nonPhotoAttachments.manuals.length > 0"
            :attachments="nonPhotoAttachments.manuals"
            :item-id="location.id"
          />
          <ItemAttachmentsList
            v-if="nonPhotoAttachments.receipts.length > 0"
            :attachments="nonPhotoAttachments.receipts"
            :item-id="location.id"
          />
        </div>
      </BaseCard>

      <!-- Items in this location, optionally including child-location descendants -->
      <section v-if="location" class="mt-4 space-y-3">
        <div class="flex flex-wrap items-end gap-3 rounded-lg border p-3" data-location-depth-controls>
          <div class="flex min-w-56 flex-col gap-1">
            <label for="location-depth-select" class="text-sm font-medium">
              {{ $t("locations.item_depth_label") }}
            </label>
            <Select :model-value="depth" @update:model-value="setDepth">
              <SelectTrigger id="location-depth-select" data-location-depth-select class="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="direct">{{ $t("locations.item_depth_direct") }}</SelectItem>
                <SelectItem value="1">{{ $t("locations.item_depth_1") }}</SelectItem>
                <SelectItem value="2">{{ $t("locations.item_depth_2") }}</SelectItem>
                <SelectItem value="3">{{ $t("locations.item_depth_3") }}</SelectItem>
                <SelectItem value="all">{{ $t("locations.item_depth_all") }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div v-if="depth !== 'direct'" class="flex items-center gap-2 pb-2">
            <Switch id="location-show-paths" v-model="preferences.showLocationPaths" />
            <label for="location-show-paths" class="text-sm">
              {{ $t("locations.show_item_paths") }}
            </label>
          </div>
          <p
            v-if="depth !== 'direct' && activeItemList"
            data-location-depth-count
            class="ml-auto text-xs text-muted-foreground"
          >
            {{
              $t("locations.direct_item_count", {
                direct: directItemTotal,
                nested: nestedItemTotal,
              })
            }}
          </p>
        </div>
        <ItemViewSelectable
          :key="'location-items-' + locationId + '-' + depth + '-' + locationPage + '-' + locationPageSize"
          :items="items"
          :pagination="depth === 'direct' ? undefined : descendantPagination"
          :disable-sort="depth !== 'direct'"
          @refresh="refreshItemList"
        />
      </section>

      <!-- Child locations -->
      <section v-if="location && location.children && location.children.length > 0" class="mt-6">
        <BaseSectionHeader class="mb-5"> {{ $t("locations.child_locations") }} </BaseSectionHeader>
        <div class="grid grid-cols-1 gap-2 sm:grid-cols-3">
          <LocationCard v-for="child in location.children" :key="child.id" :location="child" />
        </div>
      </section>
    </div>
  </div>
</template>
