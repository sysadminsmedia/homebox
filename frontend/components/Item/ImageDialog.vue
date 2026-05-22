<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { Dialog, DialogContent } from "@/components/ui/dialog";
  import { Button, buttonVariants } from "@/components/ui/button";
  import { useDialog } from "@/components/ui/dialog-provider";
  import type { ItemAttachment } from "~~/lib/api/types/data-contracts";
  import { AttachmentTypes } from "~~/lib/api/types/non-generated";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import { blobToDataUrl, dataUrlToFile, rotateImageDataUrl90Deg } from "~/composables/utils";
  import { useConfirm } from "@/composables/use-confirm";
  import { toast } from "@/components/ui/sonner";
  import MdiClose from "~icons/mdi/close";
  import MdiRotateClockwise from "~icons/mdi/rotate-clockwise";
  import MdiDownload from "~icons/mdi/download";
  import MdiDelete from "~icons/mdi/delete";

  const { t } = useI18n();
  const confirm = useConfirm();

  const { closeDialog, registerOpenDialogCallback } = useDialog();

  const api = useUserApi();

  const image = reactive<{
    attachmentId: string;
    busy: boolean;
    dirty: boolean;
    itemId: string;
    originalSrc: string;
    originalType?: string;
    thumbnailSrc?: string;
    workingDataUrl: string;
    workingSrc: string;
  }>({
    attachmentId: "",
    busy: false,
    dirty: false,
    itemId: "",
    originalSrc: "",
    originalType: undefined,
    thumbnailSrc: undefined,
    workingDataUrl: "",
    workingSrc: "",
  });

  const displaySrc = computed(() => image.workingSrc || image.thumbnailSrc || image.originalSrc || "");
  const displayType = computed(() => image.workingDataUrl.match(/^data:(.+);base64/)?.[1] || image.originalType);

  function resetImageState() {
    image.attachmentId = "";
    image.busy = false;
    image.dirty = false;
    image.itemId = "";
    image.originalSrc = "";
    image.originalType = undefined;
    image.thumbnailSrc = undefined;
    image.workingDataUrl = "";
    image.workingSrc = "";
  }

  function findNewAttachment(attachments: ItemAttachment[], existingAttachmentIds: string[]) {
    const knownIds = new Set(existingAttachmentIds);
    const candidates = attachments.filter(attachment => {
      return (
        attachment.id !== image.attachmentId &&
        attachment.type === AttachmentTypes.Photo &&
        !knownIds.has(attachment.id)
      );
    });

    const attachmentsToSearch =
      candidates.length > 0
        ? candidates
        : attachments.filter(
            attachment => attachment.id !== image.attachmentId && attachment.type === AttachmentTypes.Photo
          );

    return attachmentsToSearch.sort((left, right) => {
      return new Date(right.createdAt).getTime() - new Date(left.createdAt).getTime();
    })[0];
  }

  async function getCurrentAttachmentState() {
    const { data, error } = await api.items.get(image.itemId);
    if (error || !data) {
      throw new Error("Failed to load attachment state");
    }

    const currentAttachment = data.attachments.find(attachment => attachment.id === image.attachmentId);
    if (!currentAttachment) {
      throw new Error("Attachment not found");
    }

    return {
      attachment: currentAttachment,
      existingAttachmentIds: data.attachments.map(attachment => attachment.id),
    };
  }

  async function ensureWorkingImageLoaded() {
    if (image.workingDataUrl) {
      return;
    }

    if (!image.originalSrc) {
      throw new Error("Missing attachment source");
    }

    const response = await fetch(image.originalSrc);
    if (!response.ok) {
      throw new Error(`Failed to fetch image: ${response.status}`);
    }

    image.workingDataUrl = await blobToDataUrl(await response.blob());
    image.workingSrc = image.workingDataUrl;
  }

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.ItemImage, params => {
      image.attachmentId = params.attachmentId;
      image.busy = false;
      image.dirty = false;
      image.itemId = params.itemId;
      image.workingDataUrl = "";
      image.workingSrc = "";

      if (params.type === "preloaded") {
        image.originalSrc = params.originalSrc;
        image.originalType = params.originalType;
        image.thumbnailSrc = params.thumbnailSrc;
      } else {
        image.originalSrc = api.authURL(`/entities/${params.itemId}/attachments/${params.attachmentId}`);
        image.originalType = params.mimeType;
        image.thumbnailSrc = params.thumbnailId
          ? api.authURL(`/entities/${params.itemId}/attachments/${params.thumbnailId}`)
          : image.originalSrc;
      }
    });

    onUnmounted(cleanup);
  });

  async function rotateAttachment() {
    if (image.busy) {
      return;
    }

    image.busy = true;

    try {
      await ensureWorkingImageLoaded();
      image.workingDataUrl = await rotateImageDataUrl90Deg(image.workingDataUrl);
      image.workingSrc = image.workingDataUrl;
      image.dirty = true;
    } catch (error) {
      toast.error(t("items.toast.failed_update_attachment"));
      console.error(error);
    } finally {
      image.busy = false;
    }
  }

  async function persistRotatedAttachmentAndClose() {
    if (!image.workingDataUrl) {
      resetImageState();
      closeDialog(DialogID.ItemImage);
      return;
    }

    const { attachment: currentAttachment, existingAttachmentIds } = await getCurrentAttachmentState();
    const fileName = currentAttachment.title || `attachment-${image.attachmentId}.jpg`;
    const file = dataUrlToFile(image.workingDataUrl, fileName);
    const { data, error } = await api.items.attachments.add(
      image.itemId,
      file,
      file.name,
      AttachmentTypes.Photo,
      currentAttachment.primary
    );

    if (error || !data) {
      toast.error(t("items.toast.failed_update_attachment"));
      return;
    }

    const createdAttachment = findNewAttachment(data.attachments, existingAttachmentIds);
    if (!createdAttachment) {
      toast.error(t("items.toast.failed_update_attachment"));
      return;
    }

    const { error: deleteError } = await api.items.attachments.delete(image.itemId, image.attachmentId);
    if (deleteError) {
      const { error: rollbackError } = await api.items.attachments.delete(image.itemId, createdAttachment.id);
      if (rollbackError) {
        console.error("Failed to rollback rotated attachment", rollbackError);
      }
      toast.error(t("items.toast.failed_update_attachment"));
      return;
    }

    const { data: refreshedData, error: refreshedError } = await api.items.get(image.itemId);
    const refreshedAttachment = refreshedData?.attachments.find(attachment => attachment.id === createdAttachment.id);
    if (refreshedError || !refreshedAttachment) {
      toast.error(t("items.toast.failed_update_attachment"));
      return;
    }

    const oldId = image.attachmentId;
    resetImageState();
    closeDialog(DialogID.ItemImage, {
      action: "replace",
      oldId,
      attachment: refreshedAttachment,
    });
    toast.success(t("items.toast.attachment_updated"));
  }

  async function closeImageDialog() {
    if (image.busy) {
      return;
    }

    if (!image.dirty) {
      resetImageState();
      closeDialog(DialogID.ItemImage);
      return;
    }

    image.busy = true;

    try {
      await persistRotatedAttachmentAndClose();
    } catch (error) {
      toast.error(t("items.toast.failed_update_attachment"));
      console.error(error);
    } finally {
      image.busy = false;
    }
  }

  async function deleteAttachment() {
    if (image.busy) {
      return;
    }

    const confirmed = await confirm.open(t("items.delete_attachment_confirm"));

    if (confirmed.isCanceled) {
      return;
    }

    const { error } = await api.items.attachments.delete(image.itemId, image.attachmentId);

    if (error) {
      toast.error(t("items.toast.failed_delete_attachment"));
      return;
    }

    const deletedId = image.attachmentId;
    resetImageState();
    closeDialog(DialogID.ItemImage, {
      action: "delete",
      id: deletedId,
    });
    toast.success(t("items.toast.attachment_deleted"));
  }

  function handlePointerDownOutside(event: CustomEvent<{ originalEvent: PointerEvent }>) {
    event.preventDefault();
    void closeImageDialog();
  }
</script>

<template>
  <Dialog :dialog-id="DialogID.ItemImage">
    <DialogContent
      class="max-h-[90svh] w-auto max-w-[min(calc(100vw_-_1rem),32rem)] border-transparent bg-transparent p-0 md:max-w-lg"
      disable-close
      @pointer-down-outside="handlePointerDownOutside"
    >
      <picture>
        <source :srcset="displaySrc" :type="displayType" />
        <img
          :src="displaySrc"
          alt="attachment image"
          class="min-w-64 max-w-[min(calc(100vw_-_1rem),32rem)] md:w-auto md:max-w-lg"
        />
      </picture>
      <div class="absolute right-1 top-1 flex gap-2">
        <Button variant="destructive" size="icon" :disabled="image.busy" @click="deleteAttachment">
          <MdiDelete />
        </Button>
        <Button size="icon" :disabled="image.busy" @click="rotateAttachment">
          <MdiRotateClockwise />
        </Button>
        <a
          :class="buttonVariants({ size: 'icon' })"
          :href="image.workingSrc || image.originalSrc"
          download
          class="shrink-0"
        >
          <MdiDownload />
        </a>
        <Button size="icon" :disabled="image.busy" @click="closeImageDialog">
          <MdiClose />
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
