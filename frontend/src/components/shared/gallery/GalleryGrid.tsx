import type { FC } from "react";
import { Empty, Spin, Typography, theme } from "antd";
import { CheckOutlined, FileImageOutlined } from "@ant-design/icons";
import { type GalleryImage, getImageUrl } from "../../../types/shared";

const { Text } = Typography;

interface GalleryGridProps {
  loading: boolean;
  images: GalleryImage[];
  searchQuery: string;
  selectedImages: GalleryImage[];
  onToggle: (img: GalleryImage) => void;
}

export const GalleryGrid: FC<GalleryGridProps> = ({
  loading,
  images,
  searchQuery,
  selectedImages,
  onToggle,
}) => {
  const { token } = theme.useToken();

  const isSelected = (img: GalleryImage) =>
    selectedImages.some((i) => i.id === img.id);

  if (loading) {
    return (
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
    );
  }

  if (images.length === 0) {
    return (
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
    );
  }

  return (
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
          <button
            key={img.id}
            type="button"
            onClick={() => onToggle(img)}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                onToggle(img);
              }
            }}
            style={{
              background: "none",
              padding: 0,
              width: "100%",
              position: "relative",
              aspectRatio: "1",
              borderRadius: 3,
              overflow: "hidden",
              cursor: "pointer",
              border: selected
                ? "3px solid #ff6b2c" // Shopee brand color per design spec — exception to hardcoded color rule
                : `1px solid ${token.colorBorder}`,
              transition: "all 0.2s",
              boxShadow: selected ? token.boxShadow : "none",
            }}
          >
            <img
              src={getImageUrl(img.local_path, "medium")}
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
                  backgroundColor: "#ff6b2c", // Shopee brand color per design spec — exception to hardcoded color rule
                  color: "#fff",
                  borderRadius: "50%", // Intentional circle — 50% is correct for circular badge
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  boxShadow: "0 2px 4px rgba(0,0,0,0.2)",
                }}
              >
                <CheckOutlined style={{ fontSize: 14, fontWeight: "bold" }} />
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
          </button>
        );
      })}
    </div>
  );
};
