import { computed, ref } from "vue";
import { withRetries, type RetryOptions, type RetryResult } from "~~/lib/requests/retry";
import type { EntityOut } from "~~/lib/api/types/data-contracts";
import type { AttachmentTypes } from "~~/lib/api/types/non-generated";

export type UploadResult = RetryResult<EntityOut>;

/**
 * Wraps the attachments API with retry-on-transient-failure (see `lib/requests/retry`)
 * and exposes a reactive `uploading` flag for UI state. Errors are returned as a
 * tagged-union result with a human-readable reason — call sites just check `result.ok`.
 */
export function useAttachmentUpload(options: RetryOptions = {}) {
  const api = useUserApi();

  // Counter (not boolean) so concurrent uploads each keep the flag true while in flight.
  const inFlight = ref(0);
  const uploading = computed(() => inFlight.value > 0);

  async function track<T>(call: () => Promise<T>): Promise<T> {
    inFlight.value++;
    try {
      return await call();
    } finally {
      inFlight.value--;
    }
  }

  function uploadFile(
    itemId: string,
    file: File | Blob,
    filename: string,
    type: AttachmentTypes | null = null,
    primary?: boolean
  ): Promise<UploadResult> {
    return track(() => withRetries(() => api.items.attachments.add(itemId, file, filename, type, primary), options));
  }

  function uploadExternalLink(
    itemId: string,
    sourceType: string,
    externalId: string,
    title: string,
    attachmentType?: string
  ): Promise<UploadResult> {
    return track(() =>
      withRetries(
        () => api.items.attachments.addExternalLink(itemId, sourceType, externalId, title, attachmentType),
        options
      )
    );
  }

  return { uploading, uploadFile, uploadExternalLink };
}
