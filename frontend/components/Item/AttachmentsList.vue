<template>
  <ul role="list" class="divide-y rounded-md border">
    <li v-for="attachment in attachments" :key="attachment.id" :class="attachmentItemClass(attachment)">
      <template v-if="integrationCard(attachment)">
        <ItemIntegrationAttachmentCard
          :card="integrationCard(attachment)!"
          :fallback-title="attachment.title"
          :thumbnail-url="thumbnailURL(integrationCard(attachment)!)"
        />
      </template>

      <template v-else-if="isLink(attachment)">
        <div class="flex w-0 flex-1 items-center">
          <MdiLinkVariant class="size-5 shrink-0 text-foreground/50" aria-hidden="true" />
          <a
            v-if="safeLinkURL(attachment)"
            class="ml-2 w-0 flex-1 truncate text-primary underline"
            :href="safeLinkURL(attachment)"
            target="_blank"
            rel="noopener noreferrer"
          >
            {{ attachment.title }}
          </a>
          <span v-else class="ml-2 w-0 flex-1 truncate">{{ attachment.title }}</span>
        </div>
        <div v-if="safeLinkURL(attachment)" class="ml-4 flex shrink-0 gap-2">
          <TooltipProvider :delay-duration="0">
            <Tooltip>
              <TooltipTrigger as-child>
                <a
                  :class="buttonVariants({ size: 'icon' })"
                  :href="safeLinkURL(attachment)"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <MdiOpenInNew />
                </a>
              </TooltipTrigger>
              <TooltipContent> {{ $t("components.item.attachments_list.open_new_tab") }} </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </template>

      <template v-else>
        <div class="flex w-0 flex-1 items-center">
          <MdiPaperclip class="size-5 shrink-0 text-foreground/50" aria-hidden="true" />
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
              <TooltipContent> {{ $t("components.item.attachments_list.download") }} </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger as-child>
                <a :class="buttonVariants({ size: 'icon' })" :href="attachmentURL(attachment.id)" target="_blank">
                  <MdiOpenInNew />
                </a>
              </TooltipTrigger>
              <TooltipContent> {{ $t("components.item.attachments_list.open_new_tab") }} </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </template>
    </li>
  </ul>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref, watch } from "vue";
  import type { ItemAttachment } from "~~/lib/api/types/data-contracts";
  import type { IntegrationAttachmentCard } from "~/lib/api/classes/items";
  import ItemIntegrationAttachmentCard from "~/components/Item/IntegrationAttachmentCard.vue";
  import MdiPaperclip from "~icons/mdi/paperclip";
  import MdiLinkVariant from "~icons/mdi/link-variant";
  import MdiDownload from "~icons/mdi/download";
  import MdiOpenInNew from "~icons/mdi/open-in-new";
  import { buttonVariants } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

  const MIME_LINK = "link/url";
  const LEGACY_INTEGRATION_MIME = "paperless/document";

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
  const cards = ref<Record<string, IntegrationAttachmentCard>>({});
  let integrationCardsLoadSeq = 0;

  const attachmentSignature = computed(() =>
    props.attachments.map(a => `${a.id}:${a.mimeType}:${a.path}:${a.updatedAt}`).join("|")
  );

  function hasIntegrationCandidates(): boolean {
    return props.attachments.some(a => a.mimeType === MIME_LINK || a.mimeType === LEGACY_INTEGRATION_MIME);
  }

  async function loadIntegrationCards(): Promise<void> {
    const seq = ++integrationCardsLoadSeq;
    if (!hasIntegrationCandidates()) {
      cards.value = {};
      return;
    }

    const { data, error } = await api.items.attachments.integrationCards(props.itemId);
    if (error || seq !== integrationCardsLoadSeq) return;
    cards.value = Object.fromEntries(data.items.map(card => [card.attachmentId, card]));
  }

  function integrationCard(attachment: ItemAttachment): IntegrationAttachmentCard | undefined {
    return cards.value[attachment.id];
  }

  function attachmentItemClass(attachment: ItemAttachment): string {
    if (integrationCard(attachment)) {
      return "py-3 pl-3 pr-4 text-sm";
    }
    return "flex items-center justify-between py-3 pl-3 pr-4 text-sm";
  }

  function attachmentURL(attachmentId: string) {
    return api.authURL(`/entities/${props.itemId}/attachments/${attachmentId}`);
  }

  function thumbnailURL(card: IntegrationAttachmentCard): string {
    return card.thumbnailUrl ? api.authURL(card.thumbnailUrl) : "";
  }

  function isLink(attachment: ItemAttachment): boolean {
    return attachment.mimeType === MIME_LINK;
  }

  function safeLinkURL(attachment: ItemAttachment): string | undefined {
    try {
      const url = new URL(attachment.path);
      if ((url.protocol === "http:" || url.protocol === "https:") && !url.username && !url.password) {
        return url.toString();
      }
    } catch {
      // Render invalid link attachments as inert text.
    }
    return undefined;
  }

  onMounted(() => {
    void loadIntegrationCards();
  });

  watch(attachmentSignature, () => {
    void loadIntegrationCards();
  });
</script>
