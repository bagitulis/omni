import { Result, theme } from "antd";
import { VideoCameraOutlined } from "@ant-design/icons";

export const TiktokReportPage = () => {
  const { token } = theme.useToken();

  return (
    <div style={{ padding: 24 }}>
      <Result
        icon={
          <VideoCameraOutlined
            style={{ fontSize: 64, color: token.colorTextSecondary }}
          />
        }
        title="TikTok Report"
        subTitle="This feature has been migrated to the standalone Analytics project."
      />
    </div>
  );
};

export default TiktokReportPage;
