import type { ChangeEvent, FC } from "react";
import { Button, Input, Upload, theme } from "antd";
import { SearchOutlined, UploadOutlined } from "@ant-design/icons";
import type { UploadRequestOption } from "rc-upload/lib/interface";

interface GalleryToolbarProps {
  searchQuery: string;
  onSearchChange: (e: ChangeEvent<HTMLInputElement>) => void;
  onUpload: (options: UploadRequestOption) => void;
}

export const GalleryToolbar: FC<GalleryToolbarProps> = ({
  searchQuery,
  onSearchChange,
  onUpload,
}) => {
  const { token } = theme.useToken();

  return (
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
        onChange={onSearchChange}
        style={{ flex: 1, maxWidth: 300 }}
        allowClear
      />
      <div style={{ flex: 1 }} />
      <Upload customRequest={onUpload} showUploadList={false} accept="image/*">
        <Button type="primary" icon={<UploadOutlined />}>
          Upload
        </Button>
      </Upload>
    </div>
  );
};
