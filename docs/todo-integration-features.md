# Integration Feature Backlog

## TODO-1: Paperless Metadata Caching

**Problem:** `AttachmentsList.vue` fetches enriched metadata (title, correspondent, tags, thumbnail) on every
component mount. On slow networks or high-latency Paperless instances this causes visible flickering and
unnecessary repeated requests.

**Scope:**
- Enriched doc metadata (`/api/documents/{docId}/`)
- Correspondent/document-type/tag lookups
- Thumbnails (object URL or base64)

**Implementation ideas:**
1. **Pinia store** (`stores/integrations.ts`)
   - `paperlessCache: Map<docId, { data: PaperlessEnrichedDoc, fetchedAt: number }>`
   - TTL: 30 minutes (configurable via settings)
   - `clearCache(docId?)` action for manual invalidation (e.g. after reclassify)
2. **Thumbnail cache**
   - Store as base64 string (survives component remount)
   - Or keep in Pinia with `thumbBase64: Map<docId, string>`
   - Revoke old Object URLs before replacing
3. **Cache invalidation triggers:**
   - Reclassify button click
   - Settings change (paperless URL/token change → clear all)
   - Manual "Refresh" button per card (optional)
4. **Persistence (optional stretch goal):**
   - `localStorage` with JSON serialization for metadata (not thumbnails)
   - Key: `homebox:paperless:doc:{docId}`
   - Don't persist thumbnails (binary, storage limits)

**Files to change:**
- New: `frontend/stores/integrations.ts`
- `frontend/components/Item/AttachmentsList.vue` — replace `ref({})` maps with store reads/writes

---

## TODO-2: Offline / Error Indicator on Integration Cards

**Problem:** When Paperless is unreachable (network down, wrong URL, expired token), the card shows
nothing — no thumbnail, no title, no indication that something is wrong. User has no feedback.

**UI design:**
- Show an `⚠` badge (amber `MdiAlertCircleOutline`) in the top-right corner of the card
- Gray out the card slightly (`opacity-70`)
- Tooltip on badge: "Could not reach Paperless – check connection and token in Profile settings"
- Distinguished states:
  | State | Indicator |
  |---|---|
  | Loading | spinner (`MdiLoading animate-spin`) |
  | Loaded | normal card |
  | Fetch error (4xx/5xx) | ⚠ badge + grayed card |
  | No URL configured | info icon + "Configure Paperless in Profile" |

**Implementation:**
1. Add reactive map to component state:
   ```typescript
   const paperlessErrors = ref<Record<string, 'loading' | 'ok' | 'error' | 'unconfigured'>>({})
   ```
2. Set `paperlessErrors.value[docId] = 'loading'` before fetch, `'ok'` on success, `'error'` on failure
3. Set `'unconfigured'` if `paperlessUrlBase` is empty when `hydrateAttachments` runs
4. Template:
   ```vue
   <div class="relative">
     <!-- existing card content -->
     <MdiAlertCircleOutline
       v-if="paperlessErrors[resolveDocId(attachment)] === 'error'"
       class="absolute top-1 right-1 size-4 text-amber-500"
       title="Could not reach Paperless"
     />
   </div>
   ```
5. Same pattern for Immich (`immichErrors`)

**Files to change:**
- `frontend/components/Item/AttachmentsList.vue`

**Bonus (stretch):** Retry button on error card — re-triggers `fetchPaperlessEnriched(docId)` after
clearing the cache entry.

---

## TODO-3 (Related): Verify Connection Button in Profile Settings

Before saving Paperless/Immich credentials, allow user to test the connection:
- "Test Connection" button next to URL field
- Calls `GET /api/v1/integrations/paperless/proxy?path=/api/documents/?page_size=1`
- Shows green checkmark if 200, red X with error message otherwise
- Non-blocking (doesn't prevent saving)

**Files to change:**
- `frontend/pages/profile.vue`
