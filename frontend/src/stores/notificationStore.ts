import { create } from "zustand";
import { persist } from "zustand/middleware";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type NotificationType = "success" | "error" | "warning" | "info";

export type NotificationCategory =
  | "sync"
  | "order"
  | "product"
  | "inventory"
  | "auth"
  | "system"
  | "export";

export interface NotificationItem {
  id: string;
  type: NotificationType;
  category: NotificationCategory;
  title: string;
  message: string;
  timestamp: number;
  read: boolean;
  actionUrl?: string;
}

// ---------------------------------------------------------------------------
// Store
// ---------------------------------------------------------------------------

const MAX_ITEMS = 100;
const MAX_AGE_MS = 24 * 60 * 60 * 1000; // 24 hours

interface NotificationState {
  notifications: NotificationItem[];
  isDropdownOpen: boolean;
}

interface NotificationActions {
  addNotification: (
    item: Omit<NotificationItem, "id" | "timestamp" | "read">,
  ) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  removeNotification: (id: string) => void;
  clearAll: () => void;
  toggleDropdown: () => void;
  closeDropdown: () => void;
}

export type NotificationStore = NotificationState & NotificationActions;

function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

function pruneOld(items: NotificationItem[]): NotificationItem[] {
  const cutoff = Date.now() - MAX_AGE_MS;
  return items.filter((n) => n.timestamp > cutoff).slice(0, MAX_ITEMS);
}

export const useNotificationStore = create<NotificationStore>()(
  persist(
    (set) => ({
      // State
      notifications: [],
      isDropdownOpen: false,

      // Actions
      addNotification: (item) =>
        set((state) => {
          const newItem: NotificationItem = {
            ...item,
            id: generateId(),
            timestamp: Date.now(),
            read: false,
          };
          const updated = [newItem, ...state.notifications];
          return { notifications: pruneOld(updated) };
        }),

      markAsRead: (id) =>
        set((state) => ({
          notifications: state.notifications.map((n) =>
            n.id === id ? { ...n, read: true } : n,
          ),
        })),

      markAllAsRead: () =>
        set((state) => ({
          notifications: state.notifications.map((n) => ({ ...n, read: true })),
        })),

      removeNotification: (id) =>
        set((state) => ({
          notifications: state.notifications.filter((n) => n.id !== id),
        })),

      clearAll: () => set({ notifications: [] }),

      toggleDropdown: () =>
        set((state) => ({ isDropdownOpen: !state.isDropdownOpen })),

      closeDropdown: () => set({ isDropdownOpen: false }),
    }),
    {
      name: "omni-notifications",
      // Only persist the notification items, not UI state
      partialize: (state) => ({
        notifications: state.notifications,
      }),
    },
  ),
);

// ---------------------------------------------------------------------------
// Selectors (avoid unnecessary re-renders)
// ---------------------------------------------------------------------------

export const selectUnreadCount = (state: NotificationStore) =>
  state.notifications.filter((n) => !n.read).length;

