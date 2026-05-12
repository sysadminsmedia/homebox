<template>
  <ul role="list" class="divide-y rounded-md border">
    <li v-for="attachment in displayAttachments" :key="attachment.id" class="py-3 pl-3 pr-4 text-sm">
      <!-- ================================================================ -->
      <!-- Paperless document                                               -->
      <!-- ================================================================ -->
      <template v-if="attachment.mimeType === MIME_PAPERLESS">
        <!-- Unconfigured: URL removed — render as plain unlinkable attachment, no service hints -->
        <div v-if="!paperlessConfigured" class="flex items-center">
          <MdiPaperclip class="size-5 shrink-0 text-gray-400" aria-hidden="true" />
          <span class="ml-2 truncate">{{ attachment.title }}</span>
        </div>

        <!-- Configured: rich card with loading / ok / stale / error states -->
        <div v-else class="flex w-full gap-3">
          <!-- Thumbnail area -->
          <div class="shrink-0">
            <div
              v-if="attachState(attachment) === 'loading' && !paperlessThumbUrls[attachment.path]"
              class="flex size-14 items-center justify-center rounded border bg-muted"
            >
              <MdiLoading class="size-7 animate-spin text-muted-foreground" aria-hidden="true" />
            </div>
            <img
              v-else-if="paperlessThumbUrls[attachment.path]"
              :src="paperlessThumbUrls[attachment.path]"
              class="size-14 rounded object-cover shadow"
              alt=""
            />
            <div v-else class="flex size-14 items-center justify-center rounded border bg-muted">
              <MdiFileDocument class="size-7 text-blue-500" aria-hidden="true" />
            </div>
          </div>

          <!-- Content area -->
          <div class="min-w-0 flex-1 space-y-1">
            <p class="truncate font-medium leading-tight">
              {{ paperlessDoc(attachment)?.title || attachment.title }}
            </p>

            <p
              v-if="paperlessDoc(attachment)?.correspondent || paperlessDoc(attachment)?.document_type"
              class="flex items-center gap-1.5 text-xs text-muted-foreground"
            >
              <MdiAccountOutline v-if="paperlessDoc(attachment)?.correspondent" class="size-3.5 shrink-0" />
              <span v-if="paperlessDoc(attachment)?.correspondent" class="truncate">
                {{ paperlessDoc(attachment)?.correspondent?.name }}
              </span>
              <span
                v-if="paperlessDoc(attachment)?.correspondent && paperlessDoc(attachment)?.document_type"
                class="text-muted-foreground/40"
              >
                ·
              </span>
              <MdiTagOutline v-if="paperlessDoc(attachment)?.document_type" class="size-3.5 shrink-0" />
              <span v-if="paperlessDoc(attachment)?.document_type">
                {{ paperlessDoc(attachment)?.document_type?.name }}
              </span>
            </p>

            <div v-if="paperlessDoc(attachment)?.tags?.length" class="flex flex-wrap gap-1">
              <span
                v-for="tag in paperlessDoc(attachment)?.tags"
                :key="tag.id"
                class="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium"
                :style="{ backgroundColor: tag.color, color: tag.text_color }"
              >
                {{ tag.name }}
              </span>
            </div>

            <p class="flex items-center gap-2 text-xs text-muted-foreground">
              <span v-if="paperlessDoc(attachment)?.created_date">{{ paperlessDoc(attachment)?.created_date }}</span>
              <span v-if="paperlessDoc(attachment)?.page_count" class="text-muted-foreground/40">·</span>
              <span v-if="paperlessDoc(attachment)?.page_count"
                >{{ paperlessDoc(attachment)?.page_count }} pages</span
              >
            </p>
          </div>

          <!-- Actions area -->
          <div class="ml-2 flex shrink-0 items-start gap-1">
            <!-- ⚠ unreachable badge -->
            <TooltipProvider v-if="isUnreachable(attachment)" :delay-duration="0">
              <Tooltip>
                <TooltipTrigger>
                  <MdiAlertCircleOutline class="size-4 text-amber-500" aria-label="Service unreachable" />
                </TooltipTrigger>
                <TooltipContent>{{ attachError(attachment) }}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
            <!-- Open in Paperless -->
            <TooltipProvider :delay-duration="0">
              <Tooltip>
                <TooltipTrigger as-child>
                  <a
                    :class="buttonVariants({ size: 'icon', variant: 'outline' })"
                    :href="paperlessOpenUrl(attachment)"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <MdiOpenInNew />
                  </a>
                </TooltipTrigger>
                <TooltipContent>
                  {{ $t("components.item.attachments_list.open_in_service", { service: "Paperless" }) }}
                </TooltipContent>
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
  import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
  import type { ItemAttachment } from "~~/lib/api/types/data-contracts";
  import { useIntegrationCacheStore } from "~/stores/integration-cache";
  import type { AttachmentFetchState } from "~/stores/integration-cache";
  import { SERVICE_ADAPTERS, getAdapterByMimeType } from "~/lib/integration-adapters";
  import MdiPaperclip from "~icons/mdi/paperclip";
  import MdiLinkVariant from "~icons/mdi/link-variant";
  import MdiDownload from "~icons/mdi/download";
  import MdiOpenInNew from "~icons/mdi/open-in-new";
  import MdiFileDocument from "~icons/mdi/file-document";
  import MdiAccountOutline from "~icons/mdi/account-outline";
  import MdiTagOutline from "~icons/mdi/tag-outline";
  import MdiLoading from "~icons/mdi/loading";
  import MdiAlertCircleOutline from "~icons/mdi/alert-circle-outline";
  import { buttonVariants } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

  const MIME_LINK = "link/url";
  /** MIME type for Paperless-ngx document links — sourced from the adapter registry. */
  const MIME_PAPERLESS = getAdapterByMimeType("paperless/document")!.mimeType;

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
  const store = useIntegrationCacheStore();
  const { t } = useI18n();

  // ---------------------------------------------------------------------------
  // Settings state (reactive, driven by store)
  // ---------------------------------------------------------------------------

  const paperlessConfigured = computed(() => !!store.serviceUrls["paperless"]?.trim());

  // ---------------------------------------------------------------------------
  // Runtime lift: treat link/url attachments as their service type when the
  // URL matches.
  //  - Configured service + host matches → lifted (full card + hydration)
  //  - No configured service URL + pattern matches → lifted (demoted plain
  //    paperclip, no hydration because service is unconfigured)
  //  - Configured service + host doesn't match → NOT lifted (URL link, safe)
  // ---------------------------------------------------------------------------

  function liftAttachment(attachment: ItemAttachment): ItemAttachment {
    if (attachment.mimeType !== MIME_LINK) return attachment;
    for (const adapter of SERVICE_ADAPTERS) {
      const configuredUrl = store.serviceUrls[adapter.name]?.trim();
      if (!configuredUrl) continue; // service not configured → keep as plain link/url
      const id = adapter.extractId(attachment.path, configuredUrl);
      if (id !== null) return { ...attachment, mimeType: adapter.mimeType, path: id };
    }
    return attachment;
  }

  /**
   * Attachments as the template sees them.  link/url entries that match a
   * service URL are lifted to the correct mimeType/path so the existing
   * Paperless card branch handles them automatically.
   */
  const displayAttachments = computed(() => props.attachments.map(liftAttachment));

  // ---------------------------------------------------------------------------
  // Thumbnail state (component-local — objectURLs must be revoked on unmount)
  // ---------------------------------------------------------------------------

  const paperlessThumbUrls = ref<Record<string, string>>({});
  const objectUrls = new Set<string>();

  function createObjectUrl(blob: Blob): string {
    const url = URL.createObjectURL(blob);
    objectUrls.add(url);
    return url;
  }

  onBeforeUnmount(() => {
    for (const url of objectUrls) URL.revokeObjectURL(url);
    objectUrls.clear();
  });

  // ---------------------------------------------------------------------------
  // Enriched-data types
  // ---------------------------------------------------------------------------

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

  // ---------------------------------------------------------------------------
  // Store-backed typed getters
  // ---------------------------------------------------------------------------

  function paperlessDoc(attachment: ItemAttachment): PaperlessEnrichedDoc | null {
    return (store.getEnrichedData("paperless", attachment.path) as PaperlessEnrichedDoc | undefined) ?? null;
  }

  function attachState(attachment: ItemAttachment): AttachmentFetchState | undefined {
    return store.fetchStates[attachment.id];
  }

  function attachError(attachment: ItemAttachment): string | undefined {
    return store.fetchErrors[attachment.id];
  }

  /** True when state is stale or error (server was unreachable). */
  function isUnreachable(attachment: ItemAttachment): boolean {
    const s = attachState(attachment);
    return s === "stale" || s === "error";
  }

  // ---------------------------------------------------------------------------
  // URL helpers
  // ---------------------------------------------------------------------------

  function attachmentURL(attachmentId: string) {
    return api.authURL(`/entities/${props.itemId}/attachments/${attachmentId}`);
  }

  function proxyUrl(serviceName: string, relativePath: string): string {
    return `/api/v1/integrations/${serviceName}/proxy?path=${encodeURIComponent(relativePath)}`;
  }

  function paperlessOpenUrl(attachment: ItemAttachment): string {
    const base = (store.serviceUrls["paperless"] ?? "").replace(/\/$/, "");
    if (!base) return "#";
    return `${base}/documents/${attachment.path}/details`;
  }


  function isLink(attachment: ItemAttachment): boolean {
    return attachment.mimeType === MIME_LINK;
  }

  // ---------------------------------------------------------------------------
  // Error classification
  // ---------------------------------------------------------------------------

  /**
   * Turn a caught fetch error into a human-readable tooltip string.
   * Errors thrown by fetchX functions carry the HTTP status in their message
   * as "HTTP {status}" so we can show a precise reason.
   */
  function describeRequestError(err: unknown, baseUrl: string): string {
    const msg = err instanceof Error ? err.message : "";
    const statusMatch = msg.match(/^HTTP (\d+)$/);
    if (statusMatch) {
      const status = Number(statusMatch[1]);
      if (status === 401 || status === 403) {
        return t("components.item.attachments_list.errors.auth_failed");
      }
      return t("components.item.attachments_list.errors.request_failed", { status });
    }
    // No HTTP status → network-level failure (DNS, refused connection, etc.)
    return t("components.item.attachments_list.errors.service_unreachable", { baseUrl });
  }

  // ---------------------------------------------------------------------------
  // Paperless hydration
  // ---------------------------------------------------------------------------

  async function fetchPaperlessEnrichedDoc(docId: string): Promise<PaperlessEnrichedDoc> {
    const rawRes = await fetch(proxyUrl("paperless", `/api/documents/${docId}/`));
    if (!rawRes.ok) throw new Error(`HTTP ${rawRes.status}`);

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
    return doc;
  }

  async function fetchPaperlessThumbnail(docId: string): Promise<void> {
    if (paperlessThumbUrls.value[docId]) return;
    const res = await fetch(proxyUrl("paperless", `/api/documents/${docId}/thumb/`));
    if (!res.ok) return;
    paperlessThumbUrls.value[docId] = createObjectUrl(await res.blob());
  }

  async function hydratePaperless(attachment: ItemAttachment): Promise<void> {
    const docId = attachment.path;
    const configuredUrl = store.serviceUrls["paperless"] ?? "";
    const cached = store.getEnrichedData("paperless", docId);

    // Show cached data immediately (stale state) while we try to refresh.
    store.setState(attachment.id, cached ? "stale" : "loading");

    try {
      const doc = await fetchPaperlessEnrichedDoc(docId);
      store.setEnrichedData("paperless", docId, doc);
      store.setState(attachment.id, "ok");
      // Thumbnail is best-effort after enriched data succeeds.
      fetchPaperlessThumbnail(docId).catch(() => {});
    } catch (err) {
      store.setState(attachment.id, cached ? "stale" : "error", describeRequestError(err, configuredUrl));
    }
  }

  // ---------------------------------------------------------------------------
  // Generic dispatch — add a new service by adding a new else-if branch here
  // ---------------------------------------------------------------------------

  async function hydrateAllAttachments(attachments: ItemAttachment[]): Promise<void> {
    await Promise.all(
      attachments.map(async attachment => {
        const serviceName = mimeToServiceName(attachment.mimeType);
        if (!serviceName) return; // not a service attachment
        if (!store.serviceUrls[serviceName]?.trim()) return; // unconfigured — show degraded state
        if (serviceName === "paperless") {
          await hydratePaperless(attachment);
        }
      })
    );
  }

  /** Maps a service MIME type to its service name, or null for non-service types. */
  function mimeToServiceName(mimeType: string): string | null {
    return getAdapterByMimeType(mimeType)?.name ?? null;
  }

  // ---------------------------------------------------------------------------
  // Lifecycle
  // ---------------------------------------------------------------------------

  onMounted(async () => {
    await store.loadSettings(api);
    void hydrateAllAttachments(displayAttachments.value);
  });

  // React to new/removed attachments.
  watch(
    () => props.attachments,
    async newAttachments => {
      const lifted = newAttachments.map(liftAttachment);
      const fresh = lifted.filter(a => !(a.id in store.fetchStates));
      if (fresh.length > 0) {
        void hydrateAllAttachments(fresh);
      }
    }
  );

  /**
   * React to URL additions/removals in the store (e.g. user saves/clears the
   * service URL in Profile Settings while this component is mounted).
   * - URL added   → hydrate all attachments for that service (promote)
   * - URL removed → clear fetch state so they fall back to the unconfigured card
   */
  watch(
    () => ({ ...store.serviceUrls }),
    (newUrls, oldUrls) => {
      for (const [name, newUrl] of Object.entries(newUrls)) {
        const oldUrl = oldUrls?.[name] ?? "";
        if (newUrl === oldUrl) continue;

        const affected = displayAttachments.value.filter(a => mimeToServiceName(a.mimeType) === name);
        if (affected.length === 0) continue;

        if (!newUrl.trim()) {
          // URL removed → demote: clear state so template shows the unconfigured card.
          for (const a of affected) store.clearAttachmentState(a.id);
        } else {
          // URL added/changed → promote: re-hydrate all affected attachments.
          for (const a of affected) store.clearAttachmentState(a.id);
          void hydrateAllAttachments(affected);
        }
      }
    },
    { deep: false }
  );
</script>

<style scoped></style>
