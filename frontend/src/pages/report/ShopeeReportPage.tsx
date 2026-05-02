import { Result, theme } from "antd";
import { ShopOutlined } from "@ant-design/icons";

export const ShopeeReportPage = () => {
  const { token } = theme.useToken();

  return (
    <div style={{ padding: 24 }}>
      <Result
        icon={
          <ShopOutlined
            style={{ fontSize: 64, color: token.colorTextSecondary }}
          />
        }
        title="Shopee Report"
        subTitle="This feature has been migrated to the standalone Analytics project."
      />
    </div>
  );
};

export default ShopeeReportPage;
