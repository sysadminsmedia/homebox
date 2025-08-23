<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { Dialog, DialogContent } from "@/components/ui/dialog";
  import { buttonVariants, Button } from "@/components/ui/button";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import { useConfirm } from "@/composables/use-confirm";
  import { toast } from "@/components/ui/sonner";
  import MdiClose from "~icons/mdi/close";
  import MdiDownload from "~icons/mdi/download";
  import MdiDelete from "~icons/mdi/delete";

  const { t } = useI18n();
  const confirm = useConfirm();

  const { closeDialog, registerOpenDialogCallback } = useDialog();

  const api = useUserApi();

  const image = reactive<{
    attachmentId: string;
    itemId: string;
    originalSrc: string;
    originalType?: string;
    thumbnailSrc?: string;
  }>({
    attachmentId: "",
    itemId: "",
    originalSrc: "",
  });

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.ItemImage, params => {
      image.attachmentId = params.attachmentId;
      image.itemId = params.itemId;
      if (params.type === "preloaded") {
        image.originalSrc = params.originalSrc;
        image.originalType = params.originalType;
        image.thumbnailSrc = params.thumbnailSrc;
      } else if (params.type === "attachment") {
        image.originalSrc = api.authURL(`/items/${params.itemId}/attachments/${params.attachmentId}`);
        image.originalType = params.mimeType;
        image.thumbnailSrc = params.thumbnailId
          ? api.authURL(`/items/${params.itemId}/attachments/${params.thumbnailId}`)
          : image.originalSrc;
      }
    });

    onUnmounted(cleanup);
  });

  async function deleteAttachment() {
    const confirmed = await confirm.open(t("items.delete_attachment_confirm"));

    if (confirmed.isCanceled) {
      return;
    }

    const { error } = await api.items.attachments.delete(image.itemId, image.attachmentId);

    if (error) {
      toast.error(t("items.toast.failed_delete_attachment"));
      return;
    }

    closeDialog(DialogID.ItemImage, {
      action: "delete",
      id: image.attachmentId,
    });
    toast.success(t("items.toast.attachment_deleted"));
  }
</script>

<template>
  <Dialog :dialog-id="DialogID.ItemImage">
    <DialogContent class="w-auto border-transparent bg-transparent p-0" disable-close>
      <picture>
        <source :srcset="image.originalSrc" :type="image.originalType" />
        <img :src="image.thumbnailSrc" alt="attachment image" />
      </picture>
      <Button variant="destructive" size="icon" class="absolute right-[84px] top-1" @click="deleteAttachment">
        <MdiDelete />
      </Button>
      <a :class="buttonVariants({ size: 'icon' })" :href="image.originalSrc" download class="absolute right-11 top-1">
        <MdiDownload />
      </a>
      <Button
        size="icon"
        class="absolute right-1 top-1"
        @click="
          closeDialog(DialogID.ItemImage);
          image.originalSrc = '';
        "
      >
        <MdiClose />
      </Button>
    </DialogContent>
  </Dialog>
</template>
