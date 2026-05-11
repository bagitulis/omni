import React, { Suspense, useEffect } from "react";
import { ConfigProvider, App as AntApp } from "antd";
import {
  BrowserRouter,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "@/api/queryClient";
import apiClient from "@/api/client";
import { antdTheme, antdDarkTheme } from "./styles/theme";
import { AppLayout } from "./components/layout";
import { ProtectedRoute } from "./components/auth/ProtectedRoute";
import { ErrorBoundary } from "./components/ErrorBoundary";
import { AntStaticHolder } from "./components/AntStaticHolder";
import { useAuthStore } from "@/stores/authStore";
import { logger } from "@/lib/logger";
import LoginPage from "./pages/auth/LoginPage";
import "./styles/global.css";
import { ThemeProvider } from "./contexts/ThemeContext";
import { useTheme } from "./contexts/ThemeContext.hooks";
import { PageLoading } from "@/components/common/PageLoading";
import { NotificationProvider } from "./contexts/NotificationContext";

// Lazy-loaded pages for code splitting
const DashboardPage = React.lazy(
  () => import("./pages/dashboard/DashboardPage"),
);
const OrdersPage = React.lazy(() => import("./pages/orders/OrdersPage"));
const UnifiedProductsPage = React.lazy(
  () => import("./pages/products/UnifiedProductsPage"),
);
const MarketplaceSyncHistoryPage = React.lazy(
  () => import("./pages/products/MarketplaceSyncHistoryPage"),
);
const ProductAddPage = React.lazy(
  () => import("./pages/products/ProductAddPage"),
);
const ProductEditPage = React.lazy(
  () => import("./pages/products/ProductEditPage"),
);
const ProductImportPage = React.lazy(
  () => import("./pages/products/ProductImportPage"),
);
const SimplifiedInventoryPage = React.lazy(
  () => import("./pages/inventory/SimplifiedInventoryPage"),
);
const ShopeeReportPage = React.lazy(
  () => import("./pages/report/ShopeeReportPage"),
);
const TiktokReportPage = React.lazy(
  () => import("./pages/report/TiktokReportPage"),
);

const SettingsPage = React.lazy(() => import("./pages/settings/SettingsPage"));
const ScriptMonitorPage = React.lazy(
  () => import("./pages/script-monitor/ScriptMonitorPage"),
);
const RouteMappingPage = React.lazy(
  () => import("./pages/route-mapping/RouteMappingPage"),
);
const NotificationsPage = React.lazy(
  () => import("./pages/notifications/NotificationsPage"),
);
const UserManagementPage = React.lazy(
  () => import("./pages/users/UserManagementPage"),
);


function AppContent() {
  const { isDark } = useTheme();
  const initializeAuth = useAuthStore((state) => state.initializeAuth);

  useEffect(() => {
    initializeAuth();
  }, [initializeAuth]);

  useEffect(() => {
    // Fetch CSRF token on mount - sets cookie for API interceptor
    apiClient.get("/csrf-token").catch(() => {
      // Silent fail - CSRF will be retried on next mutation
      logger.debug(
        "[App] CSRF token fetch failed - will retry on next request",
      );
    });
  }, []);

  return (
    <ConfigProvider theme={isDark ? antdDarkTheme : antdTheme}>
      <AntApp>
        <NotificationProvider>
          <AntStaticHolder />
          <ErrorBoundary>
            <BrowserRouter>
              <Suspense fallback={<PageLoading />}>
                <Routes>
                  <Route path="/login" element={<LoginPage />} />
                  <Route
                    element={
                      <ProtectedRoute>
                        <AppLayout />
                      </ProtectedRoute>
                    }
                  >
                    <Route path="/" element={<DashboardPage />} />
                    <Route path="/order-manager" element={<OrdersPage />} />
                    <Route
                      path="/order-manager/:platform"
                      element={<OrdersPage />}
                    />
                    <Route path="/products" element={<UnifiedProductsPage />} />
                    <Route path="/products/add" element={<ProductAddPage />} />
                    <Route
                      path="/products/import"
                      element={<ProductImportPage />}
                    />
                    <Route
                      path="/products/sync-history"
                      element={<MarketplaceSyncHistoryPage />}
                    />
                    <Route
                      path="/products/:id/edit"
                      element={<ProductEditPage />}
                    />
                    <Route path="/route-mapping" element={<RouteMappingPage />} />
                    <Route
                      path="/inventory"
                      element={<SimplifiedInventoryPage />}
                    />
                    <Route
                      path="/report/shopee"
                      element={<ShopeeReportPage />}
                    />
                    <Route
                      path="/report/tiktok"
                      element={<TiktokReportPage />}
                    />
                    <Route path="/settings" element={<SettingsPage />} />
                    <Route path="/notifications" element={<NotificationsPage />} />
                    <Route
                      path="/script-monitor"
                      element={<ScriptMonitorPage />}
                    />
                    <Route
                      path="/users"
                      element={
                        <ProtectedRoute permission="users.list">
                          <UserManagementPage />
                        </ProtectedRoute>
                      }
                    />
                    <Route path="*" element={<Navigate to="/" replace />} />
                  </Route>
                </Routes>
              </Suspense>
            </BrowserRouter>
          </ErrorBoundary>
        </NotificationProvider>
      </AntApp>
    </ConfigProvider>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <AppContent />
      </ThemeProvider>
    </QueryClientProvider>
  );
}
