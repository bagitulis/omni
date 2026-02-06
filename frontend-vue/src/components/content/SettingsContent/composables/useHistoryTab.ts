import { ref, computed, watch, type Ref, type ComputedRef } from "vue";
import apiClient from "@/services/api";

/**
 * Job History Types
 * API types use snake_case to match backend JSON response
 */
export interface JobHistory {
  id: number;
  job_id: string;
  job_type?: string;
  status: string;
  duration_ms?: number;
  error_message?: string;
  created_at: Date;
  started_at?: Date;
  completed_at?: Date;
}

export interface HistoryFilters {
  search: string;
  status: string;
  jobType: string;
  pageSize: number;
}

export interface ColumnDef {
  key: string;
  label: string;
  sortable: boolean;
}

export interface UseHistoryTabReturn {
  isLoading: Ref<boolean>;
  historyData: Ref<JobHistory[]>;
  jobTypes: Ref<string[]>;
  currentPage: Ref<number>;
  totalItems: Ref<number>;
  totalPages: Ref<number>;
  sortBy: Ref<string>;
  sortOrder: Ref<"asc" | "desc">;
  filters: Ref<HistoryFilters>;
  columns: ColumnDef[];
  hasActiveFilters: ComputedRef<boolean>;
  paginationStart: ComputedRef<number>;
  paginationEnd: ComputedRef<number>;
  visiblePages: ComputedRef<number[]>;
  debouncedFetch: () => void;
  fetchHistory: () => Promise<void>;
  fetchJobTypes: () => Promise<void>;
  toggleSort: (column: string) => void;
  goToPage: (page: number) => void;
}

export function useHistoryTab(activeTabRef: Ref<string>): UseHistoryTabReturn {
  const isLoading = ref(false);
  const historyData = ref<JobHistory[]>([]);
  const jobTypes = ref<string[]>([]);
  const currentPage = ref(1);
  const totalItems = ref(0);
  const totalPages = ref(0);
  const sortBy = ref("created_at");
  const sortOrder = ref<"asc" | "desc">("desc");

  const filters = ref<HistoryFilters>({
    search: "",
    status: "all",
    jobType: "all",
    pageSize: 20,
  });

  const columns: ColumnDef[] = [
    { key: "job_type", label: "Job Type", sortable: true },
    { key: "job_id", label: "Job ID", sortable: true },
    { key: "status", label: "Status", sortable: true },
    { key: "duration_ms", label: "Duration", sortable: true },
    { key: "completed_at", label: "Completed", sortable: true },
    { key: "error_message", label: "Error", sortable: false },
  ];

  const hasActiveFilters = computed(() => {
    return (
      filters.value.search !== "" ||
      filters.value.status !== "all" ||
      filters.value.jobType !== "all"
    );
  });

  const paginationStart = computed(() => {
    if (totalItems.value === 0) return 0;
    return (currentPage.value - 1) * filters.value.pageSize + 1;
  });

  const paginationEnd = computed(() => {
    return Math.min(
      currentPage.value * filters.value.pageSize,
      totalItems.value,
    );
  });

  const visiblePages = computed(() => {
    const pages: number[] = [];
    const maxVisible = 5;
    let start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2));
    let end = Math.min(totalPages.value, start + maxVisible - 1);

    if (end - start + 1 < maxVisible) {
      start = Math.max(1, end - maxVisible + 1);
    }

    for (let i = start; i <= end; i++) {
      pages.push(i);
    }
    return pages;
  });

  let debounceTimer: number | null = null;

  function debouncedFetch() {
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = window.setTimeout(() => {
      currentPage.value = 1;
      fetchHistory();
    }, 300);
  }

  async function fetchHistory() {
    isLoading.value = true;
    try {
      const params = new URLSearchParams({
        page: currentPage.value.toString(),
        pageSize: filters.value.pageSize.toString(),
        sortBy: sortBy.value,
        sortOrder: sortOrder.value,
      });

      if (filters.value.status !== "all") {
        params.append("status", filters.value.status);
      }
      if (filters.value.jobType !== "all") {
        params.append("jobType", filters.value.jobType);
      }
      if (filters.value.search) {
        params.append("search", filters.value.search);
      }

      const response = await apiClient.client.get(
        `/jobs/history-paginated?${params}`,
      );

      if (response.data.success) {
        historyData.value = response.data.data.data;
        totalItems.value = response.data.data.total;
        totalPages.value = response.data.data.totalPages;
        currentPage.value = response.data.data.page;
      }
    } catch (error) {
      console.error("Failed to fetch history:", error);
    } finally {
      isLoading.value = false;
    }
  }

  async function fetchJobTypes() {
    try {
      const response = await apiClient.client.get("/jobs/history-job-types");
      if (response.data.success) {
        jobTypes.value = response.data.data.jobTypes;
      }
    } catch (error) {
      console.error("Failed to fetch job types:", error);
    }
  }

  function toggleSort(column: string) {
    if (sortBy.value === column) {
      sortOrder.value = sortOrder.value === "asc" ? "desc" : "asc";
    } else {
      sortBy.value = column;
      sortOrder.value = "desc";
    }
    fetchHistory();
  }

  function goToPage(page: number) {
    if (page >= 1 && page <= totalPages.value) {
      currentPage.value = page;
      fetchHistory();
    }
  }

  watch(activeTabRef, (newTab) => {
    if (newTab === "history") {
      fetchHistory();
      fetchJobTypes();
    }
  });

  return {
    isLoading,
    historyData,
    jobTypes,
    currentPage,
    totalItems,
    totalPages,
    sortBy,
    sortOrder,
    filters,
    columns,
    hasActiveFilters,
    paginationStart,
    paginationEnd,
    visiblePages,
    debouncedFetch,
    fetchHistory,
    fetchJobTypes,
    toggleSort,
    goToPage,
  };
}
