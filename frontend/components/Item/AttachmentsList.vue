<template>
  <ul role="list" class="divide-y divide-gray-400 rounded-md border border-gray-400">
    <li
      v-for="attachment in attachments"
      :key="attachment.id"
      class="flex items-center justify-between py-3 pl-3 pr-4 text-sm"
    >
      <div class="flex w-0 flex-1 items-center">
        <MdiPaperclip class="size-5 shrink-0 text-gray-400" aria-hidden="true" />
        <span class="ml-2 w-0 flex-1 truncate"> {{ attachment.title }}</span>
      </div>
      <div class="ml-4 shrink-0">
        <a class="tooltip mr-2" data-tip="Download" :href="attachmentURL(attachment.id)" target="_blank">
          <MdiDownload class="size-5" />
        </a>
        <a class="tooltip" data-tip="Open" :href="attachmentURL(attachment.id)" target="_blank">
          <MdiOpenInNew class="size-5" />
        </a>
      </div>
    </li>
  </ul>
</template>

<script setup lang="ts">
  import type { ItemAttachment } from "~~/lib/api/types/data-contracts";
  import MdiPaperclip from "~icons/mdi/paperclip";
  import MdiDownload from "~icons/mdi/download";
  import MdiOpenInNew from "~icons/mdi/open-in-new";

  const props = defineProps({
    attachments: {
      type: Object as () => ItemAttachment[],
      required: true,
    },
    itemId: {
      type: String,
      required: true,
    },
  });

  const api = useUserApi();

  function attachmentURL(attachmentId: string) {
    return api.authURL(`/items/${props.itemId}/attachments/${attachmentId}`);
  }
</script>

<style scoped></style>
