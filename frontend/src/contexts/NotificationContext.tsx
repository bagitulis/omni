/* eslint-disable react-refresh/only-export-components */
import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { App } from 'antd';
import {
  notificationApi,
  type ListParams,
  type Notification,
  type NotificationCounts,
} from '@/api/notifications';
import { useAuthStore } from '@/stores/authStore';
import { useRealtime } from '@/hooks/useRealtime';
import { sanitizeForUser } from '@/lib/notificationSecurity';
import { logger } from '@/lib/logger';

interface NotificationContextType {
  notifications: Notification[];
  counts: NotificationCounts;
  unreadCount: number;
  loading: boolean;
  fetchNotifications: (params?: ListParams) => Promise<void>;
  markAsRead: (id: number) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  bulkMarkRead: (ids: number[]) => Promise<void>;
  bulkDelete: (ids: number[]) => Promise<void>;
  snooze: (id: number, until: Date) => Promise<void>;
  deleteNotification: (id: number) => Promise<void>;
}

const DEFAULT_COUNTS: NotificationCounts = { total: 0, unread: 0, by_severity: {} };

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

export const NotificationProvider = ({ children }: { children: ReactNode }) => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [counts, setCounts] = useState<NotificationCounts>(DEFAULT_COUNTS);
  const [loading, setLoading] = useState(false);
  const { tenantId } = useAuthStore();
  const { notification } = App.useApp();
  const notificationRef = useRef(notification);

  useEffect(() => {
    notificationRef.current = notification;
  }, [notification]);

  const mergeNotifications = useCallback((incoming: Notification[]) => {
    setNotifications((prev) => {
      const byId = new Map<number, Notification>();
      [...prev, ...incoming].forEach((item) => {
        const existing = byId.get(item.id);
        byId.set(item.id, existing ? { ...existing, ...item } : item);
      });
      return Array.from(byId.values()).sort(
        (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
      );
    });
  }, []);

  const fetchCounts = useCallback(async () => {
    try {
      const res = await notificationApi.getCounts();
      if (res.success && res.data) setCounts(res.data);
    } catch (err) {
      logger.warn('Failed to fetch notification counts', { err });
    }
  }, []);

  const fetchNotifications = useCallback(async (params: ListParams = {}) => {
    setLoading(true);
    try {
      const res = await notificationApi.list({ limit: 50, ...params });
      if (res.success && res.data) mergeNotifications(res.data.items);
    } catch (err) {
      logger.warn('Failed to fetch notifications', { err });
    } finally {
      setLoading(false);
    }
  }, [mergeNotifications]);

  const markAsRead = useCallback(async (id: number) => {
    try {
      const res = await notificationApi.markAsRead(id);
      if (res.success) {
        setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, read: true } : n)));
        setCounts((prev) => ({ ...prev, unread: Math.max(0, prev.unread - 1) }));
      }
    } catch (err) {
      logger.warn('Failed to mark as read', { err });
    }
  }, []);

  const markAllAsRead = useCallback(async () => {
    try {
      const res = await notificationApi.markAllAsRead();
      if (res.success) {
        setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
        setCounts((prev) => ({ ...prev, unread: 0 }));
      }
    } catch (err) {
      logger.warn('Failed to mark all as read', { err });
    }
  }, []);

  const bulkMarkRead = useCallback(async (ids: number[]) => {
    if (ids.length === 0) return;
    try {
      const res = await notificationApi.bulkMarkRead(ids);
      if (res.success) {
        const idSet = new Set(ids);
        setNotifications((prev) => prev.map((n) => (idSet.has(n.id) ? { ...n, read: true } : n)));
        await fetchCounts();
      }
    } catch (err) {
      logger.warn('Failed to bulk mark read', { err });
    }
  }, [fetchCounts]);

  const bulkDelete = useCallback(async (ids: number[]) => {
    if (ids.length === 0) return;
    try {
      const res = await notificationApi.bulkDelete(ids);
      if (res.success) {
        const idSet = new Set(ids);
        setNotifications((prev) => prev.filter((n) => !idSet.has(n.id)));
        await fetchCounts();
      }
    } catch (err) {
      logger.warn('Failed to bulk delete', { err });
    }
  }, [fetchCounts]);

  const snooze = useCallback(async (id: number, until: Date) => {
    try {
      const res = await notificationApi.snooze(id, until.toISOString());
      if (res.success) {
        // Snoozed items are hidden from the unread view; remove from local list.
        setNotifications((prev) => prev.filter((n) => n.id !== id));
        await fetchCounts();
      }
    } catch (err) {
      logger.warn('Failed to snooze', { err });
    }
  }, [fetchCounts]);

  const deleteNotification = useCallback(async (id: number) => {
    try {
      const res = await notificationApi.delete(id);
      if (res.success) {
        setNotifications((prev) => {
          const wasUnread = prev.find((n) => n.id === id)?.read === false;
          if (wasUnread) setCounts((c) => ({ ...c, unread: Math.max(0, c.unread - 1) }));
          return prev.filter((n) => n.id !== id);
        });
      }
    } catch (err) {
      logger.warn('Failed to delete notification', { err });
    }
  }, []);

  // Realtime: listen on the shared /api/realtime/ws for TopicNotifications.
  useRealtime('notifications/updated', (env) => {
    try {
      const payload = env.payload as Partial<Notification>;
      if (!payload || typeof payload.id !== 'number') return;
      const safeMsg = sanitizeForUser(payload.message ?? '');
      const safeTitle = sanitizeForUser(payload.title ?? '');
      const merged: Notification = {
        ...(payload as Notification),
        title: safeTitle,
        message: safeMsg,
        read: false,
      };
      mergeNotifications([merged]);
      setCounts((prev) => ({ ...prev, unread: prev.unread + 1 }));
      const notifier = notificationRef.current;
      const notify = notifier[payload.type ?? 'info'] ?? notifier.info;
      notify({
        message: safeTitle,
        description: safeMsg,
        placement: 'topRight',
        duration: 5,
      });
    } catch (err) {
      logger.error('Failed to handle realtime notification', { err });
    }
  }, Boolean(tenantId));

  // Bootstrap: initial fetch + visibility refresh + periodic refresh.
  useEffect(() => {
    if (!tenantId) {
      setNotifications([]);
      setCounts(DEFAULT_COUNTS);
      return;
    }
    void fetchCounts();
    void fetchNotifications();

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        void fetchCounts();
        void fetchNotifications();
      }
    };
    document.addEventListener('visibilitychange', handleVisibilityChange);

    const periodic = setInterval(() => {
      void fetchCounts();
    }, 5 * 60 * 1000);

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      clearInterval(periodic);
    };
  }, [tenantId, fetchCounts, fetchNotifications]);

  const value = useMemo<NotificationContextType>(
    () => ({
      notifications,
      counts,
      unreadCount: counts.unread,
      loading,
      fetchNotifications,
      markAsRead,
      markAllAsRead,
      bulkMarkRead,
      bulkDelete,
      snooze,
      deleteNotification,
    }),
    [notifications, counts, loading, fetchNotifications, markAsRead, markAllAsRead, bulkMarkRead, bulkDelete, snooze, deleteNotification],
  );

  return <NotificationContext.Provider value={value}>{children}</NotificationContext.Provider>;
};

export const useNotifications = () => {
  const context = useContext(NotificationContext);
  if (context === undefined) {
    throw new Error('useNotifications must be used within a NotificationProvider');
  }
  return context;
};
