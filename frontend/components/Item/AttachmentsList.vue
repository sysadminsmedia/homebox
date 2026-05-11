<template>
  <ul role="list" class="divide-y rounded-md border">
    <li v-for="attachment in attachments" :key="attachment.id" class="py-3 pl-3 pr-4 text-sm">
      <!-- Paperless document card (native or reclassifiable link) -->
      <template v-if="isTreatAsPaperless(attachment)">
        <div class="flex w-full gap-3">
          <div class="shrink-0">
            <img
              v-if="paperlessThumbUrl(resolveDocId(attachment))"
              :src="paperlessThumbUrl(resolveDocId(attachment))"
              class="size-14 rounded object-cover shadow"
              alt=""
            />
            <div v-else class="flex size-14 items-center justify-center rounded border bg-muted">
              <MdiFileDocument class="size-7 text-blue-500" aria-hidden="true" />
            </div>
          </div>

          <div class="min-w-0 flex-1 space-y-1">
            <p class="truncate font-medium leading-tight">
              {{ enriched(resolveDocId(attachment))?.title || attachment.title }}
            </p>

            <p
              v-if="
                enriched(resolveDocId(attachment))?.correspondent || enriched(resolveDocId(attachment))?.document_type
              "
              class="flex items-center gap-1.5 text-xs text-muted-foreground"
            >
              <MdiAccountOutline v-if="enriched(resolveDocId(attachment))?.correspondent" class="size-3.5 shrink-0" />
              <span v-if="enriched(resolveDocId(attachment))?.correspondent" class="truncate">
                {{ enriched(resolveDocId(attachment))?.correspondent?.name }}
              </span>
              <span
                v-if="
                  enriched(resolveDocId(attachment))?.correspondent && enriched(resolveDocId(attachment))?.document_type
                "
                class="text-muted-foreground/40"
              >
                ·
              </span>
              <MdiTagOutline v-if="enriched(resolveDocId(attachment))?.document_type" class="size-3.5 shrink-0" />
              <span v-if="enriched(resolveDocId(attachment))?.document_type">
                {{ enriched(resolveDocId(attachment))?.document_type?.name }}
              </span>
            </p>

            <div v-if="enriched(resolveDocId(attachment))?.tags?.length" class="flex flex-wrap gap-1">
              <span
                v-for="tag in enriched(resolveDocId(attachment))?.tags"
                :key="tag.id"
                class="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium"
                :style="{ backgroundColor: tag.color, color: tag.text_color }"
              >
                {{ tag.name }}
              </span>
            </div>

            <p class="flex items-center gap-2 text-xs text-muted-foreground">
              <span v-if="enriched(resolveDocId(attachment))?.created_date">{{
                enriched(resolveDocId(attachment))?.created_date
              }}</span>
              <span v-if="enriched(resolveDocId(attachment))?.page_count" class="text-muted-foreground/40">·</span>
              <span v-if="enriched(resolveDocId(attachment))?.page_count"
                >{{ enriched(resolveDocId(attachment))?.page_count }} pages</span
              >
            </p>
          </div>

          <div class="ml-2 flex shrink-0 items-start">
            <TooltipProvider :delay-duration="0">
              <Tooltip>
                <TooltipTrigger as-child>
                  <a
                    v-if="attachment.mimeType === MIME_LINK || paperlessUrlBase"
                    :class="buttonVariants({ size: 'icon', variant: 'outline' })"
                    :href="paperlessOpenUrl(attachment)"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <MdiOpenInNew />
                  </a>
                  <NuxtLink
                    v-else
                    to="/profile"
                    :class="buttonVariants({ size: 'icon', variant: 'ghost' })"
                    title="Configure Paperless URL in profile settings"
                  >
                    <MdiCogOutline class="text-muted-foreground" />
                  </NuxtLink>
                </TooltipTrigger>
                <TooltipContent>
                  <span v-if="attachment.mimeType === MIME_LINK || paperlessUrlBase">Open in Paperless</span>
                  <span v-else>Configure Paperless URL in profile settings</span>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </div>
      </template>

      <!-- Immich Asset Attachment -->
      <template v-else-if="isImmichAsset(attachment)">
        <div v-if="isPhotoAttachment(attachment)" class="flex w-full gap-3">
          <div class="shrink-0">
            <img
              v-if="immichThumbUrl(attachment.path)"
              :src="immichThumbUrl(attachment.path)"
              class="size-14 rounded object-cover shadow"
              alt=""
            />
            <div v-else class="flex size-14 items-center justify-center rounded border bg-muted">
              <MdiImage class="size-7 text-emerald-600" aria-hidden="true" />
            </div>
          </div>
          <div class="min-w-0 flex-1 space-y-1">
            <p class="truncate font-medium leading-tight">{{ attachment.title }}</p>
            <p class="text-xs text-muted-foreground">Immich</p>
            <p v-if="immichAssetData[attachment.path]?.originalFileName" class="truncate text-xs text-muted-foreground">
              {{ immichAssetData[attachment.path]?.originalFileName }}
            </p>
            <p
              v-if="immichAssetData[attachment.path]?.exifInfo?.dateTimeOriginal"
              class="text-xs text-muted-foreground"
            >
              {{ immichAssetData[attachment.path]?.exifInfo?.dateTimeOriginal }}
            </p>
          </div>
          <div class="ml-2 flex shrink-0 items-start">
            <TooltipProvider :delay-duration="0">
              <Tooltip>
                <TooltipTrigger as-child>
                  <a
                    :class="buttonVariants({ size: 'icon', variant: 'outline' })"
                    :href="immichAssetUrl(attachment.path)"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <MdiOpenInNew />
                  </a>
                </TooltipTrigger>
                <TooltipContent>Open in Immich</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </div>
        <div v-else class="flex items-center justify-between">
          <div class="flex w-0 flex-1 items-center">
            <MdiImage class="size-5 shrink-0 text-emerald-600" aria-hidden="true" />
            <div class="ml-2 w-0 flex-1">
              <p class="truncate font-semibold">{{ attachment.title }}</p>
              <p class="text-xs text-muted-foreground">Immich</p>
            </div>
          </div>
          <div class="ml-4 flex shrink-0 gap-2">
            <TooltipProvider :delay-duration="0">
              <Tooltip>
                <TooltipTrigger as-child>
                  <a
                    :class="buttonVariants({ size: 'icon' })"
                    :href="immichAssetUrl(attachment.path)"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <MdiOpenInNew />
                  </a>
                </TooltipTrigger>
                <TooltipContent>Open in Immich</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </div>
      </template>

      <!-- Generic Link Attachment -->
      <template v-else-if="isLink(attachment)">
        <div class="flex items-center justify-between">
          <div class="flex w-0 flex-1 items-center">
            <MdiLinkVariant class="size-5 shrink-0 text-gray-400" aria-hidden="true" />
            <a
              class="ml-2 w-0 flex-1 truncate text-primary underline"
              :href="attachment.path"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ attachment.title }}
            </a>
          </div>
          <div class="ml-4 flex shrink-0 gap-2">
            <TooltipProvider :delay-duration="0">
              <Tooltip>
                <TooltipTrigger as-child>
                  <a
                    :class="buttonVariants({ size: 'icon' })"
                    :href="attachment.path"
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

      <!-- File Attachment -->
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
  import { onBeforeUnmount, onMounted, ref, watch } from "vue";
  import type { ItemAttachment } from "~~/lib/api/types/data-contracts";
  import { getAdapterByMimeType, extractPaperlessDocId, type ServiceAdapter } from "~/lib/integration-adapters";
  import MdiPaperclip from "~icons/mdi/paperclip";
  import MdiLinkVariant from "~icons/mdi/link-variant";
  import MdiDownload from "~icons/mdi/download";
  import MdiOpenInNew from "~icons/mdi/open-in-new";
  import MdiCogOutline from "~icons/mdi/cog-outline";
  import MdiFileDocument from "~icons/mdi/file-document";
  import MdiAccountOutline from "~icons/mdi/account-outline";
  import MdiTagOutline from "~icons/mdi/tag-outline";
  import MdiImage from "~icons/mdi/image";
  import { buttonVariants } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

  const MIME_PAPERLESS = "paperless/document";
  const MIME_IMMICH = "immich/asset";
  const MIME_LINK = "link/url";

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
    return api.authURL(`/entities/${props.itemId}/attachments/${attachmentId}`);
  }

  interface PaperlessTag {
    id: number;
    name: string;
    color: string;
    text_color: string;
  }

  interface PaperlessRelated {
    id: number;
    name: string;
  }

  interface PaperlessEnrichedDoc {
    id: number;
    title: string;
    created_date: string;
    page_count: number;
    correspondent: PaperlessRelated | null;
    document_type: PaperlessRelated | null;
    tags: PaperlessTag[];
  }

  interface ImmichExif {
    dateTimeOriginal?: string;
  }

  interface ImmichAsset {
    originalFileName?: string;
    exifInfo?: ImmichExif;
  }

  const paperlessUrlBase = ref("");
  const paperlessThumbUrls = ref<Record<string, string>>({});
  const paperlessEnrichedDocs = ref<Record<string, PaperlessEnrichedDoc>>({});

  const immichUrlBase = ref("");
  const immichThumbUrls = ref<Record<string, string>>({});
  const immichAssetData = ref<Record<string, ImmichAsset>>({});

  const objectUrls = new Set<string>();

  function createObjectUrl(blob: Blob): string {
    const url = URL.createObjectURL(blob);
    objectUrls.add(url);
    return url;
  }

  onBeforeUnmount(() => {
    for (const objectUrl of objectUrls) {
      URL.revokeObjectURL(objectUrl);
    }
    objectUrls.clear();
  });

  /**
   * Returns the adapter for an attachment – including link attachments that can
   * be recognised as a known service via URL pattern matching.
   */
  function resolveAdapter(attachment: ItemAttachment): ServiceAdapter | undefined {
    const direct = getAdapterByMimeType(attachment.mimeType);
    if (direct) return direct;
    if (attachment.mimeType === MIME_LINK) {
      // Dynamically import only the needed helpers via the already-imported registry helper.
      if (extractPaperlessDocId(attachment.path)) return getAdapterByMimeType(MIME_PAPERLESS);
    }
    return undefined;
  }

  /** Returns the provider ID for both native and link-reclassifiable attachments. */
  function resolveDocId(attachment: ItemAttachment): string {
    if (attachment.mimeType === MIME_PAPERLESS) return attachment.path;
    return extractPaperlessDocId(attachment.path) ?? attachment.path;
  }

  /** True for native paperless or link attachments whose path is a paperless URL. */
  function isTreatAsPaperless(attachment: ItemAttachment): boolean {
    return resolveAdapter(attachment)?.name === "paperless";
  }

  function isImmichAsset(attachment: ItemAttachment): boolean {
    return attachment.mimeType === MIME_IMMICH;
  }

  function isLink(attachment: ItemAttachment): boolean {
    return attachment.mimeType === MIME_LINK;
  }

  function isPhotoAttachment(attachment: ItemAttachment): boolean {
    return attachment.type === "photo";
  }

  function enriched(docId: string): PaperlessEnrichedDoc | undefined {
    return paperlessEnrichedDocs.value[docId];
  }

  function paperlessThumbUrl(docId: string): string {
    return paperlessThumbUrls.value[docId] || "";
  }

  function immichThumbUrl(assetId: string): string {
    return immichThumbUrls.value[assetId] || "";
  }

  function paperlessDocUrl(docId: string): string {
    if (paperlessUrlBase.value) {
      return `${paperlessUrlBase.value}/documents/${docId}/details`;
    }
    return "#";
  }

  /**
   * URL to open the Paperless document in the external system.
   * For link/url attachments, the full URL is stored directly in attachment.path.
   * For native paperless/document attachments, we construct the URL from the configured base.
   */
  function paperlessOpenUrl(attachment: ItemAttachment): string {
    if (attachment.mimeType === MIME_LINK) return attachment.path;
    return paperlessDocUrl(resolveDocId(attachment));
  }

  function immichAssetUrl(assetId: string): string {
    if (immichUrlBase.value) {
      return `${immichUrlBase.value}/assets/${assetId}`;
    }
    return "#";
  }

  function proxyUrl(name: string, relativePath: string): string {
    return `/api/v1/integrations/${name}/proxy?path=${encodeURIComponent(relativePath)}`;
  }

  async function fetchPaperlessEnriched(docId: string) {
    if (paperlessEnrichedDocs.value[docId] !== undefined) {
      return;
    }

    const rawRes = await fetch(proxyUrl("paperless", `/api/documents/${docId}/`));
    if (!rawRes.ok) return;

    const raw = (await rawRes.json()) as {
      id: number;
      title: string;
      created_date: string;
      page_count: number;
      correspondent?: number;
      document_type?: number;
      tags?: number[];
    };

    const doc: PaperlessEnrichedDoc = {
      id: raw.id,
      title: raw.title,
      created_date: raw.created_date,
      page_count: raw.page_count,
      correspondent: null,
      document_type: null,
      tags: [],
    };

    const jobs: Promise<void>[] = [];

    if (raw.correspondent) {
      jobs.push(
        fetch(proxyUrl("paperless", `/api/correspondents/${raw.correspondent}/`))
          .then(r => (r.ok ? r.json() : null))
          .then(c => {
            if (c) {
              doc.correspondent = {
                id: Number((c as { id?: number }).id),
                name: String((c as { name?: string }).name || ""),
              };
            }
          })
          .catch(() => {})
      );
    }

    if (raw.document_type) {
      jobs.push(
        fetch(proxyUrl("paperless", `/api/document_types/${raw.document_type}/`))
          .then(r => (r.ok ? r.json() : null))
          .then(d => {
            if (d) {
              doc.document_type = {
                id: Number((d as { id?: number }).id),
                name: String((d as { name?: string }).name || ""),
              };
            }
          })
          .catch(() => {})
      );
    }

    const tagIds = Array.isArray(raw.tags) ? raw.tags : [];
    const tagResults: Array<PaperlessTag | null> = new Array(tagIds.length).fill(null);
    tagIds.forEach((tagId, idx) => {
      jobs.push(
        fetch(proxyUrl("paperless", `/api/tags/${tagId}/`))
          .then(r => (r.ok ? r.json() : null))
          .then(t => {
            if (t) {
              tagResults[idx] = {
                id: Number((t as { id?: number }).id),
                name: String((t as { name?: string }).name || ""),
                color: String((t as { color?: string }).color || "#E5E7EB"),
                text_color: String((t as { text_color?: string }).text_color || "#111827"),
              };
            }
          })
          .catch(() => {})
      );
    });

    await Promise.all(jobs);
    doc.tags = tagResults.filter((tag): tag is PaperlessTag => tag !== null);
    paperlessEnrichedDocs.value[docId] = doc;
  }

  async function fetchPaperlessThumbnail(docId: string) {
    if (paperlessThumbUrls.value[docId] !== undefined) {
      return;
    }

    const thumbRes = await fetch(proxyUrl("paperless", `/api/documents/${docId}/thumb/`));
    if (!thumbRes.ok) return;
    paperlessThumbUrls.value[docId] = createObjectUrl(await thumbRes.blob());
  }

  async function fetchImmichAsset(assetId: string) {
    if (immichAssetData.value[assetId] !== undefined) {
      return;
    }

    const assetRes = await fetch(proxyUrl("immich", `/api/assets/${assetId}`));
    if (assetRes.ok) {
      const asset = (await assetRes.json()) as ImmichAsset;
      immichAssetData.value[assetId] = asset;
    }
  }

  async function fetchImmichThumbnail(assetId: string) {
    if (immichThumbUrls.value[assetId] !== undefined) {
      return;
    }

    const thumbRes = await fetch(proxyUrl("immich", `/api/assets/${assetId}/thumbnail`));
    if (!thumbRes.ok) return;
    immichThumbUrls.value[assetId] = createObjectUrl(await thumbRes.blob());
  }

  async function loadIntegrationSettings() {
    const { data, error } = await api.user.getSettings();
    if (error || !data?.item) {
      return;
    }

    const settings = data.item as Record<string, unknown>;
    paperlessUrlBase.value = ((settings.paperless_url as string) || "").replace(/\/$/, "");
    immichUrlBase.value = ((settings.immich_url as string) || "").replace(/\/$/, "");
  }

  async function hydrateAttachments(attachments: ItemAttachment[]) {
    await Promise.all(
      attachments.map(async attachment => {
        const adapter = resolveAdapter(attachment);
        if (!adapter) return;

        if (adapter.name === "paperless") {
          const id = resolveDocId(attachment);
          try {
            await fetchPaperlessEnriched(id);
          } catch {
            /* degrade gracefully */
          }
          try {
            await fetchPaperlessThumbnail(id);
          } catch {
            /* degrade gracefully */
          }
        } else if (adapter.name === "immich") {
          try {
            await fetchImmichAsset(attachment.path);
          } catch {
            /* degrade gracefully */
          }
          if (isPhotoAttachment(attachment)) {
            try {
              await fetchImmichThumbnail(attachment.path);
            } catch {
              /* degrade gracefully */
            }
          }
        }
        // New services: add an else-if block here, or extend the adapter interface
        // with optional hydrateEnrichment/hydrateThumbnail callbacks.
      })
    );
  }

  onMounted(async () => {
    await loadIntegrationSettings();
    await hydrateAttachments(props.attachments);
  });

  watch(
    () => props.attachments,
    async newAttachments => {
      await hydrateAttachments(newAttachments);
    }
  );
</script>

<style scoped></style>
