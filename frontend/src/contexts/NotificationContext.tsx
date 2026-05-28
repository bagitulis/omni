/* eslint-disable react-refresh/only-export-components */
import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { App } from 'antd';
import { getSSETicket } from '@/api/auth';
import { notificationApi } from '@/api/notifications';
import type { Notification } from '@/api/notifications';
import { API_BASE_URL } from '@/lib/constants';
import { useAuthStore } from '@/stores/authStore';

interface NotificationContextType {
  notifications: Notification[];
  unreadCount: number;
  loading: boolean;
  markAsRead: (id: number) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  deleteNotification: (id: number) => Promise<void>;
  fetchNotifications: () => Promise<void>;
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

export const NotificationProvider = ({ children }: { children: ReactNode }) => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const { tenantId, getValidToken } = useAuthStore();
  const { notification } = App.useApp();
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const setupSSERef = useRef<() => void>(() => {});
  const MAX_RECONNECT_ATTEMPTS = 10;

  const fetchUnreadCount = useCallback(async () => {
    try {
      const res = await notificationApi.getUnreadCount();
      if (res.success && res.data) {
        setUnreadCount(res.data.unread_count);
      }
    } catch (err) {
      logger.warn('Failed to fetch unread count', { err: err });
    }
  }, []);

  const fetchNotifications = useCallback(async () => {
    setLoading(true);
    try {
      const res = await notificationApi.list({ limit: 50 });
      if (res.success && res.data) {
        const items = res.data.items;
        setNotifications(prev => {
          const map = new Map(prev.map(n => [n.id, n]));
          items.forEach(n => { map.set(n.id, n); });
          return Array.from(map.values());
        });
      }
    } catch (err) {
      logger.warn('Failed to fetch notifications', { err: err });
    } finally {
      setLoading(false);
    }
  }, []);

  const markAsRead = async (id: number) => {
    try {
      const res = await notificationApi.markAsRead(id);
      if (res.success) {
        setNotifications(prev =>
          prev.map(n => (n.id === id ? { ...n, read: true } : n))
        );
        setUnreadCount(prev => Math.max(0, prev - 1));
      }
    } catch (err) {
      logger.warn('Failed to mark as read', { err: err });
    }
  };

  const markAllAsRead = async () => {
    try {
      const res = await notificationApi.markAllAsRead();
      if (res.success) {
        setNotifications(prev => prev.map(n => ({ ...n, read: true })));
        setUnreadCount(0);
      }
    } catch (err) {
      logger.warn('Failed to mark all as read', { err: err });
    }
  };

  const deleteNotification = async (id: number) => {
    try {
      const res = await notificationApi.delete(id);
      if (res.success) {
        const deletedWasUnread = notifications.find(n => n.id === id)?.read === false;
        setNotifications(prev => prev.filter(n => n.id !== id));
        if (deletedWasUnread) {
          setUnreadCount(prev => Math.max(0, prev - 1));
        }
      }
    } catch (err) {
      logger.warn('Failed to delete notification', { err: err });
    }
  };

  const setupSSEListeners = useCallback((es: EventSource) => {
    es.onopen = () => {
      reconnectAttemptsRef.current = 0;
    };

    es.addEventListener('notification', (event: MessageEvent) => {
      try {
        const newNotif: Notification = JSON.parse(event.data);
        setNotifications(prev => {
          if (prev.find(n => n.id === newNotif.id)) return prev;
          return [newNotif, ...prev].slice(0, 100);
        });
        setUnreadCount(prev => prev + 1);
        const notify = notification[newNotif.type] ?? notification.info;
        notify({
          message: newNotif.title,
          description: newNotif.message,
          placement: 'topRight',
          duration: 5,
        });
      } catch (err) {
        logger.error('Failed to parse SSE notification', { err: err });
      }
    });

    es.onerror = () => {
      es.close();
      const attempts = reconnectAttemptsRef.current;
      if (attempts >= MAX_RECONNECT_ATTEMPTS) {
        logger.warn('SSE: max reconnection attempts reached, falling back to polling');
        const pollInterval = setInterval(() => {
          fetchNotifications();
          fetchUnreadCount();
        }, 60000);
        pollIntervalRef.current = pollInterval;
        return;
      }
      const delay = Math.min(5000 * (2 ** attempts), 60000);
      reconnectAttemptsRef.current = attempts + 1;
      logger.warn(`SSE: reconnecting in ${delay / 1000}s (attempt ${attempts + 1}/${MAX_RECONNECT_ATTEMPTS})`);
      reconnectTimerRef.current = setTimeout(() => setupSSERef.current(), delay);
    };

    eventSourceRef.current = es;
  }, [notification, fetchNotifications, fetchUnreadCount]);

  const setupSSE = useCallback(async () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const token = await getValidToken();
    if (!token || !tenantId) return;

    // Exchange JWT for a short-lived one-time ticket (prevents JWT exposure in URL)
    const ticket = await getSSETicket();
    if (!ticket) {
      logger.warn('[SSE] Failed to get ticket, SSE unavailable');
      return;
    }

    const sseUrl = `${API_BASE_URL}/notifications/stream?ticket=${ticket}`;
    const es = new EventSource(sseUrl);
    setupSSEListeners(es);
  }, [tenantId, getValidToken, setupSSEListeners]);

  useEffect(() => {
    setupSSERef.current = setupSSE;
  }, [setupSSE]);

  useEffect(() => {
    if (tenantId) {
      fetchUnreadCount();
      fetchNotifications();
      setupSSE();

      // Periodic refresh every 5 minutes
      const periodicRefresh = setInterval(() => {
        fetchUnreadCount();
      }, 5 * 60 * 1000);

      // Refresh on tab focus
      const handleVisibilityChange = () => {
        if (document.visibilityState === 'visible') {
          fetchUnreadCount();
          fetchNotifications();
        }
      };
      document.addEventListener('visibilitychange', handleVisibilityChange);

      return () => {
        if (eventSourceRef.current) {
          eventSourceRef.current.close();
        }
        if (pollIntervalRef.current) {
          clearInterval(pollIntervalRef.current);
        }
        if (reconnectTimerRef.current) {
          clearTimeout(reconnectTimerRef.current);
        }
        clearInterval(periodicRefresh);
        document.removeEventListener('visibilitychange', handleVisibilityChange);
      };
    }
  }, [tenantId, fetchUnreadCount, fetchNotifications, setupSSE]);

  return (
    <NotificationContext.Provider
      value={{
        notifications,
        unreadCount,
        loading,
        markAsRead,
        markAllAsRead,
        deleteNotification,
        fetchNotifications,
      }}
    >
      {children}
    </NotificationContext.Provider>
  );
};

export const useNotifications = () => {
  const context = useContext(NotificationContext);
  if (context === undefined) {
    throw new Error('useNotifications must be used within a NotificationProvider');
  }
  return context;
};
