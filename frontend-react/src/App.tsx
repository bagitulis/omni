import { useState } from "react";
import { ConfigProvider } from "antd";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
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
const Analytics = () => <h1>Analytics</h1>;
const Settings = () => <h1>Settings</h1>;

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
              <Route path="/dashboard" element={<DashboardPage />} />
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
              <Route path="/orders" element={<OrdersPage />} />
              <Route path="/products" element={<ProductListPage />} />
              <Route path="/products/add" element={<ProductAddPage />} />
              <Route path="/products/edit/:id" element={<ProductEditPage />} />
              <Route path="/products/import" element={<ProductImportPage />} />
              <Route path="/inventory" element={<InventoryPage />} />
              <Route path="/analytics" element={<Analytics />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="*" element={<Navigate to="/dashboard" replace />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </ConfigProvider>
    </QueryClientProvider>
  );
}
