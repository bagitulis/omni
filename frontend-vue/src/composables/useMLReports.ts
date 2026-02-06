/**
 * ML Reports Composable
 * Handles API calls for ML-generated reports
 * Backend integration: /api/ml/reports/*
 */

import { ref } from "vue";
import { useApi } from "./useApi";

const api = useApi();

export interface MLReport {
  id: number;
  tenant_id: string;
  platform: "shopee" | "tiktok";
  report_type: "full" | "executive" | "quick";
  period_start: string;
  period_end: string;
  period_label: string;
  file_path: string;
  file_name: string;
  file_size: number;
  status: "pending" | "completed" | "failed";
  error_msg?: string;
  created_at: string;
  updated_at: string;
}

export interface MLJob {
  id: number;
  tenant_id: string;
  job_type: string;
  platform: "shopee" | "tiktok";
  status: "pending" | "running" | "completed" | "failed";
  progress: number;
  result_id?: number;
  error_msg?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
}

export interface GenerateReportRequest {
  platform: "shopee" | "tiktok";
  report_type?: "full" | "executive" | "quick";
  period_label?: string;
}

export function useMLReports() {
  const loading = ref(false);
  const generating = ref(false);
  const error = ref<string | null>(null);
  const reports = ref<MLReport[]>([]);
  const latestReport = ref<MLReport | null>(null);
  const reportHTML = ref<string>("");

  /**
   * Generate new ML report
   * POST /api/ml/reports/generate
   */
  async function generateReport(
    request: GenerateReportRequest,
  ): Promise<MLJob> {
    generating.value = true;
    error.value = null;

    try {
      const response = await api.post("/ml/reports/generate", request);
      return response.data.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      throw err;
    } finally {
      generating.value = false;
    }
  }

  /**
   * Get list of reports for platform
   * GET /api/ml/reports/:platform/list
   */
  async function fetchReports(
    platform: "shopee" | "tiktok",
    page: number = 1,
    limit: number = 20,
  ): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      const response = await api.get(`/ml/reports/${platform}/list`, {
        params: { page, limit },
      });
      reports.value = response.data.data.reports;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * Get latest report for platform
   * GET /api/ml/reports/:platform/latest
   */
  async function fetchLatestReport(
    platform: "shopee" | "tiktok",
  ): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      const response = await api.get(`/ml/reports/${platform}/latest`);
      latestReport.value = response.data.data;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      latestReport.value = null;
    } finally {
      loading.value = false;
    }
  }

  /**
   * Get specific report by filename
   * GET /api/ml/reports/:platform/:filename
   */
  async function fetchReport(
    platform: "shopee" | "tiktok",
    filename: string,
  ): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      const response = await api.get(`/ml/reports/${platform}/${filename}`);
      reportHTML.value = response.data.data.html;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      reportHTML.value = "";
    } finally {
      loading.value = false;
    }
  }

  /**
   * Download report HTML
   */
  function downloadReport(report: MLReport): void {
    const blob = new Blob([reportHTML.value], { type: "text/html" });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = report.file_name;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  }

  /**
   * Format file size
   */
  function formatFileSize(bytes: number): string {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
  }

  /**
   * Format date
   */
  function formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString("id-ID", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  return {
    // State
    loading,
    generating,
    error,
    reports,
    latestReport,
    reportHTML,

    // Methods
    generateReport,
    fetchReports,
    fetchLatestReport,
    fetchReport,
    downloadReport,

    // Formatters
    formatFileSize,
    formatDate,
  };
}
