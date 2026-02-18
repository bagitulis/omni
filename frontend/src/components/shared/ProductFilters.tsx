import React, { useCallback, useRef, useState, useEffect } from "react";
import { Input, Select, Button, theme } from "antd";
import {
  SearchOutlined,
  FilterOutlined,
  ClearOutlined,
} from "@ant-design/icons";
import type { ProductFilterValues } from "@/types/shared";

interface ProductFiltersProps {
  values: ProductFilterValues;
  onChange: (values: ProductFilterValues) => void;
  showCategory?: boolean; // Products page has category, Inventory doesn't
}

const PLATFORM_OPTIONS = [
  { value: "all", label: "All Platforms" },
  { value: "shopee", label: "🟠 Shopee" },
  { value: "tiktok", label: "⬛ TikTok" },
  { value: "lazada", label: "🔵 Lazada" },
];

const STATUS_OPTIONS = [
  { value: "all", label: "All Status" },
  { value: "active", label: "Active" },
  { value: "draft", label: "Draft" },
  { value: "archived", label: "Archived" },
];

export const ProductFilters: React.FC<ProductFiltersProps> = ({
  values,
  onChange,
  showCategory = true,
}) => {
  const { token } = theme.useToken();
  const searchTimeout = useRef<ReturnType<typeof setTimeout>>(null);
  const [searchInput, setSearchInput] = useState(values.search || "");

  // Sync internal state if values.search changes externally (e.g. clear filters)
  useEffect(() => {
    setSearchInput(values.search || "");
  }, [values.search]);

  const handleSearchChange = useCallback(
    (value: string) => {
      setSearchInput(value);
      if (searchTimeout.current) clearTimeout(searchTimeout.current);
      searchTimeout.current = setTimeout(() => {
        onChange({ ...values, search: value });
      }, 500);
    },
    [values, onChange],
  );

  const handleFilterChange = useCallback(
    (key: keyof ProductFilterValues, value: string) => {
      onChange({ ...values, [key]: value });
    },
    [values, onChange],
  );

  const handleClear = useCallback(() => {
    setSearchInput("");
    onChange({
      search: "",
      platform: "all",
      status: "all",
      category: "all",
    });
  }, [onChange]);

  const hasActiveFilters =
    (values.search && values.search !== "") ||
    values.platform !== "all" ||
    values.status !== "all" ||
    (showCategory && values.category !== "all");

  return (
    <div
      style={{
        display: "flex",
        flexWrap: "wrap",
        gap: 12,
        alignItems: "center",
        padding: "12px 0",
      }}
      data-testid="product-filters"
    >
      {/* Search */}
      <Input
        placeholder="Search products..."
        prefix={<SearchOutlined style={{ color: token.colorTextQuaternary }} />}
        value={searchInput}
        onChange={(e) => handleSearchChange(e.target.value)}
        allowClear
        style={{ width: 280, maxWidth: "100%" }}
        data-testid="search-input"
      />

      {/* Platform filter */}
      <Select
        value={values.platform || "all"}
        onChange={(v) => handleFilterChange("platform", v)}
        options={PLATFORM_OPTIONS}
        style={{ width: 160 }}
        suffixIcon={<FilterOutlined />}
        data-testid="platform-select"
      />

      {/* Status filter */}
      <Select
        value={values.status || "all"}
        onChange={(v) => handleFilterChange("status", v)}
        options={STATUS_OPTIONS}
        style={{ width: 140 }}
        data-testid="status-select"
      />

      {/* Category filter (optional) */}
      {showCategory && (
        <Select
          value={values.category || "all"}
          onChange={(v) => handleFilterChange("category", v)}
          placeholder="Category"
          style={{ width: 160 }}
          options={[
            { value: "all", label: "All Categories" },
            // Categories would be loaded dynamically from API in a real app
            // For now, we just show "All Categories" as placeholder per plan
          ]}
          data-testid="category-select"
        />
      )}

      {/* Clear all filters */}
      {hasActiveFilters && (
        <Button
          type="text"
          icon={<ClearOutlined />}
          onClick={handleClear}
          style={{ color: token.colorTextSecondary }}
          data-testid="clear-filters-btn"
        >
          Clear filters
        </Button>
      )}
    </div>
  );
};
