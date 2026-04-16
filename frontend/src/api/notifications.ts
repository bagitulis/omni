import { client } from './client';

export interface Notification {
  id: number;
  type: 'success' | 'error' | 'warning' | 'info';
  category: string;
  title: string;
  message: string;
  read: boolean;
  action_url?: string;
  created_at: string;
}

export interface NotificationSettings {
  retention_days: number;
}

export const notificationApi = {
  list: (params: { limit?: number; since_id?: number; unread_only?: boolean } = {}) =>
    client.get<{ items: Notification[]; count: number }>('/notifications', { params }),

  create: (data: { type: string; category: string; title: string; message?: string; action_url?: string }) =>
    client.post<Notification>('/notifications', data),

  getUnreadCount: () =>
    client.get<{ unread_count: number }>('/notifications/unread-count'),

  markAsRead: (id: number) =>
    client.patch(`/notifications/${id}/read`),

  markAllAsRead: () =>
    client.patch('/notifications/read-all'),

  delete: (id: number) =>
    client.delete(`/notifications/${id}`),

  deleteAll: () =>
    client.delete('/notifications'),

  getSettings: () =>
    client.get<NotificationSettings>('/notifications/settings'),

  updateSettings: (settings: NotificationSettings) =>
    client.put('/notifications/settings', settings),
};
