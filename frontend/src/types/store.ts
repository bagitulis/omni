// Store state types

export interface ModalStateType {
  visible: boolean;
  platform?: string | null;
  [key: string]: any;
}

export interface ModalsStateType {
  token: ModalStateType;
  wallet: ModalStateType;
  shipping: ModalStateType;
  price: ModalStateType;
  exportOrders: ModalStateType;
  debug: ModalStateType;
}

export interface AppStoreState {
  status: string | null;
  loading: boolean;
  logs: string;
  connectionStatus: "connecting" | "connected" | "error" | "disconnected";
  requestCount: number;
  errorCount: number;
  modals: ModalsStateType;
  shippingFiles: any[];
  debugInfo: any;
  spreadsheets: any[];
  lastStatusUpdate: string | null;
  isReconnecting: boolean;
  statusPollingInterval: any;
}

export interface GoogleSheetsStoreState {
  authStatus: {
    authenticated: boolean;
    has_credentials: boolean;
    expires_at?: string;
  };
  sheetsList: any[];
  selectedSheet?: string;
  syncStatus: "idle" | "syncing" | "success" | "error";
  lastSync?: string;
  syncMessage?: string;
}

export interface UIStoreState {
  sidebarOpen: boolean;
  theme: "light" | "dark";
  locale: string;
  [key: string]: any;
}
