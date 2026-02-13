export interface ProgressState {
  current: number;
  total: number;
  status: "idle" | "processing" | "done";
}

export interface BulkActionFailure {
  order_sn: string;
  error: string;
}

export interface BulkResult {
  succeeded: string[];
  failed: BulkActionFailure[];
}
