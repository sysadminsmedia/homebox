<template>
  <div>
    <div class="flex w-full flex-col gap-1.5">
      <Label for="image-create-photo" class="flex w-full px-1">
        {{ $t("components.item.create_modal.item_photo") }}
      </Label>
      <div class="relative inline-block">
        <Button type="button" variant="outline" class="w-full" aria-hidden="true">
          {{ $t("components.item.create_modal.upload_photos") }}
        </Button>
        <Input
          id="image-create-photo"
          class="absolute left-0 top-0 size-full cursor-pointer opacity-0"
          type="file"
          accept="image/png,image/jpeg,image/gif,image/avif,image/webp;capture=camera"
          multiple
          @change="handleFileChange"
        />
      </div>
    </div>
    <div v-if="photos.length > 0" class="mt-4 border-t border-gray-300 px-4 pb-4">
      <div v-for="(photo, index) in photos" :key="index">
        <div class="mt-8 w-full">
          <img
            :src="photo.fileBase64"
            class="w-full rounded border-gray-300 object-fill shadow-sm"
            alt="Uploaded Photo"
          />
        </div>
        <div class="mt-2 flex items-center gap-2">
          <TooltipProvider class="flex gap-2">
            <Tooltip>
              <TooltipTrigger>
                <Button size="icon" type="button" variant="destructive" @click.prevent="deleteImage(index)">
                  <MdiDelete />
                  <div class="sr-only">Delete photo</div>
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Delete photo</p>
              </TooltipContent>
            </Tooltip>
            <!-- TODO: re-enable when we have a way to set primary photos -->
            <!-- <Tooltip>
                <TooltipTrigger>
                  <Button
                    size="icon"
                    type="button"
                    :variant="photo.primary ? 'default' : 'outline'"
                    @click.prevent="setPrimary(index)"
                  >
                    <MdiStar v-if="photo.primary" />
                    <MdiStarOutline v-else />
                    <div class="sr-only">Set as {{ photo.primary ? "non" : "" }} primary photo</div>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Set as {{ photo.primary ? "non" : "" }} primary photo</p>
                </TooltipContent>
              </Tooltip> -->
          </TooltipProvider>
          <p class="mt-1 text-sm" style="overflow-wrap: anywhere">{{ photo.photoName }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref } from "vue";
  import { Label } from "@/components/ui/label";
  import { Input } from "@/components/ui/input";
  import { Button } from "~/components/ui/button";
  import MdiDelete from "~icons/mdi/delete";
  // import MdiStarOutline from "~icons/mdi/star-outline";
  // import MdiStar from "~icons/mdi/star";

  export type PhotoPreview = {
    photoName: string;
    file: File;
    fileBase64: string;
    primary: boolean;
  };

  const props = defineProps<{ initialPhotos: PhotoPreview[] }>();
  const emits = defineEmits<{
    (e: "update:photos", photos: PhotoPreview[]): void;
  }>();

  const photos = ref<PhotoPreview[]>(props.initialPhotos);

  function handleFileChange(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      for (const file of input.files) {
        const reader = new FileReader();
        reader.onload = e => {
          const photo = {
            photoName: file.name,
            fileBase64: e.target?.result as string,
            file,
            primary: photos.value.length === 0,
          };
          photos.value.push(photo);
          emits("update:photos", photos.value);
        };
        reader.readAsDataURL(file);
      }
      input.value = "";
    }
  }

  function deleteImage(index: number) {
    photos.value.splice(index, 1);
    emits("update:photos", photos.value);
  }

  // function setPrimary(index: number) {
  //   const primary = photos.value.findIndex(p => p.primary);

  //   if (primary !== -1) photos.value[primary].primary = false;
  //   if (primary !== index) photos.value[index].primary = true;

  //   toast.error("Currently this does not do anything, the first photo will always be primary");
  // }
</script>
