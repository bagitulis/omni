import React, { Suspense, useEffect } from "react";
import { ConfigProvider, App as AntApp } from "antd";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "@/api/queryClient";
import apiClient from "@/api/client";
import { antdTheme, antdDarkTheme } from "./styles/theme";
import { AppLayout } from "./components/layout";
import { ProtectedRoute } from "./components/auth/ProtectedRoute";
import { ErrorBoundary } from "./components/ErrorBoundary";
import { useAuthStore } from "@/stores/authStore";
import { logger } from "@/lib/logger";
import LoginPage from "./pages/auth/LoginPage";
import "./styles/global.css";
import { ThemeProvider } from "./contexts/ThemeContext";
import { useTheme } from "./contexts/ThemeContext.hooks";
import { PageLoading } from "@/components/common/PageLoading";

// Lazy-loaded pages for code splitting
const DashboardPage = React.lazy(
  () => import("./pages/dashboard/DashboardPage"),
);
const OrdersPage = React.lazy(() => import("./pages/orders/OrdersPage"));
const ProductListPage = React.lazy(
  () => import("./pages/products/ProductListPage"),
);
const ProductManagerPage = React.lazy(
  () => import("./pages/product-manager/ProductManagerPage"),
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
const InventoryPage = React.lazy(
  () => import("./pages/inventory/InventoryPage"),
);
const AnalyticsHubPage = React.lazy(
  () => import("./pages/analytics/AnalyticsHubPage"),
);
const ShopeeAnalyticsPage = React.lazy(
  () => import("./pages/analytics/ShopeeAnalyticsPage"),
);
const TiktokAnalyticsPage = React.lazy(
  () => import("./pages/analytics/TiktokAnalyticsPage"),
);
const ShopeeAdsAnalyticsPage = React.lazy(
  () => import("./pages/analytics/ShopeeAdsAnalyticsPage"),
);
const TiktokAdsAnalyticsPage = React.lazy(
  () => import("./pages/analytics/TiktokAdsAnalyticsPage"),
);
const MLDashboardPage = React.lazy(
  () => import("./pages/analytics/MLDashboardPage"),
);
const BudgetSimulatorPage = React.lazy(
  () => import("./pages/analytics/BudgetSimulatorPage"),
);
const ProductClassificationPage = React.lazy(
  () => import("./pages/analytics/ProductClassificationPage"),
);
const AIReportGalleryPage = React.lazy(
  () => import("./pages/analytics/AIReportGalleryPage"),
);
const SettingsPage = React.lazy(() => import("./pages/settings/SettingsPage"));
const ScriptMonitorPage = React.lazy(
  () => import("./pages/script-monitor/ScriptMonitorPage"),
);
const RouteMappingPage = React.lazy(
  () => import("./pages/route-mapping/RouteMappingPage"),
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
                  <Route
                    path="/dashboard"
                    element={<Navigate to="/" replace />}
                  />
                  <Route path="/order-manager" element={<OrdersPage />} />
                  <Route
                    path="/order-manager/:platform"
                    element={<OrdersPage />}
                  />
                  <Route
                    path="/orders"
                    element={<Navigate to="/order-manager" replace />}
                  />
                  <Route
                    path="/master-products"
                    element={<ProductListPage />}
                  />
                  <Route
                    path="/products"
                    element={<Navigate to="/master-products" replace />}
                  />
                  <Route
                    path="/master-products/add"
                    element={<ProductAddPage />}
                  />
                  <Route
                    path="/master-products/import"
                    element={<ProductImportPage />}
                  />
                  <Route
                    path="/master-products/:id"
                    element={<ProductEditPage />}
                  />
                  <Route
                    path="/product-manager"
                    element={<ProductManagerPage />}
                  />
                  <Route
                    path="/product-manager/:platform"
                    element={<ProductManagerPage />}
                  />
                  <Route path="/route-mapping" element={<RouteMappingPage />} />
                  <Route path="/inventory" element={<InventoryPage />} />
                  <Route path="/analytics" element={<AnalyticsHubPage />} />
                  <Route
                    path="/analytics/shopee"
                    element={<ShopeeAnalyticsPage />}
                  />
                  <Route
                    path="/analytics/tiktok"
                    element={<TiktokAnalyticsPage />}
                  />
                  <Route
                    path="/analytics/shopee-ads"
                    element={<ShopeeAdsAnalyticsPage />}
                  />
                  <Route
                    path="/analytics/tiktok-ads"
                    element={<TiktokAdsAnalyticsPage />}
                  />
                  <Route path="/analytics/ml" element={<MLDashboardPage />} />
                  <Route
                    path="/analytics/budget-simulator"
                    element={<BudgetSimulatorPage />}
                  />
                  <Route
                    path="/analytics/product-classification"
                    element={<ProductClassificationPage />}
                  />
                  <Route
                    path="/analytics/ai-reports"
                    element={<AIReportGalleryPage />}
                  />
                  <Route path="/settings" element={<SettingsPage />} />
                  <Route
                    path="/script-monitor"
                    element={<ScriptMonitorPage />}
                  />
                  <Route path="*" element={<Navigate to="/" replace />} />
                </Route>
              </Routes>
            </Suspense>
          </BrowserRouter>
        </ErrorBoundary>
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
