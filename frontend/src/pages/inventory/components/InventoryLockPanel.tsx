import { useMemo, useState } from "react";
import {
  Button,
  Checkbox,
  Collapse,
  Empty,
  Flex,
  Space,
  Tag,
  Typography,
  theme,
} from "antd";
import { LockOutlined, UnlockOutlined } from "@ant-design/icons";
import { useAvailableColumns } from "@/hooks/useInventory";

const { Panel } = Collapse;

export function InventoryLockPanel() {
  const {
    token: { colorTextSecondary, colorBgContainerDisabled },
  } = theme.useToken();
  const { data: availableColumns = [], isLoading } = useAvailableColumns();
  const [lockedColumns, setLockedColumns] = useState<string[]>([]);

  const sortedColumns = useMemo(
    () => [...availableColumns].sort((a, b) => a.localeCompare(b)),
    [availableColumns],
  );

  const lockedCount = lockedColumns.length;

  const toggleColumnLock = (column: string, checked: boolean) => {
    setLockedColumns((previous) => {
      if (checked) {
        if (previous.includes(column)) {
          return previous;
        }
        return [...previous, column];
      }
      return previous.filter((item) => item !== column);
    });
  };

  const handleLockAll = () => {
    setLockedColumns(sortedColumns);
  };

  const handleUnlockAll = () => {
    setLockedColumns([]);
  };

  return (
    <Collapse
      defaultActiveKey={["inventory-lock"]}
      style={{ marginBottom: 16 }}
    >
      <Panel
        key="inventory-lock"
        header={
          <Flex align="center" justify="space-between">
            <Typography.Text strong>Lock Column Edit</Typography.Text>
            <Typography.Text type="secondary">
              Locked Columns: {lockedCount}
            </Typography.Text>
          </Flex>
        }
      >
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            Select columns that should not be editable.
          </Typography.Text>

          {sortedColumns.length === 0 && !isLoading ? (
            <Empty description="No columns available for configuration" />
          ) : (
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fill, minmax(180px, 1fr))",
                gap: 8,
              }}
            >
              {sortedColumns.map((column) => {
                const isChecked = lockedColumns.includes(column);
                return (
                  <Checkbox
                    key={column}
                    checked={isChecked}
                    onChange={(event) =>
                      toggleColumnLock(column, event.target.checked)
                    }
                    style={{
                      marginInlineStart: 0,
                      background: isChecked
                        ? colorBgContainerDisabled
                        : "transparent",
                      borderRadius: 3,
                      padding: 6,
                    }}
                  >
                    {column}
                  </Checkbox>
                );
              })}
            </div>
          )}

          <Space>
            <Button
              icon={<LockOutlined />}
              disabled={
                sortedColumns.length === 0 ||
                lockedCount === sortedColumns.length
              }
              onClick={handleLockAll}
            >
              Lock All
            </Button>
            <Button
              icon={<UnlockOutlined />}
              disabled={lockedCount === 0}
              onClick={handleUnlockAll}
            >
              Unlock All
            </Button>
          </Space>

          <div>
            <Typography.Text
              style={{ fontSize: 12, color: colorTextSecondary }}
            >
              Locked list
            </Typography.Text>
            <div style={{ marginTop: 8 }}>
              {lockedCount > 0 ? (
                <Space size={[4, 8]} wrap>
                  {lockedColumns.map((column) => (
                    <Tag key={column} color="error" style={{ borderRadius: 3 }}>
                      {column}
                    </Tag>
                  ))}
                </Space>
              ) : (
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  All columns are editable
                </Typography.Text>
              )}
            </div>
          </div>
        </Space>
      </Panel>
    </Collapse>
  );
}
