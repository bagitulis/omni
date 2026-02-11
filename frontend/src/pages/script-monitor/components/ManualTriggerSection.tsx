import { Button, Card, Popconfirm, Space, Tag, Typography } from "antd";
import { PlayCircleOutlined } from "@ant-design/icons";
import type { AutoFunctionConfig } from "@/types/scriptMonitor";

interface Props {
  configs: AutoFunctionConfig[];
  onTrigger: (name: string) => void;
}

export function ManualTriggerSection({ configs, onTrigger }: Props) {
  return (
    <Card title="Manual Trigger" style={{ marginTop: 16 }}>
      <Typography.Paragraph type="secondary" style={{ marginBottom: 16 }}>
        Trigger an auto-function immediately without waiting for the next
        schedule.
      </Typography.Paragraph>

      <Space wrap>
        {configs.map((config) => (
          <Popconfirm
            key={config.id}
            title={`Trigger ${config.name} now?`}
            onConfirm={() => onTrigger(config.name)}
          >
            <Button icon={<PlayCircleOutlined />}>
              <Space size={8}>
                <Tag color="blue" style={{ marginInlineEnd: 0 }}>
                  {config.name}
                </Tag>
                <span>Trigger Now</span>
              </Space>
            </Button>
          </Popconfirm>
        ))}
      </Space>

      {!configs.length && (
        <Typography.Text type="secondary">
          No auto-function configuration available.
        </Typography.Text>
      )}
    </Card>
  );
}
