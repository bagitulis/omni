import { useEffect } from "react";
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
import LoginPage from "./pages/auth/LoginPage";
import { DashboardPage } from "./pages/dashboard/DashboardPage";
import OrdersPage from "./pages/orders/OrdersPage";
import { ProductListPage } from "./pages/products/ProductListPage";
import { ProductManagerPage } from "./pages/product-manager/ProductManagerPage";
import ProductAddPage from "./pages/products/ProductAddPage";
import ProductEditPage from "./pages/products/ProductEditPage";
import ProductImportPage from "./pages/products/ProductImportPage";
import "./styles/global.css";
import { ThemeProvider, useTheme } from "./contexts/ThemeContext";

// Placeholder pages for routes not yet implemented
import InventoryPage from "./pages/inventory/InventoryPage";
import AnalyticsHubPage from "./pages/analytics/AnalyticsHubPage";
import ShopeeAnalyticsPage from "./pages/analytics/ShopeeAnalyticsPage";
import TiktokAnalyticsPage from "./pages/analytics/TiktokAnalyticsPage";
import ShopeeAdsAnalyticsPage from "./pages/analytics/ShopeeAdsAnalyticsPage";
import TiktokAdsAnalyticsPage from "./pages/analytics/TiktokAdsAnalyticsPage";
import { MLDashboardPage } from "./pages/analytics/MLDashboardPage";
import BudgetSimulatorPage from "./pages/analytics/BudgetSimulatorPage";
import ProductClassificationPage from "./pages/analytics/ProductClassificationPage";
import AIReportGalleryPage from "./pages/analytics/AIReportGalleryPage";
import SettingsPage from "./pages/settings/SettingsPage";
import ScriptMonitorPage from "./pages/script-monitor/ScriptMonitorPage";
import { RouteMappingPage } from "./pages/route-mapping/RouteMappingPage";

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
      console.debug(
        "[App] CSRF token fetch failed - will retry on next request",
      );
    });
  }, []);

  return (
    <ConfigProvider theme={isDark ? antdDarkTheme : antdTheme}>
      <AntApp>
        <ErrorBoundary>
          <BrowserRouter>
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
                <Route path="/master-products" element={<ProductListPage />} />
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
                <Route path="/script-monitor" element={<ScriptMonitorPage />} />
                <Route path="*" element={<Navigate to="/" replace />} />
              </Route>
            </Routes>
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
