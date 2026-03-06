import { useMutation, useQuery } from "@tanstack/react-query";
import { message } from "@/components/AntStaticHolder";
import {
  getImportPreview,
  autoMapSkus,
  getMappingStatus,
  importProducts,
} from "@/api/products";
import type { ImportPreviewData, ImportRow } from "@/types/product";

const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

/**
 * Hook for product import preview
 * Uploads file and returns parsed preview data with validation
 */
export function useImportPreview() {
  return useMutation({
    mutationFn: async (file: File) => {
      // File size validation
      if (file.size > MAX_FILE_SIZE) {
        throw new Error("File size exceeds 10MB limit");
      }

      // File type validation
      const isCSV = file.type === "text/csv" || file.name.endsWith(".csv");
      const isExcel =
        file.type ===
          "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" ||
        file.type === "application/vnd.ms-excel" ||
        file.name.endsWith(".xlsx") ||
        file.name.endsWith(".xls");

      if (!isCSV && !isExcel) {
        throw new Error("Only CSV and Excel files are supported");
      }

      return getImportPreview(file);
    },
    onSuccess: (data: ImportPreviewData) => {
      message.success(
        `File parsed successfully: ${data.valid_rows} valid rows found`,
      );
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to preview import file");
    },
  });
}

/**
 * Hook for importing products
 * Executes the actual import with validated rows
 */
export function useImportProducts() {
  return useMutation({
    mutationFn: async (rows: ImportRow[]) => {
      const validRows = rows.filter((row) => row.valid);

      if (validRows.length === 0) {
        throw new Error("No valid rows to import");
      }

      return importProducts(validRows);
    },
    onSuccess: (data) => {
      message.success(`Successfully imported ${data.imported} products`);
    },
    onError: (error: Error) => {
      message.error(error.message || "Import failed");
    },
  });
}

/**
 * Hook for auto-mapping SKUs to platform products
 * Uses backend API to find and link matching products
 */
export function useAutoMapSkus() {
  return useMutation({
    mutationFn: async (skus: string[]) => {
      if (skus.length === 0) {
        throw new Error("No SKUs provided for auto-mapping");
      }

      return autoMapSkus(skus);
    },
    onSuccess: (result) => {
      if (result.mapped_count > 0) {
        message.success(
          `Auto-mapped ${result.mapped_count} SKUs to platform products`,
        );
      } else {
        message.warning("No matching platform products found for auto-mapping");
      }
    },
    onError: (error: Error) => {
      message.error(error.message || "Auto-mapping failed");
    },
  });
}

/**
 * Hook for fetching mapping status
 * Shows overview of mapped vs unmapped SKUs
 */
export function useMappingStatus() {
  return useQuery({
    queryKey: ["mapping-status"],
    queryFn: getMappingStatus,
    staleTime: 30_000, // 30 seconds
  });
}
