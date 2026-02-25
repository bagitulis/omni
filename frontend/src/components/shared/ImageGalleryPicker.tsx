import {
  type ChangeEvent,
  type FC,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import { Modal, message, theme } from "antd";
import type { UploadRequestOption } from "rc-upload/lib/interface";
import apiClient from "@/api/client";
import type { GalleryImage, GalleryPaginationMeta } from "../../types/shared";
import { GalleryFooter } from "./gallery/GalleryFooter";
import { GalleryGrid } from "./gallery/GalleryGrid";
import { GalleryToolbar } from "./gallery/GalleryToolbar";
import { logger } from "@/lib/logger";

// Stable empty array to prevent infinite render loop
const EMPTY_SELECTED: GalleryImage[] = [];

interface ImageGalleryPickerProps {
  open: boolean;
  onClose: () => void;
  onConfirm: (images: GalleryImage[]) => void;
  maxSelect?: number;
  initialSelected?: GalleryImage[];
}

interface GalleryResponse {
  data: GalleryImage[];
  meta: GalleryPaginationMeta;
}

export const ImageGalleryPicker: FC<ImageGalleryPickerProps> = ({
  open,
  onClose,
  onConfirm,
  maxSelect = 8,
  initialSelected,
}) => {
  const { token } = theme.useToken();
  const [images, setImages] = useState<GalleryImage[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedImages, setSelectedImages] = useState<GalleryImage[]>(
    initialSelected ?? EMPTY_SELECTED,
  );
  const [meta, setMeta] = useState<GalleryPaginationMeta>({
    page: 1,
    pages: 1,
    total: 0,
    page_size: 20,
  });

  // Debounce timer ref
  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Fetch images from API
  const fetchImages = useCallback(async (page: number, search: string) => {
    setLoading(true);
    try {
      const response = await apiClient.get<GalleryResponse>("/images/gallery", {
        params: {
          page,
          limit: 20,
          search,
        },
      });

      if (response.success && response.data) {
        const galleryData = response.data;
        setImages(galleryData.data || []);
        setMeta(
          galleryData.meta || {
            page: 1,
            pages: 1,
            total: 0,
            page_size: 20,
          },
        );
      }
    } catch (error) {
      message.error("Failed to load images");
      logger.error("Image gallery error", { error });
    } finally {
      setLoading(false);
    }
  }, []);

  // Initial load and reset when modal opens
  // biome-ignore lint/correctness/useExhaustiveDependencies: prevent infinite loop on search change
  useEffect(() => {
    if (open) {
      document.body.style.overflow = "hidden";
      const selected = initialSelected ?? EMPTY_SELECTED;
      setSelectedImages(selected);
      fetchImages(1, searchQuery);
    } else {
      document.body.style.overflow = "";
    }
    return () => {
      document.body.style.overflow = "";
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- searchQuery is intentionally excluded; it is passed directly to fetchImages in the debounce handler to avoid re-triggering on every keystroke
  }, [open, initialSelected, fetchImages]);

  // Debounced search handler
  const handleSearchChange = (e: ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setSearchQuery(value);

    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }

    searchTimeoutRef.current = setTimeout(() => {
      fetchImages(1, value);
    }, 500);
  };

  const handlePageChange = (page: number) => {
    fetchImages(page, searchQuery);
  };

  // Upload handler
  const handleUpload = async (options: UploadRequestOption) => {
    const { file, onSuccess, onError } = options;
    const formData = new FormData();
    formData.append("image", file); // Must be 'image' per requirements

    try {
      const response = await apiClient.post("/images/upload", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      if (response.success) {
        message.success("Image uploaded successfully");
        onSuccess?.(response.data);
        // Refresh gallery
        fetchImages(1, searchQuery);
      } else {
        throw new Error(response.error || "Upload failed");
      }
    } catch (err) {
      const error = err as Error;
      message.error(error.message || "Failed to upload image");
      onError?.(error);
    }
  };

  // Selection Logic
  const toggleSelection = (img: GalleryImage) => {
    const isSelected = selectedImages.some((i) => i.id === img.id);

    if (isSelected) {
      setSelectedImages((prev) => prev.filter((i) => i.id !== img.id));
    } else {
      if (selectedImages.length >= maxSelect) {
        message.warning(`Maximum ${maxSelect} images allowed`);
        return;
      }
      setSelectedImages((prev) => [...prev, img]);
    }
  };

  const handleConfirm = () => {
    onConfirm(selectedImages);
    onClose();
  };

  return (
    <Modal
      open={open}
      onCancel={onClose}
      title="Image Gallery"
      width="90%"
      style={{ maxWidth: 900, top: 20 }}
      footer={null}
      destroyOnHidden
      maskClosable={false} // Prevent accidental close during selection
      classNames={{
        body: "gallery-modal-body",
      }}
      styles={{
        body: {
          height: "70vh",
          display: "flex",
          flexDirection: "column",
          padding: 0,
          overflow: "hidden",
        },
      }}
    >
      <GalleryToolbar
        searchQuery={searchQuery}
        onSearchChange={handleSearchChange}
        onUpload={handleUpload}
      />

      <div
        style={{
          flex: 1,
          overflowY: "auto",
          padding: token.padding,
          backgroundColor: token.colorBgContainer,
        }}
      >
        <GalleryGrid
          loading={loading}
          images={images}
          searchQuery={searchQuery}
          selectedImages={selectedImages}
          onToggle={toggleSelection}
        />
      </div>

      <GalleryFooter
        selectedCount={selectedImages.length}
        maxSelect={maxSelect}
        meta={meta}
        onPageChange={handlePageChange}
        onClose={onClose}
        onConfirm={handleConfirm}
        confirmDisabled={selectedImages.length === 0}
      />
    </Modal>
  );
};
