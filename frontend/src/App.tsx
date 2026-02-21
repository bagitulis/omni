import React, { Suspense, useEffect } from "react";
import { ConfigProvider, App as AntApp } from "antd";
import {
  BrowserRouter,
  Routes,
  Route,
  Navigate,
  useParams,
} from "react-router-dom";
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
const AnalyticsHubPage = React.lazy(
  () => import("./pages/analytics/AnalyticsHubPage"),
);
const ShopeeReportPage = React.lazy(
  () => import("./pages/report/ShopeeReportPage"),
);
const TiktokReportPage = React.lazy(
  () => import("./pages/report/TiktokReportPage"),
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

// Redirect component for /master-products/:id -> /products/:id/edit
function MasterProductRedirect() {
  const { id } = useParams();
  return <Navigate to={`/products/${id}/edit`} replace />;
}

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
                  <Route
                    path="/master-products"
                    element={<Navigate to="/products" replace />}
                  />
                  <Route
                    path="/master-products/add"
                    element={<Navigate to="/products/add" replace />}
                  />
                  <Route
                    path="/master-products/import"
                    element={<Navigate to="/products/import" replace />}
                  />
                  <Route
                    path="/master-products/:id"
                    element={<MasterProductRedirect />}
                  />
                  <Route
                    path="/product-manager"
                    element={<Navigate to="/products" replace />}
                  />
                  <Route
                    path="/product-manager/:platform"
                    element={<Navigate to="/products" replace />}
                  />
                  <Route path="/route-mapping" element={<RouteMappingPage />} />
                  <Route
                    path="/inventory"
                    element={<SimplifiedInventoryPage />}
                  />
                  <Route path="/analytics" element={<AnalyticsHubPage />} />
                  <Route
                    path="/report/shopee"
                    element={<ShopeeReportPage />}
                  />
                  <Route
                    path="/report/tiktok"
                    element={<TiktokReportPage />}
                  />
                  {/* Redirects from old analytics URLs */}
                  <Route
                    path="/analytics/shopee"
                    element={<Navigate to="/report/shopee" replace />}
                  />
                  <Route
                    path="/analytics/tiktok"
                    element={<Navigate to="/report/tiktok" replace />}
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
