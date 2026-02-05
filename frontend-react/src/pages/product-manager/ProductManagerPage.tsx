import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Tabs, Card, Empty, Typography } from "antd";

const { Title } = Typography;

export function ProductManagerPage() {
  const navigate = useNavigate();
  const { platform } = useParams<{ platform: string }>();
  const [activeTab, setActiveTab] = useState<string>("shopee");

  useEffect(() => {
    if (platform) {
      setActiveTab(platform);
    }
  }, [platform]);

  const handleTabChange = (key: string) => {
    setActiveTab(key);
    navigate(`/product-manager/${key}`);
  };

  const items = [
    {
      key: "shopee",
      label: "Shopee",
      children: (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="Connect your Shopee account to view products"
        />
      ),
    },
    {
      key: "lazada",
      label: "Lazada",
      children: (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="Connect your Lazada account to view products"
        />
      ),
    },
    {
      key: "tiktok",
      label: "TikTok",
      children: (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="Connect your TikTok account to view products"
        />
      ),
    },
  ];

  return (
    <div className="p-6">
      <Title level={2}>Product Manager</Title>
      <Card>
        <Tabs activeKey={activeTab} onChange={handleTabChange} items={items} />
      </Card>
    </div>
  );
}
