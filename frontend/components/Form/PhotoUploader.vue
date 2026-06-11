<template>
  <div class="w-full">
    <div class="flex w-full flex-col gap-1.5">
      <Label for="photo-uploader" class="flex w-full px-1">
        {{ label }}
      </Label>

      <div class="relative inline-block">
        <Button type="button" variant="outline" class="w-full" aria-hidden="true" @click.prevent="openFilePicker">
          {{ buttonLabel }}
        </Button>
        <Input
          id="photo-uploader"
          ref="fileInput"
          class="absolute left-0 top-0 size-full cursor-pointer opacity-0"
          type="file"
          accept="image/png,image/jpeg,image/gif,image/avif,image/webp,android/force-camera-workaround"
          multiple
          @change="onFilesSelected"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from "vue";
  import { useI18n } from "vue-i18n";
  import { Label } from "~/components/ui/label";
  import { Input } from "~/components/ui/input";
  import { Button } from "~/components/ui/button";
  import { filesToPhotoPreviews, type PhotoPreview } from "./photo-uploader";

  const props = withDefaults(
    defineProps<{
      label?: string;
      buttonLabel?: string;
      existingCount?: number;
    }>(),
    {
      label: undefined,
      buttonLabel: undefined,
      existingCount: 0,
    }
  );

  const emit = defineEmits<{
    (e: "selected", photos: PhotoPreview[]): void;
  }>();

  const { t } = useI18n();
  const fileInput = ref<HTMLInputElement | null>(null);

  const label = computed(() => props.label || t("components.item.create_modal.item_photo"));
  const buttonLabel = computed(() => props.buttonLabel || t("components.item.create_modal.upload_photos"));

  function openFilePicker() {
    fileInput.value?.click();
  }

  async function onFilesSelected(event: Event) {
    const input = event.target as HTMLInputElement;
    if (!input.files || input.files.length === 0) return;

    const photos = await filesToPhotoPreviews(input.files, props.existingCount);

    emit("selected", photos);
    input.value = "";
  }
</script>
