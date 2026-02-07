import { Button, Card, Empty, Typography, theme } from "antd";
import { InboxOutlined } from "@ant-design/icons";
import { SHOPEE_ORANGE } from "./types";

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
        image={<InboxOutlined style={{ fontSize: 64, color: SHOPEE_ORANGE }} />}
        description={
          <span>
            No Shopee Ads data available.
            <br />
            <Text type="secondary">
              Go to the Upload tab to import your Shopee Ads CSV export.
            </Text>
          </span>
        }
      >
        <Button
          type="primary"
          style={{ backgroundColor: SHOPEE_ORANGE }}
          onClick={onUploadClick}
        >
          Upload CSV
        </Button>
      </Empty>
    </Card>
  );
};
