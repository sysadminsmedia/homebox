<script setup lang="ts">
  import { ref, onMounted } from "vue";
  import { Cropper } from "vue-advanced-cropper";
  import { Label } from "@/components/ui/label";
  import { Input } from "@/components/ui/input";
  import { Button } from "~/components/ui/button";
  import MdiDelete from "~icons/mdi/delete";
  import MdiRotateLeft from "~icons/mdi/rotate-left";
  import MdiRotateRight from "~icons/mdi/rotate-right";
  import MdiFlipHorizontal from "~icons/mdi/flip-horizontal";
  import MdiFlipVertical from "~icons/mdi/flip-vertical";
  // import MdiStarOutline from "~icons/mdi/star-outline";
  // import MdiStar from "~icons/mdi/star";

  import "vue-advanced-cropper/dist/style.css";

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
  const croppers = ref<(InstanceType<typeof Cropper> | null)[]>([]);

  onMounted(() => {
    croppers.value = Array(photos.value.length).fill(null);
  });

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
    croppers.value.splice(index, 1);
    emits("update:photos", photos.value);
  }

  // function setPrimary(index: number) {
  //   const primary = photos.value.findIndex(p => p.primary);

  //   if (primary !== -1) photos.value[primary].primary = false;
  //   if (primary !== index) photos.value[index].primary = true;

  //   toast.error("Currently this does not do anything, the first photo will always be primary");
  // }

  const setSize = (index: number) => {
    const cropper = croppers.value[index];
    const img = new Image();
    img.src = photos.value[index].fileBase64;
    img.onload = () => {
      // get the image size
      cropper?.setCoordinates({
        width: img.naturalWidth,
        height: img.naturalHeight,
        left: 0,
        top: 0,
      });
    };
  };
</script>

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
    <div v-if="photos.length > 0" class="mt-4 border-t border-gray-300">
      <div v-for="(photo, index) in photos" :key="index">
        <div class="mt-8 w-full">
          <cropper
            ref="croppers"
            :src="photo.fileBase64"
            alt="Uploaded Photo"
            background-class="image-cropper-bg"
            class="image-cropper"
            @ready="
              () => {
                setSize(index);
              }
            "
          />
          <!-- class="w-full rounded border-gray-300 object-fill shadow-sm" -->
        </div>
        <div class="mt-2 flex justify-center gap-2">
          <TooltipProvider class="flex gap-2">
            <Tooltip>
              <TooltipTrigger>
                <Button size="icon" type="button" variant="outline" @click.prevent="croppers[index]?.rotate(-90)">
                  <MdiRotateLeft />
                  <div class="sr-only">Rotate left</div>
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Rotate left</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger>
                <Button size="icon" type="button" variant="outline" @click.prevent="croppers[index]?.flip(true, false)">
                  <MdiFlipHorizontal />
                  <div class="sr-only">Flip horizontal</div>
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Flip horizontal</p>
              </TooltipContent>
            </Tooltip>
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
            <Tooltip>
              <TooltipTrigger>
                <Button size="icon" type="button" variant="outline" @click.prevent="croppers[index]?.flip(false, true)">
                  <MdiFlipVertical />
                  <div class="sr-only">Flip vertical</div>
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Flip vertical</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger>
                <Button size="icon" type="button" variant="outline" @click.prevent="croppers[index]?.rotate(90)">
                  <MdiRotateRight />
                  <div class="sr-only">Rotate right</div>
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Rotate right</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
        <p class="mt-1 text-center text-sm" style="overflow-wrap: anywhere">{{ photo.photoName }}</p>
      </div>
    </div>
  </div>
</template>

<style>
  .image-cropper {
    width: 462px;
  }

  .image-cropper-bg {
    background-color: white;
  }
</style>
