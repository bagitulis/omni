/**
 * Inventory Pagination Composable
 * Handles pagination logic and page size persistence
 */

import { loadPageSize, savePageSize } from "./usePageSizePersistence";

interface PaginationState {
  currentOffset: number;
  pageSize: number;
  totalRecords: number;
}

export function useInventoryPagination(state: PaginationState) {
  /**
   * Initialize page size from persistence
   */
  function initializePageSize(): number {
    return loadPageSize();
  }

  /**
   * Go to previous page
   */
  function previousPage(): void {
    if (state.currentOffset > 0) {
      state.currentOffset = Math.max(0, state.currentOffset - state.pageSize);
    }
  }

  /**
   * Go to next page
   */
  function nextPage(): void {
    if (state.currentOffset + state.pageSize < state.totalRecords) {
      state.currentOffset += state.pageSize;
    }
  }

  /**
   * Handle page size change
   */
  function handlePageSizeChange(): void {
    savePageSize(state.pageSize);
    state.currentOffset = 0;
  }

  /**
   * Paginate a list based on current offset and page size
   */
  function paginateList<T>(list: T[]): T[] {
    const start = state.currentOffset;
    const end = state.currentOffset + state.pageSize;
    return list.slice(start, end);
  }

  /**
   * Get current page number (1-indexed)
   */
  function getCurrentPage(): number {
    return Math.floor(state.currentOffset / state.pageSize) + 1;
  }

  /**
   * Get total pages
   */
  function getTotalPages(): number {
    return Math.ceil(state.totalRecords / state.pageSize);
  }

  /**
   * Reset to first page
   */
  function resetToFirstPage(): void {
    state.currentOffset = 0;
  }

  return {
    initializePageSize,
    previousPage,
    nextPage,
    handlePageSizeChange,
    paginateList,
    getCurrentPage,
    getTotalPages,
    resetToFirstPage,
  };
}
