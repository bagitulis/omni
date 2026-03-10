import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Typography, Tag, theme } from "antd";
import type { GraphNode } from "./GraphViewHelpers";

const { Text } = Typography;
const { useToken } = theme;

/**
 * Custom React Flow node for frontend components.
 */
export const ComponentNode = ({ data }: NodeProps<GraphNode>) => {
  const { token } = useToken();
  return (
    <div
      style={{
        height: "100%",
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
      }}
    >
      <Text strong style={{ fontSize: 12 }}>
        {data.label}
      </Text>
      {data.details && (
        <Text type="secondary" style={{ fontSize: 10 }} ellipsis>
          {data.details}
        </Text>
      )}
      <Handle
        type="source"
        position={Position.Right}
        style={{ background: token.colorBorder }}
      />
    </div>
  );
};

/**
 * Custom React Flow node for API routes.
 */
export const RouteNode = ({ data }: NodeProps<GraphNode>) => {
  const { token } = useToken();
  const handleColor = data.color || token.colorPrimary;
  const textColor = data.color || token.colorText;
  return (
    <div
      style={{ height: "100%", display: "flex", alignItems: "center", gap: 4 }}
    >
      <Handle
        type="target"
        position={Position.Left}
        style={{ background: handleColor }}
      />
      <Tag
        color={data.details}
        style={{ margin: 0, fontSize: 10, padding: "0 4px", lineHeight: "16px" }}
      >
        {data.method}
      </Tag>
      <Text style={{ fontSize: 11, color: textColor }} ellipsis>
        {data.label}
      </Text>
    </div>
  );
};
