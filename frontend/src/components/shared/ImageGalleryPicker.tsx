import React, { useState, useEffect, useCallback, useRef } from "react";
import {
  Modal,
  Button,
  Input,
  Pagination,
  Upload,
  message,
  Spin,
  Empty,
  Typography,
  theme,
} from "antd";
import {
  SearchOutlined,
  UploadOutlined,
  FileImageOutlined,
  CheckOutlined,
  CloseCircleOutlined,
} from "@ant-design/icons";
import type { UploadRequestOption } from "rc-upload/lib/interface";
import apiClient from "@/api/client";
import {
  GalleryImage,
  GalleryPaginationMeta,
  getImageUrl,
} from "../../types/shared";

const { Text } = Typography;

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

export const ImageGalleryPicker: React.FC<ImageGalleryPickerProps> = ({
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
  const searchTimeoutRef = useRef<NodeJS.Timeout | null>(null);

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
      console.error(error);
    } finally {
      setLoading(false);
    }
  }, []);

  // Initial load and reset when modal opens
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, initialSelected]);

  // Debounced search handler
  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
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

  const isSelected = (img: GalleryImage) =>
    selectedImages.some((i) => i.id === img.id);

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
      {/* Toolbar */}
      <div
        style={{
          padding: token.padding,
          borderBottom: `1px solid ${token.colorBorderSecondary}`,
          display: "flex",
          gap: token.marginSM,
          alignItems: "center",
          backgroundColor: token.colorBgLayout,
        }}
      >
        <Input
          placeholder="Search images..."
          prefix={
            <SearchOutlined style={{ color: token.colorTextDescription }} />
          }
          value={searchQuery}
          onChange={handleSearchChange}
          style={{ flex: 1, maxWidth: 300 }}
          allowClear
        />
        <div style={{ flex: 1 }} />
        <Upload
          customRequest={handleUpload}
          showUploadList={false}
          accept="image/*"
        >
          <Button type="primary" icon={<UploadOutlined />}>
            Upload
          </Button>
        </Upload>
      </div>

      {/* Content */}
      <div
        style={{
          flex: 1,
          overflowY: "auto",
          padding: token.padding,
          backgroundColor: token.colorBgContainer,
        }}
      >
        {loading ? (
          <div
            style={{
              height: "100%",
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
              flexDirection: "column",
              gap: token.margin,
            }}
          >
            <Spin size="large" />
            <Text type="secondary">Loading images...</Text>
          </div>
        ) : images.length === 0 ? (
          <div
            style={{
              height: "100%",
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
            }}
          >
            <Empty
              image={
                <FileImageOutlined
                  style={{ fontSize: 48, color: token.colorTextQuaternary }}
                />
              }
              description={
                searchQuery ? "No matching images found" : "Gallery is empty"
              }
            />
          </div>
        ) : (
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fill, minmax(140px, 1fr))",
              gap: token.margin,
            }}
          >
            {images.map((img) => {
              const selected = isSelected(img);
              return (
                <div
                  key={img.id}
                  role="button"
                  tabIndex={0}
                  onClick={() => toggleSelection(img)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault();
                      toggleSelection(img);
                    }
                  }}
                  style={{
                    position: "relative",
                    aspectRatio: "1",
                    borderRadius: 3,
                    overflow: "hidden",
                    cursor: "pointer",
                    border: selected
                      ? "3px solid #ff6b2c" // Shopee orange per spec
                      : `1px solid ${token.colorBorder}`,
                    transition: "all 0.2s",
                    boxShadow: selected ? token.boxShadow : "none",
                  }}
                >
                  <img
                    src={getImageUrl(img.local_path, "thumb")}
                    alt={img.filename}
                    loading="lazy"
                    style={{
                      width: "100%",
                      height: "100%",
                      objectFit: "cover", // NOT contain per constraints
                      display: "block",
                    }}
                  />

                  {/* Selection Indicator Overlay */}
                  {selected && (
                    <div
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 4,
                        width: 24,
                        height: 24,
                        backgroundColor: "#ff6b2c",
                        color: "#fff",
                        borderRadius: "50%",
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        boxShadow: "0 2px 4px rgba(0,0,0,0.2)",
                      }}
                    >
                      <CheckOutlined
                        style={{ fontSize: 14, fontWeight: "bold" }}
                      />
                    </div>
                  )}

                  {/* Filename Overlay on Hover/Always */}
                  <div
                    style={{
                      position: "absolute",
                      bottom: 0,
                      left: 0,
                      right: 0,
                      backgroundColor: "rgba(0,0,0,0.6)",
                      color: "#fff",
                      padding: "4px 8px",
                      fontSize: "10px",
                      whiteSpace: "nowrap",
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                    }}
                    title={img.filename}
                  >
                    {img.filename}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Footer */}
      <div
        style={{
          padding: `${token.paddingSM}px ${token.padding}px`,
          borderTop: `1px solid ${token.colorBorderSecondary}`,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          backgroundColor: token.colorBgLayout,
        }}
      >
        <div
          style={{ display: "flex", gap: token.margin, alignItems: "center" }}
        >
          <Text strong>
            {selectedImages.length} / {maxSelect} selected
          </Text>
        </div>

        <div
          style={{ display: "flex", gap: token.margin, alignItems: "center" }}
        >
          <Pagination
            simple
            current={meta.page}
            total={meta.total}
            pageSize={meta.page_size}
            onChange={handlePageChange}
            size="small"
          />
          <div
            style={{
              width: 1,
              height: 24,
              backgroundColor: token.colorBorder,
              margin: `0 ${token.marginXS}px`,
            }}
          />
          <Button onClick={onClose} icon={<CloseCircleOutlined />}>
            Cancel
          </Button>
          <Button
            type="primary"
            onClick={handleConfirm}
            icon={<CheckOutlined />}
            disabled={selectedImages.length === 0}
          >
            Confirm
          </Button>
        </div>
      </div>
    </Modal>
  );
};
