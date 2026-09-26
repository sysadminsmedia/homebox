export interface PhotoPreview {
  photoName: string;
  file: File;
  fileBase64: string;
  primary: boolean;
}

export function dataURLtoFile(dataURL: string, fileName: string) {
  const arr = dataURL.split(",");
  const mimeMatch = arr[0]!.match(/:(.*?);/);
  if (!mimeMatch || !mimeMatch[1]) {
    throw new Error("Invalid data URL format");
  }

  const mime = mimeMatch[1];
  if (!mime.startsWith("image/")) {
    throw new Error("Invalid mime type, expected image");
  }

  const bstr = atob(arr[arr.length - 1]!);
  let n = bstr.length;
  const u8arr = new Uint8Array(n);
  while (n--) {
    u8arr[n] = bstr.charCodeAt(n);
  }

  return new File([u8arr], fileName, { type: mime });
}

export async function fileToPhotoPreview(file: File, primary = false): Promise<PhotoPreview> {
  const fileBase64 = await readFileAsDataUrl(file);

  return {
    photoName: file.name,
    file,
    fileBase64,
    primary,
  };
}

export async function filesToPhotoPreviews(files: FileList | File[], existingCount = 0): Promise<PhotoPreview[]> {
  const nextPhotos: PhotoPreview[] = [];

  for (const file of Array.from(files)) {
    nextPhotos.push(await fileToPhotoPreview(file, existingCount + nextPhotos.length === 0));
  }

  return nextPhotos;
}

export async function rotatePhotoPreview(photo: PhotoPreview): Promise<PhotoPreview> {
  const offScreenCanvas = document.createElement("canvas");
  const offScreenCanvasCtx = offScreenCanvas.getContext("2d");

  if (!offScreenCanvasCtx) {
    throw new Error("Canvas not supported");
  }

  const img = new Image();
  await new Promise<void>((resolve, reject) => {
    img.onload = () => resolve();
    img.onerror = () => reject(new Error("Failed to load image"));
    img.src = photo.fileBase64;
  });

  offScreenCanvas.height = img.width;
  offScreenCanvas.width = img.height;

  offScreenCanvasCtx.rotate((90 * Math.PI) / 180);
  offScreenCanvasCtx.translate(0, -offScreenCanvas.width);
  offScreenCanvasCtx.drawImage(img, 0, 0);

  const imageType = photo.fileBase64.match(/^data:(.+);base64/)?.[1] || "image/jpeg";
  const fileBase64 = offScreenCanvas.toDataURL(imageType, 1);

  offScreenCanvas.width = 0;
  offScreenCanvas.height = 0;

  return {
    ...photo,
    fileBase64,
    file: dataURLtoFile(fileBase64, photo.photoName),
  };
}

export function setPrimaryPhoto(photos: PhotoPreview[], index: number): PhotoPreview[] {
  const nextPhotos = photos.map(photo => ({ ...photo, primary: false }));
  if (nextPhotos[index]) {
    nextPhotos[index] = { ...nextPhotos[index]!, primary: true };
  }
  return nextPhotos;
}

export function deletePhoto(photos: PhotoPreview[], index: number): PhotoPreview[] {
  const nextPhotos = photos.filter((_, photoIndex) => photoIndex !== index);
  if (nextPhotos.length > 0 && !nextPhotos.some(photo => photo.primary)) {
    nextPhotos[0] = { ...nextPhotos[0]!, primary: true };
  }
  return nextPhotos;
}

async function readFileAsDataUrl(file: File): Promise<string> {
  return await new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = event => resolve(event.target?.result as string);
    reader.onerror = () => reject(new Error(`Failed to read file: ${file.name}`));
    reader.readAsDataURL(file);
  });
}
