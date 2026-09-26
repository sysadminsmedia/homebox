<template>
  <div v-if="photos.length > 0" class="mt-4 border-t px-4 pb-4">
    <div v-for="(photo, index) in photos" :key="index">
      <div class="mt-8 w-full">
        <img
          :src="photo.fileBase64"
          class="w-full rounded object-fill shadow-sm"
          :alt="$t('components.entity.create_modal.uploaded')"
        />
      </div>

      <div class="mt-2 flex items-center gap-2">
        <TooltipProvider class="flex gap-2" :delay-duration="0">
          <Tooltip>
            <TooltipTrigger>
              <Button size="icon" type="button" variant="destructive" @click.prevent="emit('delete', index)">
                <MdiDelete />
                <div class="sr-only">{{ $t("components.entity.create_modal.delete_photo") }}</div>
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>{{ $t("components.entity.create_modal.delete_photo") }}</p>
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger>
              <Button size="icon" type="button" variant="default" @click.prevent="emit('rotate', index)">
                <MdiRotateClockwise />
                <div class="sr-only">{{ $t("components.entity.create_modal.rotate_photo") }}</div>
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>{{ $t("components.entity.create_modal.rotate_photo") }}</p>
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger>
              <Button
                size="icon"
                type="button"
                :variant="photo.primary ? 'default' : 'outline'"
                @click.prevent="emit('setPrimary', index)"
              >
                <MdiStar v-if="photo.primary" />
                <MdiStarOutline v-else />
                <div class="sr-only">
                  {{ $t("components.entity.create_modal.set_as_primary_photo", { isPrimary: photo.primary }) }}
                </div>
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>{{ $t("components.entity.create_modal.set_as_primary_photo", { isPrimary: photo.primary }) }}</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>

        <p class="mt-1 text-sm" style="overflow-wrap: anywhere">{{ photo.photoName }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { Button } from "~/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "~/components/ui/tooltip";
  import MdiDelete from "~icons/mdi/delete";
  import MdiRotateClockwise from "~icons/mdi/rotate-clockwise";
  import MdiStarOutline from "~icons/mdi/star-outline";
  import MdiStar from "~icons/mdi/star";
  import type { PhotoPreview } from "./photo-uploader";

  defineProps<{
    photos: PhotoPreview[];
  }>();

  const emit = defineEmits<{
    (e: "delete", index: number): void;
    (e: "rotate", index: number): void;
    (e: "setPrimary", index: number): void;
  }>();
</script>
