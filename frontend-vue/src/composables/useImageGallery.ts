import { ref } from "vue";
import imageService, { type GalleryImage } from "@/services/imageService";

export function useImageGallery() {
  const images = ref<GalleryImage[]>([]);
  const loading = ref(false);
  const uploading = ref(false);
  const errorMessage = ref("");
  const meta = ref({ total: 0, page: 1, page_size: 20, pages: 1 });

  const getBackendUrl = () => {
    if (import.meta.env.DEV) {
      const currentUrl = window.location.origin;
      const isLocalhost =
        window.location.hostname === "localhost" ||
        window.location.hostname === "127.0.0.1";
      return isLocalhost
        ? `${currentUrl.replace(/:\d+$/, "")}:3000`
        : currentUrl;
    }
    return import.meta.env.VITE_API_URL || "";
  };

  const getImageUrl = (img: GalleryImage) => {
    if (img.local_path.startsWith("http")) return img.local_path;
    const baseUrl = getBackendUrl();
    const path = img.local_path.startsWith("/")
      ? img.local_path
      : `/${img.local_path}`;
    if (path.startsWith("/uploads")) return `${baseUrl}${path}`;
    return `${baseUrl}/uploads/images/${img.tenant_id}/${img.local_path}`;
  };

  const fetchImages = async (page = 1, searchQuery = "") => {
    loading.value = true;
    errorMessage.value = "";
    try {
      const response = await imageService.getGallery(
        page,
        meta.value.page_size,
        searchQuery,
      );
      if (response.success) {
        images.value = response.data;
        meta.value = response.meta;
      }
    } catch (error) {
      console.error("Failed to fetch gallery:", error);
      errorMessage.value = "Gagal memuat galeri gambar";
    } finally {
      loading.value = false;
    }
  };

  const uploadFiles = async (files: FileList) => {
    uploading.value = true;
    errorMessage.value = "";
    let successCount = 0;
    try {
      const uploads = Array.from(files).map((file) =>
        imageService.uploadImage(file, "gallery"),
      );
      const results = await Promise.allSettled(uploads);
      results.forEach((res) => {
        if (res.status === "fulfilled" && res.value.success) successCount++;
      });
      return { successCount, total: files.length };
    } catch (error) {
      console.error("Upload error:", error);
      errorMessage.value = "Terjadi kesalahan saat upload";
      return { successCount: 0, total: files.length };
    } finally {
      uploading.value = false;
    }
  };

  return {
    images,
    loading,
    uploading,
    errorMessage,
    meta,
    getImageUrl,
    fetchImages,
    uploadFiles,
  };
}
