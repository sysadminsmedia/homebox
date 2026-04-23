<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import type { AnyDetail, Details } from "~~/components/global/DetailsSection/types";
  import { filterZeroValues } from "~~/components/global/DetailsSection/types";
  import type { ItemAttachment } from "~~/lib/api/types/data-contracts";
  import { useLocationStore } from "~~/stores/locations";
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
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import BaseCard from "@/components/Base/Card.vue";
  import Currency from "~/components/global/Currency.vue";
  import DateTime from "~/components/global/DateTime.vue";
  import LabelMaker from "~/components/global/LabelMaker.vue";
  import Markdown from "~/components/global/Markdown.vue";
  import DetailsSection from "~/components/global/DetailsSection/DetailsSection.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import ItemViewSelectable from "~/components/Item/View/Selectable.vue";
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

  const { data: location, refresh } = useAsyncData(locationId.value, async () => {
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
    openDialog(DialogID.CreateItem);
  }

  function goToEdit() {
    navigateTo(`/location/${locationId.value}/edit`);
  }

  const locationStore = useLocationStore();

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
      { attachments: [] as ItemAttachment[], warranty: [] as ItemAttachment[], manuals: [] as ItemAttachment[], receipts: [] as ItemAttachment[] }
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
      ...((location.value.fields || []).map(field => {
        return {
          name: field.name,
          text: field.textValue,
        } as AnyDetail;
      })),
    ];

    if (!preferences.value.showEmpty) {
      return filterZeroValues(ret);
    }

    return ret;
  });

  const { data: items, refresh: refreshItemList } = useAsyncData(
    () => locationId.value + "_item_list",
    async () => {
      if (!locationId.value) {
        return [];
      }

      const resp = await api.items.getAll({
        parentIds: [locationId.value],
      });

      if (resp.error) {
        toast.error(t("items.toast.failed_load_items"));
        return [];
      }

      return resp.data.items;
    },
    {
      watch: [locationId],
    }
  );
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
            class="group relative aspect-square overflow-hidden rounded-lg border bg-muted"
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
              <Breadcrumb v-if="location?.parent" data-testid="location-breadcrumb">
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
              <h1 class="flex items-center gap-3 pb-1 text-2xl" data-testid="location-detail-name">
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
                  {{ $t("components.item.create_modal.title") }}
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

      <!-- Items in this location -->
      <section v-if="location && items">
        <ItemViewSelectable :items="items" @refresh="refreshItemList" />
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
