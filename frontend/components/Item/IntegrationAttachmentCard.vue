<template>
  <div class="flex w-full gap-3">
    <div class="shrink-0">
      <img v-if="thumbnailUrl" :src="thumbnailUrl" class="size-14 rounded object-cover shadow" alt="" />
      <div v-else class="flex size-14 items-center justify-center rounded border bg-muted">
        <MdiFileDocument class="size-7 text-blue-500" aria-hidden="true" />
      </div>
    </div>

    <div class="min-w-0 flex-1 space-y-1">
      <p class="truncate font-medium leading-tight">
        {{ card.title || fallbackTitle }}
      </p>

      <p
        v-if="fields.correspondent || fields.documentType"
        class="flex items-center gap-1.5 text-xs text-muted-foreground"
      >
        <MdiAccountOutline v-if="fields.correspondent" class="size-3.5 shrink-0" />
        <span v-if="fields.correspondent" class="truncate">
          {{ fields.correspondent.name }}
        </span>
        <span v-if="fields.correspondent && fields.documentType" class="text-muted-foreground/40"> &middot; </span>
        <MdiTagOutline v-if="fields.documentType" class="size-3.5 shrink-0" />
        <span v-if="fields.documentType">
          {{ fields.documentType.name }}
        </span>
      </p>

      <div v-if="tags.length" class="flex flex-wrap gap-1">
        <span
          v-for="tag in tags"
          :key="tag.id"
          class="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium"
          :style="{ backgroundColor: tag.color, color: tag.textColor }"
        >
          {{ tag.name }}
        </span>
      </div>

      <p class="flex items-center gap-2 text-xs text-muted-foreground">
        <span v-if="fields.createdDate">{{ fields.createdDate }}</span>
        <span v-if="fields.pageCount" class="text-muted-foreground/40">&middot;</span>
        <span v-if="fields.pageCount">
          {{ $t("components.item.attachments_list.page_count", { count: fields.pageCount }) }}
        </span>
      </p>
    </div>

    <div class="ml-2 flex shrink-0 items-start gap-1">
      <TooltipProvider v-if="card.state === 'error'" :delay-duration="0">
        <Tooltip>
          <TooltipTrigger>
            <MdiAlertCircleOutline
              class="size-4 text-amber-500"
              :aria-label="$t('components.item.attachments_list.integration_error')"
            />
          </TooltipTrigger>
          <TooltipContent>{{ card.error }}</TooltipContent>
        </Tooltip>
      </TooltipProvider>
      <TooltipProvider :delay-duration="0">
        <Tooltip>
          <TooltipTrigger as-child>
            <a
              :class="buttonVariants({ size: 'icon', variant: 'outline' })"
              :href="card.openUrl"
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

<script setup lang="ts">
  import { computed } from "vue";
  import type { IntegrationAttachmentCard } from "~/lib/api/classes/items";
  import MdiFileDocument from "~icons/mdi/file-document";
  import MdiAccountOutline from "~icons/mdi/account-outline";
  import MdiTagOutline from "~icons/mdi/tag-outline";
  import MdiAlertCircleOutline from "~icons/mdi/alert-circle-outline";
  import MdiOpenInNew from "~icons/mdi/open-in-new";
  import { buttonVariants } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

  const props = defineProps<{
    card: IntegrationAttachmentCard;
    fallbackTitle: string;
    thumbnailUrl: string;
  }>();

  const fields = computed(() => props.card.fields ?? {});
  const tags = computed(() => fields.value.tags ?? []);
</script>
