export interface Job {
  id: string;
  type: string;
  status: string;
  priority: string;
  data: Record<string, any>;
  started_at?: string;
  created_at: string;
}

export interface JobHistory {
  id: number;
  job_id: string;
  job_type?: string;
  status: string;
  duration_ms?: number;
  error_message?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
}

export interface AutoFunctionConfig {
  id: number;
  name: string;
  enabled: boolean;
  interval_minutes: number;
  start_time?: string;
  end_time?: string;
  last_executed?: string;
  next_scheduled_execution?: string;
}

export interface MonitorData {
  current_job: Job | null;
  pending_queue: Job[];
  recent_history: JobHistory[];
  total_pending: number;
  total_completed: number;
}
