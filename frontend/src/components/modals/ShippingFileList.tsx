import { Button, Table, Typography, Flex, Spin } from "antd";
import { ReloadOutlined } from "@ant-design/icons";
import type { ShippingFileResponse } from "@/api/client";

const { Title, Text } = Typography;

interface ShippingFileListProps {
  filesData: ShippingFileResponse | undefined;
  isLoadingFiles: boolean;
  onRefetch: () => void;
  onProcessFile: (filename: string) => void;
}

export function ShippingFileList({
  filesData,
  isLoadingFiles,
  onRefetch,
  onProcessFile,
}: ShippingFileListProps) {
  const columns = [
    {
      title: "Filename",
      dataIndex: "name",
      key: "name",
      render: (text: string) => <Text strong>{text}</Text>,
    },
    {
      title: "Action",
      key: "action",
      width: 120,
      render: (_: unknown, record: { name: string }) => (
        <Button
          type="primary"
          size="small"
          onClick={() => onProcessFile(record.name)}
        >
          Process
        </Button>
      ),
    },
  ];

  const files = filesData?.data?.files;
  const tableData =
    (Array.isArray(files) ? files : []).map((file) => ({ key: file, name: file }));

  return (
    <div style={{ paddingTop: 16, paddingBottom: 16 }}>
      <Flex justify="space-between" align="center" style={{ marginBottom: 16 }}>
        <Title level={5} style={{ margin: 0 }}>
          Available Files
        </Title>
        <Button
          icon={<ReloadOutlined />}
          onClick={() => onRefetch()}
          loading={isLoadingFiles}
        >
          Refresh
        </Button>
      </Flex>

      {isLoadingFiles ? (
        <div style={{ textAlign: "center", paddingTop: 32, paddingBottom: 32 }}>
          <Spin tip="Loading files..." />
        </div>
      ) : (
        <Table
          dataSource={tableData}
          columns={columns}
          pagination={{ pageSize: 5 }}
          size="small"
          bordered
          locale={{ emptyText: "No shipping files found" }}
        />
      )}
    </div>
  );
}
