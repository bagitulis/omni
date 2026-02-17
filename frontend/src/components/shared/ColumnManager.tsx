import React, { useMemo, useState } from "react";
import { Button, Checkbox, Popover, Tooltip, Typography, theme } from "antd";
import {
  HolderOutlined,
  LockOutlined,
  SettingOutlined,
  UndoOutlined,
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
  const [dragIndex, setDragIndex] = useState<number | null>(null);

  const sortedColumns = useMemo(
    () => [...columns].sort((a, b) => a.order - b.order),
    [columns],
  );

  const toggleColumnVisibility = (columnKey: string) => {
    const nextColumns = columns.map((column) => {
      if (column.key !== columnKey || column.locked) {
        return column;
      }

      return {
        ...column,
        visible: !column.visible,
      };
    });

    onChange(nextColumns);
  };

  const reorderColumns = (fromIndex: number, toIndex: number) => {
    if (fromIndex === toIndex) {
      return;
    }

    const fromColumn = sortedColumns[fromIndex];
    const toColumn = sortedColumns[toIndex];

    if (!fromColumn || !toColumn || fromColumn.locked || toColumn.locked) {
      return;
    }

    const nextColumns = [...sortedColumns];
    const [movedColumn] = nextColumns.splice(fromIndex, 1);
    nextColumns.splice(toIndex, 0, movedColumn);

    onChange(
      nextColumns.map((column, index) => ({
        ...column,
        order: index,
      })),
    );
  };

  const handleDragStart = (
    event: React.DragEvent<HTMLElement>,
    index: number,
  ) => {
    if (sortedColumns[index]?.locked) {
      event.preventDefault();
      return;
    }

    event.dataTransfer.effectAllowed = "move";
    setDragIndex(index);
  };

  const handleDragOver = (
    event: React.DragEvent<HTMLElement>,
    index: number,
  ) => {
    event.preventDefault();

    if (dragIndex === null) {
      return;
    }

    reorderColumns(dragIndex, index);
    setDragIndex(index);
  };

  const handleDragEnd = () => {
    setDragIndex(null);
  };

  const content = (
    <div style={{ width: 240 }} data-testid="column-manager-content">
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
        <Typography.Text strong style={{ fontSize: token.fontSize }}>
          Columns
        </Typography.Text>
        <Tooltip title="Reset to default columns">
          <Button
            type="text"
            size="small"
            icon={<UndoOutlined />}
            onClick={onReset}
            aria-label="Reset columns"
            data-testid="column-reset-button"
          />
        </Tooltip>
      </div>

      <ul
        style={{
          maxHeight: 300,
          overflowY: "auto",
          display: "flex",
          flexDirection: "column",
          gap: 4,
          listStyle: "none",
          margin: 0,
          padding: 0,
        }}
      >
        {sortedColumns.map((column, index) => {
          const isLocked = Boolean(column.locked);
          const isDragging = dragIndex === index;

          return (
            <li
              key={column.key}
              draggable={!isLocked}
              onDragStart={(event) => handleDragStart(event, index)}
              onDragOver={(event) => handleDragOver(event, index)}
              onDragEnd={handleDragEnd}
              data-testid={`column-row-${column.key}`}
              style={{
                display: "flex",
                alignItems: "center",
                gap: 8,
                padding: "6px 4px",
                borderRadius: 3,
                background: isDragging ? token.colorFillAlter : "transparent",
                opacity: isDragging ? 0.6 : 1,
                cursor: isLocked ? "default" : "grab",
              }}
            >
              <span
                style={{
                  display: "inline-flex",
                  alignItems: "center",
                  color: isLocked
                    ? token.colorTextQuaternary
                    : token.colorTextSecondary,
                  width: 14,
                }}
                data-testid={`drag-handle-${column.key}`}
                aria-hidden="true"
              >
                {!isLocked ? <HolderOutlined /> : null}
              </span>

              <Checkbox
                checked={column.visible}
                disabled={isLocked}
                onChange={() => toggleColumnVisibility(column.key)}
                data-testid={`column-checkbox-${column.key}`}
              >
                <span style={{ fontSize: token.fontSize }}>
                  {column.title}
                  {isLocked ? (
                    <span
                      data-testid={`column-locked-marker-${column.key}`}
                      style={{ marginLeft: 6, color: token.colorTextTertiary }}
                    >
                      <LockOutlined style={{ fontSize: 11, marginRight: 4 }} />
                      (locked)
                    </span>
                  ) : null}
                </span>
              </Checkbox>
            </li>
          );
        })}
      </ul>
    </div>
  );

  return (
    <Popover
      content={content}
      trigger="click"
      placement="bottomRight"
      arrow={false}
      destroyOnHidden={true}
    >
      <Tooltip title="Column settings">
        <Button
          type="text"
          icon={<SettingOutlined />}
          aria-label="Column settings"
          data-testid="column-manager-trigger"
        />
      </Tooltip>
    </Popover>
  );
};
