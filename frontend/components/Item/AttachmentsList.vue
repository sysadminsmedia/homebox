<template>
  <ul role="list" class="divide-y rounded-md border">
    <li
      v-for="attachment in attachments"
      :key="attachment.id"
      class="flex items-center justify-between py-3 pl-3 pr-4 text-sm"
    >
      <div class="flex w-0 flex-1 items-center">
        <MdiPaperclip class="size-5 shrink-0 text-gray-400" aria-hidden="true" />
        <span class="ml-2 w-0 flex-1 truncate"> {{ attachment.title }}</span>
      </div>
      <div class="ml-4 flex shrink-0 gap-2">
        <TooltipProvider :delay-duration="0">
          <Tooltip>
            <TooltipTrigger as-child>
              <a
                :class="buttonVariants({ size: 'icon' })"
                :href="attachmentURL(attachment.id)"
                :download="attachment.title"
              >
                <MdiDownload />
              </a>
            </TooltipTrigger>
            <TooltipContent> Download </TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger as-child>
              <a :class="buttonVariants({ size: 'icon' })" :href="attachmentURL(attachment.id)" target="_blank">
                <MdiOpenInNew />
              </a>
            </TooltipTrigger>
            <TooltipContent> Open in new tab </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
    </li>
  </ul>
</template>

<script setup lang="ts">
  import type { ItemAttachment } from "~~/lib/api/types/data-contracts";
  import MdiPaperclip from "~icons/mdi/paperclip";
  import MdiDownload from "~icons/mdi/download";
  import MdiOpenInNew from "~icons/mdi/open-in-new";
  import { buttonVariants } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip";

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
