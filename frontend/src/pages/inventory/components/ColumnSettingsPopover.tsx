import {
  Button,
  Checkbox,
  Popover,
  Space,
  Tooltip,
  Typography,
  theme,
} from "antd";
import {
  SettingOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
} from "@ant-design/icons";
import type { InventoryColumnMoveDirection } from "../utils/inventoryColumnOrder";

interface ColumnSettingsPopoverProps {
  selectedColumns: string[];
  hiddenColumns: string[];
  onToggle: (column: string, checked: boolean) => void;
  onMove: (column: string, direction: InventoryColumnMoveDirection) => void;
  isMobile: boolean;
}

/**
 * Popover for managing visible/hidden columns and their order.
 * Extracted from InventoryHeader for SRP.
 */
export function ColumnSettingsPopover({
  selectedColumns,
  hiddenColumns,
  onToggle,
  onMove,
  isMobile,
}: ColumnSettingsPopoverProps) {
  const { token } = theme.useToken();
  const hasColumns = selectedColumns.length > 0 || hiddenColumns.length > 0;

  const content = hasColumns ? (
    <>
      <Typography.Text type="secondary" style={{ fontSize: 12 }}>
        Selected columns (drag order equivalent)
      </Typography.Text>

      {selectedColumns.map((col, index) => (
        <div
          key={col}
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            gap: 8,
          }}
        >
          <Checkbox
            checked
            onChange={(e) => onToggle(col, e.target.checked)}
          >
            {col}
          </Checkbox>
          <Space size={4}>
            <Tooltip title="Move up">
              <Button
                size="small"
                icon={<ArrowUpOutlined />}
                onClick={() => onMove(col, "up")}
                disabled={index === 0}
              />
            </Tooltip>
            <Tooltip title="Move down">
              <Button
                size="small"
                icon={<ArrowDownOutlined />}
                onClick={() => onMove(col, "down")}
                disabled={index === selectedColumns.length - 1}
              />
            </Tooltip>
          </Space>
        </div>
      ))}

      {hiddenColumns.length > 0 && (
        <>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            Hidden columns
          </Typography.Text>
          {hiddenColumns.map((col) => (
            <Checkbox
              key={col}
              checked={false}
              onChange={(e) => onToggle(col, e.target.checked)}
            >
              {col}
            </Checkbox>
          ))}
        </>
      )}
    </>
  ) : (
    <div style={{ padding: 8, color: token.colorTextSecondary }}>
      No columns available
    </div>
  );

  return (
    <Popover
      trigger="click"
      placement="bottomRight"
      title="Columns & Order"
      content={
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            maxHeight: 300,
            overflowY: "auto",
            minWidth: 280,
            gap: 8,
          }}
        >
          {content}
        </div>
      }
    >
      <Button
        icon={<SettingOutlined />}
        size="middle"
        style={
          isMobile
            ? { flex: "1 1 calc(50% - 8px)", minWidth: 100 }
            : undefined
        }
      >
        Cols
      </Button>
    </Popover>
  );
}
