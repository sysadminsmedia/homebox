<template>
  <ul role="list" class="divide-y rounded-md border">
    <li v-for="attachment in attachments" :key="attachment.id" class="py-3 pl-3 pr-4 text-sm">
      <template v-if="integrationCard(attachment)">
        <div class="flex w-full gap-3">
          <div class="shrink-0">
            <img
              v-if="integrationCard(attachment)?.thumbnailUrl"
              :src="thumbnailURL(integrationCard(attachment)!)"
              class="size-14 rounded object-cover shadow"
              alt=""
            />
            <div v-else class="flex size-14 items-center justify-center rounded border bg-muted">
              <MdiFileDocument class="size-7 text-blue-500" aria-hidden="true" />
            </div>
          </div>

          <div class="min-w-0 flex-1 space-y-1">
            <p class="truncate font-medium leading-tight">
              {{ integrationCard(attachment)?.title || attachment.title }}
            </p>

            <p
              v-if="cardFields(attachment).correspondent || cardFields(attachment).documentType"
              class="flex items-center gap-1.5 text-xs text-muted-foreground"
            >
              <MdiAccountOutline v-if="cardFields(attachment).correspondent" class="size-3.5 shrink-0" />
              <span v-if="cardFields(attachment).correspondent" class="truncate">
                {{ cardFields(attachment).correspondent?.name }}
              </span>
              <span
                v-if="cardFields(attachment).correspondent && cardFields(attachment).documentType"
                class="text-muted-foreground/40"
              >
                ·
              </span>
              <MdiTagOutline v-if="cardFields(attachment).documentType" class="size-3.5 shrink-0" />
              <span v-if="cardFields(attachment).documentType">
                {{ cardFields(attachment).documentType?.name }}
              </span>
            </p>

            <div v-if="cardTags(attachment).length" class="flex flex-wrap gap-1">
              <span
                v-for="tag in cardTags(attachment)"
                :key="tag.id"
                class="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium"
                :style="{ backgroundColor: tag.color, color: tag.textColor }"
              >
                {{ tag.name }}
              </span>
            </div>

            <p class="flex items-center gap-2 text-xs text-muted-foreground">
              <span v-if="cardFields(attachment).createdDate">{{ cardFields(attachment).createdDate }}</span>
              <span v-if="cardFields(attachment).pageCount" class="text-muted-foreground/40">·</span>
              <span v-if="cardFields(attachment).pageCount">
                {{ $t("components.item.attachments_list.page_count", { count: cardFields(attachment).pageCount }) }}
              </span>
            </p>
          </div>

          <div class="ml-2 flex shrink-0 items-start gap-1">
            <TooltipProvider v-if="integrationCard(attachment)?.state === 'error'" :delay-duration="0">
              <Tooltip>
                <TooltipTrigger>
                  <MdiAlertCircleOutline
                    class="size-4 text-amber-500"
                    :aria-label="$t('components.item.attachments_list.integration_error')"
                  />
                </TooltipTrigger>
                <TooltipContent>{{ integrationCard(attachment)?.error }}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
            <TooltipProvider :delay-duration="0">
              <Tooltip>
                <TooltipTrigger as-child>
                  <a
                    :class="buttonVariants({ size: 'icon', variant: 'outline' })"
                    :href="integrationCard(attachment)?.openUrl"
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
        </div>
      </template>

      <template v-else-if="isLink(attachment)">
        <div class="flex items-center justify-between">
          <div class="flex w-0 flex-1 items-center">
            <MdiLinkVariant class="size-5 shrink-0 text-gray-400" aria-hidden="true" />
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
        </div>
      </template>

      <template v-else>
        <div class="flex items-center justify-between">
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
        </div>
      </template>
    </li>
  </ul>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref, watch } from "vue";
  import type { ItemAttachment } from "~~/lib/api/types/data-contracts";
  import type { IntegrationAttachmentCard } from "~/lib/api/classes/items";
  import MdiPaperclip from "~icons/mdi/paperclip";
  import MdiLinkVariant from "~icons/mdi/link-variant";
  import MdiDownload from "~icons/mdi/download";
  import MdiOpenInNew from "~icons/mdi/open-in-new";
  import MdiFileDocument from "~icons/mdi/file-document";
  import MdiAccountOutline from "~icons/mdi/account-outline";
  import MdiTagOutline from "~icons/mdi/tag-outline";
  import MdiAlertCircleOutline from "~icons/mdi/alert-circle-outline";
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

  function cardFields(attachment: ItemAttachment): NonNullable<IntegrationAttachmentCard["fields"]> {
    return integrationCard(attachment)?.fields ?? {};
  }

  function cardTags(attachment: ItemAttachment) {
    return cardFields(attachment).tags ?? [];
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
