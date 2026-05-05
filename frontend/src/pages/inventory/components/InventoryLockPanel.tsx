import { useMemo } from "react";
import {
  Button,
  Checkbox,
  Collapse,
  Empty,
  Flex,
  Grid,
  Space,
  Tag,
  Typography,
  theme,
} from "antd";
import { LockOutlined, UnlockOutlined } from "@ant-design/icons";

const { Panel } = Collapse;

interface InventoryLockPanelProps {
  availableColumns: string[];
  visibleColumns: string[];
  lockedColumns: string[];
  onChangeLockedColumns: (columns: string[]) => void;
}

export function InventoryLockPanel({
  availableColumns,
  visibleColumns,
  lockedColumns,
  onChangeLockedColumns,
}: InventoryLockPanelProps) {
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;

  const {
    token: { colorTextSecondary, colorBgContainerDisabled },
  } = theme.useToken();
  const lockableColumns = useMemo(
    () => (visibleColumns.length > 0 ? visibleColumns : availableColumns),
    [availableColumns, visibleColumns],
  );

  const sortedColumns = useMemo(
    () => [...lockableColumns].sort((a, b) => a.localeCompare(b)),
    [lockableColumns],
  );

  const lockedCount = lockedColumns.length;

  const toggleColumnLock = (column: string, checked: boolean) => {
    onChangeLockedColumns(
      checked
        ? [...lockedColumns, column]
        : lockedColumns.filter((item) => item !== column),
    );
  };

  const handleLockAll = () => {
    onChangeLockedColumns(sortedColumns);
  };

  const handleUnlockAll = () => {
    onChangeLockedColumns([]);
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

          {sortedColumns.length === 0 ? (
            <Empty description="No columns available for configuration" />
          ) : (
            <div
              style={{
                display: "grid",
                gridTemplateColumns: isMobile
                  ? "1fr"
                  : "repeat(auto-fill, minmax(180px, 1fr))",
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
                      padding: 8,
                    }}
                  >
                    {column}
                  </Checkbox>
                );
              })}
            </div>
          )}

          <Space wrap>
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
