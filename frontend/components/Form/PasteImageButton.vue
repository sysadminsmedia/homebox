<template>
  <Button v-if="supported" type="button" variant="outline" data-testid="paste-image-button" @click.prevent="paste">
    <MdiContentPaste class="mr-1 size-5" />
    {{ $t("global.paste_image") }}
  </Button>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import { Button } from "~/components/ui/button";
  import MdiContentPaste from "~icons/mdi/content-paste";

  const emit = defineEmits<{
    (e: "paste", files: File[]): void;
  }>();

  const { t } = useI18n();

  // Where the async clipboard API is missing, Ctrl+V is still the way in - hide the
  // button rather than offer something that can only fail.
  const supported = canReadClipboardImages();

  async function paste() {
    let images: File[];

    try {
      images = await readClipboardImages();
    } catch {
      toast.error(t("items.toast.clipboard_read_failed"));
      return;
    }

    if (images.length === 0) {
      toast.error(t("items.toast.no_image_in_clipboard"));
      return;
    }

    emit("paste", images);
  }
</script>
