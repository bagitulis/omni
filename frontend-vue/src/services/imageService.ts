import apiService from "./api";

export interface GalleryImage {
  id: number;
  tenant_id: string;
  filename: string;
  local_path: string;
  width: number;
  height: number;
  size: number;
  mime_type: string;
  category: string;
  created_at: string;
}

export interface GalleryResponse {
  success: boolean;
  data: GalleryImage[];
  meta: {
    total: number;
    page: number;
    page_size: number;
    pages: number;
  };
}

export interface UploadResponse {
  success: boolean;
  data: GalleryImage;
}

/**
 * Compress image before upload
 * - Resizes if dimensions exceed maxDimension
 * - Compresses to target file size
 * - Returns compressed file
 */
async function compressImage(
  file: File,
  maxSizeBytes: number = 2 * 1024 * 1024, // 2MB
  maxDimension: number = 2000,
): Promise<File> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => {
      // Calculate new dimensions
      let { width, height } = { width: img.width, height: img.height };
      if (width > maxDimension || height > maxDimension) {
        const ratio = Math.min(maxDimension / width, maxDimension / height);
        width = Math.round(width * ratio);
        height = Math.round(height * ratio);
      }

      // Draw to canvas
      const canvas = document.createElement("canvas");
      canvas.width = width;
      canvas.height = height;
      const ctx = canvas.getContext("2d");
      if (!ctx) {
        reject(new Error("Failed to get canvas context"));
        return;
      }
      ctx.drawImage(img, 0, 0, width, height);

      // Convert to blob with compression
      let quality = 0.9;
      const tryCompress = () => {
        canvas.toBlob(
          (blob) => {
            if (!blob) {
              reject(new Error("Compression failed"));
              return;
            }

            if (blob.size > maxSizeBytes && quality > 0.1) {
              quality -= 0.1;
              tryCompress();
            } else {
              // Create new file with original name
              const compressedFile = new File([blob], file.name, {
                type: "image/jpeg",
                lastModified: Date.now(),
              });
              resolve(compressedFile);
            }
          },
          "image/jpeg",
          quality,
        );
      };
      tryCompress();
    };
    img.onerror = () => reject(new Error("Failed to load image"));
    img.src = URL.createObjectURL(file);
  });
}

class ImageService {
  /**
   * Get images from gallery with pagination and filtering
   */
  async getGallery(
    page: number = 1,
    limit: number = 20,
    search: string = "",
    category: string = "",
  ): Promise<GalleryResponse> {
    const params = new URLSearchParams();
    params.append("page", page.toString());
    params.append("limit", limit.toString());
    if (search) params.append("search", search);
    if (category) params.append("category", category);

    return await apiService.get<GalleryResponse>(
      `/images/gallery?${params.toString()}`,
    );
  }

  /**
   * Upload a new image to the gallery
   */
  async uploadImage(
    file: File,
    category: string = "",
  ): Promise<UploadResponse> {
    // Compress image before upload
    const compressedFile = await compressImage(file);

    const formData = new FormData();
    formData.append("image", compressedFile);
    if (category) formData.append("category", category);

    return await apiService.post<UploadResponse>("/images/upload", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
  }
}

export { compressImage };
export default new ImageService();
