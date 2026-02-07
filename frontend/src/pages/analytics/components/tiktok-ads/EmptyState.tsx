import { Card, Empty, Button, Typography, theme } from "antd";
import { InboxOutlined } from "@ant-design/icons";
import { TIKTOK_BLACK } from "./types";

const { Text } = Typography;
const { useToken } = theme;

interface EmptyStateProps {
  onUploadClick: () => void;
}

export const EmptyState = ({ onUploadClick }: EmptyStateProps) => {
  const { token } = useToken();

  return (
    <Card style={{ borderRadius: token.borderRadius }}>
      <Empty
        image={<InboxOutlined style={{ fontSize: 64, color: TIKTOK_BLACK }} />}
        description={
          <span>
            No TikTok Ads data available.
            <br />
            <Text type="secondary">
              Go to the Upload tab to import your TikTok Ads CSV export.
            </Text>
          </span>
        }
      >
        <Button
          type="primary"
          style={{ backgroundColor: TIKTOK_BLACK }}
          onClick={onUploadClick}
        >
          Upload CSV
        </Button>
      </Empty>
    </Card>
  );
};
