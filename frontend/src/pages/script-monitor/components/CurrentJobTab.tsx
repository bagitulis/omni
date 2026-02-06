import {
  Card,
  Descriptions,
  Button,
  Tag,
  Space,
  Typography,
  Empty,
} from "antd";
import { Job } from "@/types/scriptMonitor";
import { PauseCircleOutlined, StopOutlined } from "@ant-design/icons";

interface Props {
  job: Job | null;
  onCancel: (id: string) => void;
  onForceCancel: (id: string) => void;
}

export function CurrentJobTab({ job, onCancel, onForceCancel }: Props) {
  if (!job) {
    return (
      <Card>
        <Empty description="No active job running" />
      </Card>
    );
  }

  return (
    <Card title="Current Job Execution">
      <Space direction="vertical" style={{ width: "100%" }} size="large">
        <Descriptions bordered column={1}>
          <Descriptions.Item label="Job ID">{job.id}</Descriptions.Item>
          <Descriptions.Item label="Type">
            <Tag color="blue">{job.type}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Status">
            <Tag color={job.status === "processing" ? "processing" : "default"}>
              {job.status.toUpperCase()}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Priority">
            <Tag color={job.priority === "high" ? "red" : "orange"}>
              {job.priority.toUpperCase()}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Started At">
            {job.started_at ? new Date(job.started_at).toLocaleString() : "-"}
          </Descriptions.Item>
          <Descriptions.Item label="Job Data">
            <pre style={{ maxHeight: "200px", overflow: "auto" }}>
              {JSON.stringify(job.data, null, 2)}
            </pre>
          </Descriptions.Item>
        </Descriptions>

        <Card type="inner" title="Execution Control">
          <Space>
            <Button
              danger
              icon={<PauseCircleOutlined />}
              onClick={() => onCancel(job.id)}
            >
              Cancel Job
            </Button>
            <Button
              danger
              type="primary"
              icon={<StopOutlined />}
              onClick={() => onForceCancel(job.id)}
            >
              Force Cancel
            </Button>
          </Space>
          <Typography.Text
            type="secondary"
            style={{ display: "block", marginTop: 8 }}
          >
            Force cancel should only be used if the job is stuck and not
            responding to normal cancellation.
          </Typography.Text>
        </Card>
      </Space>
    </Card>
  );
}
