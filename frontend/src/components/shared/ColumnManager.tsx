import React, { useState } from "react";
import {
  Button,
  Popover,
  List,
  Checkbox,
  Tooltip,
  Typography,
  Space,
  theme,
} from "antd";
import {
  SettingOutlined,
  HolderOutlined,
  LockOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import type { ColumnConfig } from "@/types/shared";

interface ColumnManagerProps {
  columns: ColumnConfig[];
  onChange: (columns: ColumnConfig[]) => void;
  onReset: () => void;
}

export const ColumnManager: React.FC<ColumnManagerProps> = ({
  columns,
  onChange,
  onReset,
}) => {
  const { token } = theme.useToken();
  const [draggedItemIndex, setDraggedItemIndex] = useState<number | null>(null);
  const [dragOverItemIndex, setDragOverItemIndex] = useState<number | null>(
    null,
  );

  // Filter and sort columns by order to display them correctly
  const sortedColumns = [...columns].sort((a, b) => a.order - b.order);

  const handleToggle = (key: string) => {
    // Check if column is locked before doing anything
    const column = columns.find((c) => c.key === key);
    if (column?.locked) return;

    const newColumns = columns.map((col) => {
      if (col.key === key) {
        return { ...col, visible: !col.visible };
      }
      return col;
    });
    onChange(newColumns);
  };

  const onDragStart = (e: React.DragEvent, index: number) => {
    // Prevent dragging locked columns
    if (sortedColumns[index].locked) {
      e.preventDefault();
      return;
    }
    setDraggedItemIndex(index);
    // Required for Firefox
    e.dataTransfer.effectAllowed = "move";
    // Set a transparent drag image or similar if desired, but default is usually fine
  };

  const onDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault(); // Necessary to allow dropping
    if (draggedItemIndex === null) return;
    if (sortedColumns[index].locked) return; // Cannot drop onto/swap with locked column

    setDragOverItemIndex(index);
  };

  const onDragEnd = () => {
    setDraggedItemIndex(null);
    setDragOverItemIndex(null);
  };

  const onDrop = (e: React.DragEvent, dropIndex: number) => {
    e.preventDefault();
    if (draggedItemIndex === null) return;
    if (draggedItemIndex === dropIndex) return;

    // Perform the reorder
    const newOrderedColumns = [...sortedColumns];
    const [draggedItem] = newOrderedColumns.splice(draggedItemIndex, 1);
    newOrderedColumns.splice(dropIndex, 0, draggedItem);

    // Update the 'order' property for all columns based on new index
    const updatedColumns = newOrderedColumns.map((col, index) => ({
      ...col,
      order: index,
    }));

    // We need to pass back the full list including any that might have been filtered out
    // But here we are working with all columns (just sorted).
    // So we just need to ensure we map back to the original full set if 'sortedColumns' wasn't all.
    // In this component, we assume 'columns' prop contains all columns.

    // Re-merge with original unsorted list isn't needed if we reconstruct 'columns' from 'updatedColumns'
    // essentially we are replacing the whole state.
    onChange(updatedColumns);

    setDraggedItemIndex(null);
    setDragOverItemIndex(null);
  };

  const content = (
    <div style={{ width: 300 }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: token.marginSM,
          borderBottom: `1px solid ${token.colorBorderSecondary}`,
          paddingBottom: token.paddingXS,
        }}
      >
        <Typography.Text strong>Column Settings</Typography.Text>
        <Button
          type="text"
          size="small"
          icon={<ReloadOutlined />}
          onClick={onReset}
          style={{ fontSize: token.fontSizeSM }}
        >
          Reset Defaults
        </Button>
      </div>

      <List
        size="small"
        dataSource={sortedColumns}
        renderItem={(item, index) => {
          const isLocked = !!item.locked;
          const isDragging = draggedItemIndex === index;
          const isDragOver = dragOverItemIndex === index;

          return (
            <List.Item
              style={{
                padding: "8px 0",
                backgroundColor: isDragOver
                  ? token.colorBgLayout
                  : "transparent",
                opacity: isDragging ? 0.5 : 1,
                cursor: isLocked ? "default" : "move",
                borderBottom: isDragOver
                  ? `2px solid ${token.colorPrimary}`
                  : "none",
                transition: "all 0.2s",
              }}
              draggable={!isLocked}
              onDragStart={(e) => onDragStart(e, index)}
              onDragOver={(e) => onDragOver(e, index)}
              onDragEnd={onDragEnd}
              onDrop={(e) => onDrop(e, index)}
              // Add data-testid for testing
              data-testid={`column-item-${item.key}`}
            >
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  width: "100%",
                }}
              >
                {/* Drag Handle */}
                <span
                  style={{
                    marginRight: token.marginXS,
                    color: isLocked
                      ? token.colorTextQuaternary
                      : token.colorTextSecondary,
                    cursor: isLocked ? "not-allowed" : "grab",
                    display: "flex",
                    alignItems: "center",
                  }}
                  data-testid={`drag-handle-${item.key}`}
                >
                  <HolderOutlined />
                </span>

                {/* Checkbox */}
                <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
                  {isLocked ? (
                    <Tooltip title="This column cannot be hidden">
                      {/* Wrapper for disabled element tooltip */}
                      <span
                        style={{
                          display: "inline-block",
                          pointerEvents: "none",
                        }}
                      >
                        <Checkbox
                          checked={item.visible}
                          disabled={true}
                          data-testid={`checkbox-${item.key}`}
                          // Explicitly prevent change even if disabled prop fails in test env
                          onChange={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                          }}
                        >
                          <Space size={4}>
                            <Typography.Text
                              style={{
                                color: isLocked
                                  ? token.colorTextSecondary
                                  : token.colorText,
                              }}
                            >
                              {item.title}
                            </Typography.Text>
                            <LockOutlined
                              style={{
                                fontSize: 10,
                                color: token.colorTextTertiary,
                              }}
                            />
                          </Space>
                        </Checkbox>
                      </span>
                    </Tooltip>
                  ) : (
                    <Checkbox
                      checked={item.visible}
                      onChange={() => handleToggle(item.key)}
                      data-testid={`checkbox-${item.key}`}
                    >
                      <Typography.Text>{item.title}</Typography.Text>
                    </Checkbox>
                  )}
                </div>
              </div>
            </List.Item>
          );
        }}
      />
    </div>
  );

  return (
    <Popover
      content={content}
      trigger="click"
      placement="bottomRight"
      arrow={false}
      // Destroy on close to reset any drag state if closed mid-drag
      destroyOnHidden={true}
    >
      <Button icon={<SettingOutlined />} data-testid="column-manager-trigger">
        Columns
      </Button>
    </Popover>
  );
};
