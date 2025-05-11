<template>
  <BaseModal dialog-id="create-item" :title="$t('components.item.create_modal.title')" >
    <form class="flex flex-col gap-2" @submit.prevent="create()">
      <LocationSelector v-model="form.location" />
      <ItemSelector 
              :label="$t('components.item.create_modal.parent_item')" 
              v-model="parent" 
              v-if="subItemCreate"
              v-model:search="query"
              :items="results"
              item-text="name"
              no-results-text="Type to search..."
             />
      <FormTextField
        ref="nameInput"
        v-model="form.name"
        :trigger-focus="focused"
        :autofocus="true"
        :label="$t('components.item.create_modal.item_name')"
        :max-length="255"
        :min-length="1"
      />
      <FormTextField v-model="form.quantity" :label="$t('components.item.create_modal.item_quantity')" type="number" />
      <FormTextArea
        v-model="form.description"
        :label="$t('components.item.create_modal.item_description')"
        :max-length="1000"
      />
      <LabelSelector v-model="form.labels" :labels="labels ?? []" />
      <div class="flex w-full flex-col gap-1.5">
        <Label for="image-create-photo" class="flex w-full px-1">
          {{ $t("components.item.create_modal.item_photo") }}
        </Label>
        <div class="relative inline-block">
          <Button type="button" variant="outline" class="w-full" aria-hidden="true" @click.prevent="">
            {{ $t("components.item.create_modal.upload_photos") }}
          </Button>
          <Input
            id="image-create-photo"
            ref="fileInput"
            class="absolute left-0 top-0 size-full cursor-pointer opacity-0"
            type="file"
            accept="image/png,image/jpeg,image/gif,image/avif,image/webp;capture=camera"
            multiple
            @change="previewImage"
          />
        </div>
      </div>
      <div class="mt-4 flex flex-row-reverse">
        <ButtonGroup>
          <Button :disabled="loading" type="submit" class="group">
            <div class="relative mx-2">
              <div
                class="absolute inset-0 flex items-center justify-center transition-transform duration-300 group-hover:rotate-[360deg]"
              >
                <MdiPackageVariant class="size-5 group-hover:hidden" />
                <MdiPackageVariantClosed class="hidden size-5 group-hover:block" />
              </div>
            </div>
            {{ $t("global.create") }}
          </Button>
          <Button variant="outline" :disabled="loading" type="button" @click="create(false)">
            {{ $t("global.create_and_add") }}
          </Button>
        </ButtonGroup>
      </div>

      <!-- photo preview area is AFTER the create button, to avoid pushing the button below the screen on small displays -->
      <div v-if="form.photos.length > 0" class="mt-4 border-t px-4 pb-4">
        <div v-for="(photo, index) in form.photos" :key="index">
          <div class="mt-8 w-full">
            <img :src="photo.fileBase64" class="w-full rounded object-fill shadow-sm" alt="Uploaded Photo" />
          </div>
          <div class="mt-2 flex items-center gap-2">
            <TooltipProvider class="flex gap-2" :delay-duration="0">
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
              <Tooltip>
                <TooltipTrigger>
                  <Button
                    size="icon"
                    type="button"
                    variant="default"
                    @click.prevent="
                      async () => {
                        await rotateBase64Image90Deg(photo.fileBase64, index);
                      }
                    "
                  >
                    <MdiRotateClockwise />
                    <div class="sr-only">Rotate photo</div>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Rotate photo</p>
                </TooltipContent>
              </Tooltip>
              <Tooltip>
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
              </Tooltip>
            </TooltipProvider>
            <p class="mt-1 text-sm" style="overflow-wrap: anywhere">{{ photo.photoName }}</p>
          </div>
        </div>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import { Button, ButtonGroup } from "~/components/ui/button";
  import BaseModal from "@/components/App/CreateModal.vue";
  import { Label } from "@/components/ui/label";
  import { Input } from "@/components/ui/input";
  import type { ItemCreate, LocationOut, ItemOut } from "~~/lib/api/types/data-contracts";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPackageVariantClosed from "~icons/mdi/package-variant-closed";
  import MdiDelete from "~icons/mdi/delete";
  import MdiRotateClockwise from "~icons/mdi/rotate-clockwise";
  import MdiStarOutline from "~icons/mdi/star-outline";
  import MdiStar from "~icons/mdi/star";
  import { AttachmentTypes } from "~~/lib/api/types/non-generated";
  import { useDialog, useDialogHotkey } from "~/components/ui/dialog-provider";
  import LabelSelector from "~/components/Label/Selector.vue";
  import ItemSelector from "~/components/Item/Selector.vue";

  interface PhotoPreview {
    photoName: string;
    file: File;
    fileBase64: string;
    primary: boolean;
  }

  const { activeDialog, closeDialog } = useDialog();

  useDialogHotkey("create-item", { code: "Digit1", shift: true });

  const api = useUserApi();

  const locationsStore = useLocationStore();
  const locations = computed(() => locationsStore.allLocations);

  const labelStore = useLabelStore();
  const labels = computed(() => labelStore.labels);

  const route = useRoute();
  const router = useRouter();

  const parent = ref();
  const { query, results } = useItemSearch(api, { immediate: false });
  const subItemCreateParam = useRouteQuery("subItemCreate", false);
  const subItemCreate = ref();
  

  const labelId = computed(() => {
    if (route.fullPath.includes("/label/")) {
      return route.params.id;
    }
    return null;
  });

  const locationId = computed(() => {
    if (route.fullPath.includes("/location/")) {
      return route.params.id;
    }
    return null;
  });

  const itemId = computed(() => {
    if (route.fullPath.includes("/item/")) {
      return route.params.id;
    }
    return null;
  });

  const nameInput = ref<HTMLInputElement | null>(null);

  const loading = ref(false);
  const focused = ref(false);
  const form = reactive({
    location: locations.value && locations.value.length > 0 ? locations.value[0] : ({} as LocationOut),
    parentId: parent.value && parent.value.id && subItemCreate.value ? parent.value.id : null,
    name: "",
    quantity: 1,
    description: "",
    color: "",
    labels: [] as string[],
    photos: [] as PhotoPreview[],
  });

  const { shift } = useMagicKeys();

  function deleteImage(index: number) {
    form.photos.splice(index, 1);
  }

  function setPrimary(index: number) {
    const primary = form.photos.findIndex(p => p.primary);

    if (primary !== -1) form.photos[primary].primary = false;
    if (primary !== index) form.photos[index].primary = true;
  }

  function previewImage(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      for (const file of input.files) {
        const reader = new FileReader();
        reader.onload = e => {
          form.photos.push({
            photoName: file.name,
            fileBase64: e.target?.result as string,
            file,
            primary: form.photos.length === 0,
          });
        };
        reader.readAsDataURL(file);
      }
      input.value = "";
    }
  }

  watch(
    () => activeDialog.value,
    async active => {
      if (active === "create-item") {
        subItemCreate.value = subItemCreateParam.value;
        let parentItemLocationId = null;
        
        if (subItemCreate.value && itemId.value){
          const { data, error } = await api.items.get(itemId.value);
          if (error) {
            toast.error("Failed to load parent item - please select manually");
          }
          
          parentItemLocationId = data.location.id;
          parent.value = data; 
          form.parentId = data.id;
          await router.push({query: {},});
        }
        
        const locId = locationId.value ? locationId.value : parentItemLocationId;

        if (locId) {
          const found = locations.value.find(l => l.id === locId);
          if (found) {
            form.location = found;
          }
        }
        if (labelId.value) {
          form.labels = labels.value.filter(l => l.id === labelId.value).map(l => l.id);
        }
      }
    }
  );

  async function create(close = true) {
    if (!form.location?.id) {
      toast.error("Please select a location.");
      return;
    }

    if (loading.value) {
      toast.error("Already creating an item");
      return;
    }

    loading.value = true;

    if (shift.value) close = false;

    const out: ItemCreate = {
      parentId: form.parentId,
      name: form.name,
      quantity: form.quantity,
      description: form.description,
      locationId: form.location.id as string,
      labelIds: form.labels,
    };

    const { error, data } = await api.items.create(out);

    if (error) {
      loading.value = false;
      toast.error("Couldn't create item");
      return;
    }

    toast.success("Item created");

    if (form.photos.length > 0) {
      toast.info(`Uploading ${form.photos.length} photo(s)...`);
      let uploadError = false;
      for (const photo of form.photos) {
        const { error: attachError } = await api.items.attachments.add(
          data.id,
          photo.file,
          photo.photoName,
          AttachmentTypes.Photo,
          photo.primary
        );

        if (attachError) {
          uploadError = true;
          toast.error(`Failed to upload Photo: ${photo.photoName}`);
          console.error(attachError);
        }
      }
      if (uploadError) {
        toast.warning("Some photos failed to upload.");
      } else {
        toast.success("All photos uploaded successfully.");
      }
    }

    form.name = "";
    form.quantity = 1;
    form.description = "";
    form.color = "";
    form.photos = [];
    form.labels = [];
    focused.value = false;
    loading.value = false;

    if (close) {
      closeDialog("create-item");
      navigateTo(`/item/${data.id}`);
    }
  }

  function dataURLtoFile(dataURL: string, fileName: string) {
    try {
      const arr = dataURL.split(",");
      const mimeMatch = arr[0].match(/:(.*?);/);
      if (!mimeMatch || !mimeMatch[1]) {
        throw new Error("Invalid data URL format");
      }
      const mime = mimeMatch[1];

      // Validate mime type is an image
      if (!mime.startsWith("image/")) {
        throw new Error("Invalid mime type, expected image");
      }

      const bstr = atob(arr[arr.length - 1]);
      let n = bstr.length;
      const u8arr = new Uint8Array(n);
      while (n--) {
        u8arr[n] = bstr.charCodeAt(n);
      }
      return new File([u8arr], fileName, { type: mime });
    } catch (error) {
      console.error("Error converting data URL to file:", error);
      // Return a fallback or rethrow based on your error handling strategy
      throw error;
    }
  }

  async function rotateBase64Image90Deg(base64Image: string, index: number) {
    // Create an off-screen canvas
    const offScreenCanvas = document.createElement("canvas");
    const offScreenCanvasCtx = offScreenCanvas.getContext("2d");

    if (!offScreenCanvasCtx) {
      toast.error("Your browser doesn't support canvas operations");
      return;
    }

    // Create an image
    const img = new Image();

    // Create a promise to handle the image loading
    await new Promise<void>((resolve, reject) => {
      img.onload = () => resolve();
      img.onerror = () => reject(new Error("Failed to load image"));
      img.src = base64Image;
    }).catch(error => {
      toast.error("Failed to rotate image: " + error.message);
    });

    // Set its dimensions to rotated size
    offScreenCanvas.height = img.width;
    offScreenCanvas.width = img.height;

    // Rotate and draw source image into the off-screen canvas
    offScreenCanvasCtx.rotate((90 * Math.PI) / 180);
    offScreenCanvasCtx.translate(0, -offScreenCanvas.width);
    offScreenCanvasCtx.drawImage(img, 0, 0);

    const imageType = base64Image.match(/^data:(.+);base64/)?.[1] || "image/jpeg";

    // Encode image to data-uri with base64
    try {
      form.photos[index].fileBase64 = offScreenCanvas.toDataURL(imageType, 100);
      form.photos[index].file = dataURLtoFile(form.photos[index].fileBase64, form.photos[index].photoName);
    } catch (error) {
      toast.error("Failed to process rotated image");
      console.error(error);
    } finally {
      // Clean up resources
      offScreenCanvas.width = 0;
      offScreenCanvas.height = 0;
    }
  }
</script>
