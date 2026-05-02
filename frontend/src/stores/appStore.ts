import { create } from "zustand";

type ConnectionStatus = "connecting" | "connected" | "error" | "disconnected";

export type ActiveTab =
  | "product-management"
  | "order-management"
  | "inventory"
  | "settings"
  | "script-monitor"
  | "logs";

export type Platform = "shopee" | "lazada" | "tiktok";

interface AppState {
  // Connection
  connectionStatus: ConnectionStatus;
  isConnected: boolean; // derived: connectionStatus === 'connected'

  // UI State
  isLoading: boolean;
  sidebarCollapsed: boolean;
  activeTab: ActiveTab;
  activePlatform: Platform;

  // Actions
  setConnectionStatus: (status: ConnectionStatus) => void;
  setIsLoading: (loading: boolean) => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  toggleSidebar: () => void;
  setActiveTab: (tab: ActiveTab) => void;
  setActivePlatform: (platform: Platform) => void;
}

export const useAppStore = create<AppState>((set) => ({
  // Initial State
  connectionStatus: "connecting",
  isConnected: false,
  isLoading: false,
  sidebarCollapsed: false,
  activeTab: "order-management", // Default to order management as it's a common start
  activePlatform: "shopee", // Default platform

  // Actions
  setConnectionStatus: (status: ConnectionStatus) =>
    set({
      connectionStatus: status,
      isConnected: status === "connected",
    }),

  setIsLoading: (loading: boolean) => set({ isLoading: loading }),

  setSidebarCollapsed: (collapsed: boolean) =>
    set({ sidebarCollapsed: collapsed }),

  toggleSidebar: () =>
    set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),

  setActiveTab: (tab: ActiveTab) => set({ activeTab: tab }),

  setActivePlatform: (platform: Platform) => set({ activePlatform: platform }),
}));
