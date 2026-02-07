import { useState } from "react";
import { ConfigProvider } from "antd";
import { BrowserRouter, Routes, Route, Navigate, useParams } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { antdTheme, antdDarkTheme } from "./styles/theme";
import { AppLayout } from "./components/layout";
import LoginPage from "./pages/auth/LoginPage";
import { DashboardPage } from "./pages/dashboard/DashboardPage";
import OrdersPage from "./pages/orders/OrdersPage";
import { ProductListPage } from "./pages/products/ProductListPage";
import ProductAddPage from "./pages/products/ProductAddPage";
import ProductEditPage from "./pages/products/ProductEditPage";
import ProductImportPage from "./pages/products/ProductImportPage";
import "./styles/global.css";

// Create a QueryClient instance for TanStack Query
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      retry: 1,
    },
  },
});

// Placeholder pages for routes not yet implemented
import InventoryPage from "./pages/inventory/InventoryPage";
import AnalyticsHubPage from "./pages/analytics/AnalyticsHubPage";
import ShopeeAnalyticsPage from "./pages/analytics/ShopeeAnalyticsPage";
import TiktokAnalyticsPage from "./pages/analytics/TiktokAnalyticsPage";
import ShopeeAdsAnalyticsPage from "./pages/analytics/ShopeeAdsAnalyticsPage";
import TiktokAdsAnalyticsPage from "./pages/analytics/TiktokAdsAnalyticsPage";
import { MLDashboardPage } from "./pages/analytics/MLDashboardPage";
import SettingsPage from "./pages/settings/SettingsPage";
import ScriptMonitorPage from "./pages/script-monitor/ScriptMonitorPage";
import { RouteMappingPage } from "./pages/route-mapping/RouteMappingPage";

function ProductManagerRedirect() {
  const { platform } = useParams<{ platform?: string }>();

  if (platform) {
    const next = `/master-products?platform=${encodeURIComponent(platform)}`;
    return <Navigate to={next} replace />;
  }

  return <Navigate to="/master-products" replace />;
}

export default function App() {
  // Theme state kept for future implementation
  const [isDark] = useState(false);

  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider theme={isDark ? antdDarkTheme : antdTheme}>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route element={<AppLayout />}>
              <Route path="/" element={<DashboardPage />} />
              <Route path="/dashboard" element={<Navigate to="/" replace />} />
              <Route path="/order-manager" element={<OrdersPage />} />
              <Route
                path="/orders"
                element={<Navigate to="/order-manager" replace />}
              />
              <Route path="/master-products" element={<ProductListPage />} />
              <Route
                path="/products"
                element={<Navigate to="/master-products" replace />}
              />
              <Route path="/master-products/add" element={<ProductAddPage />} />
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
                element={<ProductManagerRedirect />}
              />
              <Route
                path="/product-manager/:platform"
                element={<ProductManagerRedirect />}
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
              <Route path="/settings" element={<SettingsPage />} />
              <Route path="/script-monitor" element={<ScriptMonitorPage />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </ConfigProvider>
    </QueryClientProvider>
  );
}
