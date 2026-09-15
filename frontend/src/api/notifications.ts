import { apiClient as client } from './client';

export type NotificationSeverity = 10 | 20 | 30 | 40 | 50;

export const SEVERITY = {
  info: 10,
  low: 20,
  medium: 30,
  high: 40,
  critical: 50,
} as const;

export interface Notification {
  id: number;
  type: 'success' | 'error' | 'warning' | 'info';
  category: string;
  severity: NotificationSeverity;
  title: string;
  message: string;
  metadata?: string;
  read: boolean;
  action_url?: string;
  dedup_key?: string | null;
  dedup_count: number;
  source?: string;
  snoozed_until?: string | null;
  expires_at?: string | null;
  created_at: string;
  updated_at?: string;
}

export interface NotificationCounts {
  total: number;
  unread: number;
  by_severity: Record<string, number>;
}

export interface NotificationSettings {
  retention_days: number;
  min_severity_toast?: number;
}

export interface ListParams {
  limit?: number;
  since_id?: number;
  unread_only?: boolean;
  category?: string;
  min_severity?: number;
  q?: string;
  from?: string;
  to?: string;
}

export const notificationApi = {
  list: (params: ListParams = {}) =>
    client.get<{ items: Notification[]; count: number }>('/notifications', { params }),

  create: (data: { type: string; category: string; title: string; message?: string; action_url?: string }) =>
    client.post<Notification>('/notifications', data),

  getCounts: () =>
    client.get<NotificationCounts>('/notifications/counts'),

  // Legacy alias used by pre-V2 clients; kept so the bell can degrade gracefully.
  getUnreadCount: () =>
    client.get<{ unread_count: number }>('/notifications/unread-count'),

  markAsRead: (id: number) =>
    client.patch<{ affected: number }>(`/notifications/${id}/read`),

  markAllAsRead: () =>
    client.patch<{ affected: number }>('/notifications/read-all'),

  bulkMarkRead: (ids: number[]) =>
    client.post<{ affected: number }>('/notifications/bulk/read', { ids }),

  bulkDelete: (ids: number[]) =>
    client.post<{ affected: number }>('/notifications/bulk/delete', { ids }),

  snooze: (id: number, until: string) =>
    client.post<{ snoozed_until: string }>(`/notifications/${id}/snooze`, { until }),

  delete: (id: number) =>
    client.delete<{ affected: number }>(`/notifications/${id}`),

  deleteAll: () =>
    client.delete('/notifications'),

  getSettings: () =>
    client.get<NotificationSettings>('/notifications/settings'),

  updateSettings: (settings: NotificationSettings) =>
    client.put('/notifications/settings', settings),
};
