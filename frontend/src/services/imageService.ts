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
    const formData = new FormData();
    formData.append("image", file);
    if (category) formData.append("category", category);

    // Note: apiService.post handles JSON by default, but axios handles FormData correctly
    // when passed as data. We might need to override headers if apiService forces JSON.
    // Looking at apiService, it sets 'Content-Type': 'application/json' in constructor.
    // Axios usually detects FormData and sets content-type to multipart/form-data with boundary.
    // However, if the interceptor or default headers force application/json, it might be an issue.
    // Let's try passing it directly, axios usually overrides it for FormData.

    return await apiService.post<UploadResponse>("/images/upload", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
  }
}

export default new ImageService();
