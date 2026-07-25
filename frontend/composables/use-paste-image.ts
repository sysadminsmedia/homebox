const MIME_EXTENSIONS: Record<string, string> = {
  "image/png": "png",
  "image/jpeg": "jpg",
  "image/gif": "gif",
  "image/webp": "webp",
  "image/avif": "avif",
};

function timestamp(date = new Date()): string {
  const pad = (value: number) => String(value).padStart(2, "0");

  return (
    `${date.getFullYear()}${pad(date.getMonth() + 1)}${pad(date.getDate())}` +
    `-${pad(date.getHours())}${pad(date.getMinutes())}${pad(date.getSeconds())}`
  );
}

/**
 * Clipboard images arrive with a generic name ("image.png") or no name at all, so give
 * every pasted image a unique, extension-correct one before it becomes an attachment.
 */
export function pastedImageName(mime: string, index = 0): string {
  const extension = MIME_EXTENSIONS[mime] || mime.replace("image/", "").split("+")[0] || "png";
  const suffix = index === 0 ? "" : `-${index}`;

  return `pasted-image-${timestamp()}${suffix}.${extension}`;
}

function isImage(file: File | null): file is File {
  return !!file && file.type.startsWith("image/");
}

function rename(file: File, index: number): File {
  return new File([file], pastedImageName(file.type, index), { type: file.type });
}

/**
 * Extracts the images out of a paste (or drop) payload, ignoring any text or non-image
 * files that came along with them.
 */
export function imageFilesFromClipboard(data: DataTransfer | null | undefined): File[] {
  if (!data) {
    return [];
  }

  const files = data.files?.length
    ? Array.from(data.files)
    : Array.from(data.items || [])
        .filter(item => item.kind === "file")
        .map(item => item.getAsFile());

  return files.filter(isImage).map(rename);
}

export function canReadClipboardImages(): boolean {
  return typeof navigator !== "undefined" && typeof navigator.clipboard?.read === "function";
}

/**
 * Pulls images out of the clipboard on demand, for the "paste image" button. Returns an
 * empty list when the clipboard holds no image or the user denied permission.
 */
export async function readClipboardImages(): Promise<File[]> {
  if (!canReadClipboardImages()) {
    return [];
  }

  const images: File[] = [];

  for (const item of await navigator.clipboard.read()) {
    const mime = item.types.find(type => type.startsWith("image/"));
    if (!mime) {
      continue;
    }

    const blob = await item.getType(mime);
    images.push(new File([blob], pastedImageName(mime, images.length), { type: mime }));
  }

  return images;
}

/**
 * Handles Ctrl/Cmd+V anywhere on the page while the caller is mounted. Pastes that carry
 * no image are left alone, so pasting text into a form field still works normally.
 */
export function usePasteImage(onImages: (files: File[]) => void) {
  useEventListener(document, "paste", (event: ClipboardEvent) => {
    const images = imageFilesFromClipboard(event.clipboardData);
    if (images.length === 0) {
      return;
    }

    event.preventDefault();
    onImages(images);
  });
}
