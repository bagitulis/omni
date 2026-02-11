import { Tag, Typography } from "antd";
import { DisconnectOutlined } from "@ant-design/icons";

const { Text } = Typography;

export const columnsConnected = [
  {
    title: "Endpoint",
    dataIndex: "endpoint",
    key: "endpoint",
    render: (text: string) => (
      <Text code copyable>
        {text}
      </Text>
    ),
  },
  {
    title: "Method",
    dataIndex: "method",
    key: "method",
    width: 100,
    render: (method: string) => (
      <Tag
        color={
          method === "GET" ? "blue" : method === "POST" ? "green" : "orange"
        }
      >
        {method}
      </Tag>
    ),
  },
  {
    title: "Category",
    dataIndex: "category",
    key: "category",
    render: (cat: string) => <Tag>{cat}</Tag>,
  },
  {
    title: "Type",
    dataIndex: "is_dynamic",
    key: "is_dynamic",
    width: 100,
    render: (dynamic: boolean) => (
      <Tag color={dynamic ? "purple" : "cyan"}>
        {dynamic ? "Dynamic" : "Static"}
      </Tag>
    ),
  },
];

export const columnsFrontendOnly = [
  {
    title: "Endpoint",
    dataIndex: "endpoint",
    key: "endpoint",
    render: (text: string) => (
      <Text code copyable>
        {text}
      </Text>
    ),
  },
  {
    title: "Used In Components",
    dataIndex: "components",
    key: "components",
    render: (components: string[]) => (
      <>
        {components.map((c) => (
          <Tag key={c} color="geekblue">
            {c}
          </Tag>
        ))}
      </>
    ),
  },
  {
    title: "Status",
    key: "status",
    render: () => (
      <Tag color="error" icon={<DisconnectOutlined />}>
        Disconnected
      </Tag>
    ),
  },
];

export const columnsBackendOnly = [
  {
    title: "Endpoint",
    dataIndex: "endpoint",
    key: "endpoint",
    render: (text: string) => (
      <Text code copyable>
        {text}
      </Text>
    ),
  },
  {
    title: "Method",
    dataIndex: "method",
    key: "method",
    render: (method: string) => <Tag>{method}</Tag>,
  },
  {
    title: "Category",
    dataIndex: "category",
    key: "category",
    render: (cat: string) => <Tag>{cat}</Tag>,
  },
];

export const columnsUnused = [
  {
    title: "Endpoint",
    dataIndex: "endpoint",
    key: "endpoint",
    render: (text: string) => (
      <Text code copyable>
        {text}
      </Text>
    ),
  },
  {
    title: "Category",
    dataIndex: "category",
    key: "category",
    render: (cat: string) => <Tag>{cat}</Tag>,
  },
];
