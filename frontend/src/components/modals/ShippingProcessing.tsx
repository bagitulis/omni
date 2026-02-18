import { Spin, Typography } from "antd";

const { Title, Text } = Typography;

interface ShippingProcessingProps {
  selectedOption: "wallet" | "file" | null;
}

export function ShippingProcessing({
  selectedOption,
}: ShippingProcessingProps) {
  return (
    <div style={{ textAlign: "center", paddingTop: 48, paddingBottom: 48 }}>
      <Spin size="large" />
      <div style={{ marginTop: 16 }}>
        <Title level={4}>Processing...</Title>
        <Text type="secondary">
          {selectedOption === "wallet"
            ? "Exporting shipping data to Google Sheets"
            : "Processing shipping file records"}
        </Text>
      </div>
    </div>
  );
}
