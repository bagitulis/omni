import type { FC } from "react";
import { Button, Pagination, Typography, theme } from "antd";
import { CheckOutlined, CloseCircleOutlined } from "@ant-design/icons";
import type { GalleryPaginationMeta } from "@/types/shared";

const { Text } = Typography;

interface GalleryFooterProps {
  selectedCount: number;
  maxSelect: number;
  meta: GalleryPaginationMeta;
  onPageChange: (page: number, pageSize: number) => void;
  onClose: () => void;
  onConfirm: () => void;
  confirmDisabled: boolean;
}

export const GalleryFooter: FC<GalleryFooterProps> = ({
  selectedCount,
  maxSelect,
  meta,
  onPageChange,
  onClose,
  onConfirm,
  confirmDisabled,
}) => {
  const { token } = theme.useToken();

  return (
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
      <div style={{ display: "flex", gap: token.margin, alignItems: "center" }}>
        <Text strong>
          {selectedCount} / {maxSelect} selected
        </Text>
      </div>

      <div style={{ display: "flex", gap: token.margin, alignItems: "center" }}>
        <Pagination
          simple
          current={meta.page}
          total={meta.total}
          pageSize={meta.page_size}
          onChange={onPageChange}
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
          onClick={onConfirm}
          icon={<CheckOutlined />}
          disabled={confirmDisabled}
        >
          Confirm
        </Button>
      </div>
    </div>
  );
};
