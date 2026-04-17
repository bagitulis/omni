import React, { createContext, useContext, useEffect, useState, useCallback, useRef } from 'react';
import { notificationApi, Notification } from '@/api/notifications';
import { useAuthStore } from '@/stores/authStore';
import { API_BASE_URL } from '@/lib/constants';
import { App } from 'antd';

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

export const NotificationProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const { tenantId, getValidToken } = useAuthStore();
  const { notification } = App.useApp();
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const MAX_RECONNECT_ATTEMPTS = 10;

  const fetchUnreadCount = useCallback(async () => {
    try {
      const res = await notificationApi.getUnreadCount();
      if (res.success && res.data) {
        setUnreadCount(res.data.unread_count);
      }
    } catch (err) {
      console.error('Failed to fetch unread count', err);
    }
  }, []);

  const fetchNotifications = useCallback(async () => {
    setLoading(true);
    try {
      const res = await notificationApi.list({ limit: 50 });
      if (res.success && res.data) {
        setNotifications(res.data.items);
      }
    } catch (err) {
      console.error('Failed to fetch notifications', err);
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
      console.error('Failed to mark as read', err);
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
      console.error('Failed to mark all as read', err);
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
      console.error('Failed to delete notification', err);
    }
  };

  const setupSSE = useCallback(async () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const token = await getValidToken();
    if (!token || !tenantId) return;

    // Use token as query param for SSE since headers aren't supported
    const sseUrl = `${API_BASE_URL}/notifications/stream?token=${token}`;
    const es = new EventSource(sseUrl);

    es.onopen = () => {
      // Reset reconnect counter on successful connection
      reconnectAttemptsRef.current = 0;
    };

    es.addEventListener('notification', (event: MessageEvent) => {
      try {
        const newNotif: Notification = JSON.parse(event.data);
        
        // Add to list if not already present
        setNotifications(prev => {
          if (prev.find(n => n.id === newNotif.id)) return prev;
          return [newNotif, ...prev].slice(0, 100);
        });
        
        // Increment unread count
        setUnreadCount(prev => prev + 1);

        // Show browser-style toast
        notification[newNotif.type]({
          message: newNotif.title,
          description: newNotif.message,
          placement: 'topRight',
          duration: 5,
        });
      } catch (err) {
        console.error('Failed to parse SSE notification', err);
      }
    });

    es.onerror = () => {
      es.close();
      const attempts = reconnectAttemptsRef.current;
      if (attempts >= MAX_RECONNECT_ATTEMPTS) {
        console.warn('SSE: max reconnection attempts reached, falling back to polling');
        return;
      }
      // Exponential backoff: 5s, 10s, 20s, 40s, ... max 60s
      const delay = Math.min(5000 * Math.pow(2, attempts), 60000);
      reconnectAttemptsRef.current = attempts + 1;
      console.warn(`SSE: reconnecting in ${delay / 1000}s (attempt ${attempts + 1}/${MAX_RECONNECT_ATTEMPTS})`);
      setTimeout(setupSSE, delay);
    };

    eventSourceRef.current = es;
  }, [tenantId, getValidToken, notification]);

  useEffect(() => {
    if (tenantId) {
      fetchUnreadCount();
      fetchNotifications();
      setupSSE();
    }
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
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
