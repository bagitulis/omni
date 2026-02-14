import type { RouteRecordRaw } from "vue-router";
import {
  Dashboard,
  Inventory,
  Settings,
  RouteMapping,
  ScriptMonitor,
  LoginView,
  DevPreview,
  ShopeeAnalytics,
  TiktokAnalytics,
  TiktokAdsAnalytics,
  ShopeeAdsAnalytics,
  AIReportGallery,
  MLDashboard,
  AnalyticsHub,
  BudgetSimulator,
  ProductClassification,
  MasterProductList,
  MasterProductAdd,
  MasterProductEdit,
  MasterProductImport,
} from "./lazyComponents";

// Auth routes
export const authRoutes: RouteRecordRaw[] = [
  {
    path: "/login",
    name: "Login",
    component: LoginView,
    meta: { requiresAuth: false },
  },
];

// Dev Preview routes (development only - no auth required)
export const devRoutes: RouteRecordRaw[] = [
  {
    path: "/dev-preview",
    name: "DevPreview",
    component: DevPreview,
    meta: { requiresAuth: false },
  },
  {
    path: "/dev-preview/:component",
    name: "DevPreviewComponent",
    component: DevPreview,
    meta: { requiresAuth: false },
  },
];

// Product manager routes
export const productManagerRoutes: RouteRecordRaw[] = [
  {
    path: "/product-manager",
    name: "ProductManager",
    component: Dashboard,
    children: [
      {
        path: "shopee",
        name: "ProductManagerShopee",
        component: Dashboard,
        meta: { section: "product-manager", platform: "shopee" },
      },
      {
        path: "lazada",
        name: "ProductManagerLazada",
        component: Dashboard,
        meta: { section: "product-manager", platform: "lazada" },
      },
      {
        path: "tiktok",
        name: "ProductManagerTiktok",
        component: Dashboard,
        meta: { section: "product-manager", platform: "tiktok" },
      },
    ],
  },
];

// Order manager routes
export const orderManagerRoutes: RouteRecordRaw[] = [
  {
    path: "/order-manager",
    name: "OrderManager",
    component: Dashboard,
    meta: { section: "order-manager" },
  },
  {
    path: "/order-manager/shopee",
    redirect: (to) => ({ path: "/order-manager", query: to.query }),
  },
  {
    path: "/order-manager/lazada",
    redirect: (to) => ({ path: "/order-manager", query: to.query }),
  },
  {
    path: "/order-manager/tiktok",
    redirect: (to) => ({ path: "/order-manager", query: to.query }),
  },
];

// Script monitor routes
export const scriptMonitorRoutes: RouteRecordRaw[] = [
  {
    path: "/script-monitor",
    name: "ScriptMonitor",
    component: ScriptMonitor,
    meta: { section: "script-monitor" },
  },
];

// Report routes
export const reportRoutes: RouteRecordRaw[] = [
  {
    path: "/report",
    name: "Report",
    redirect: "/report/shopee",
    meta: { section: "report" },
    children: [
      {
        path: "shopee",
        name: "ShopeeReport",
        component: ShopeeAnalytics,
        meta: { section: "report", subsection: "shopee" },
      },
      {
        path: "tiktok",
        name: "TiktokReport",
        component: TiktokAnalytics,
        meta: { section: "report", subsection: "tiktok" },
      },
    ],
  },
];

// Analytics routes - Ads Performance
export const analyticsRoutes: RouteRecordRaw[] = [
  {
    path: "/analytics",
    name: "Analytics",
    redirect: "/analytics/hub",
    meta: { section: "analytics" },
    children: [
      {
        path: "hub",
        name: "AnalyticsHub",
        component: AnalyticsHub,
        meta: { section: "analytics", subsection: "hub" },
      },
      {
        path: "simulator",
        name: "BudgetSimulator",
        component: BudgetSimulator,
        meta: { section: "analytics", subsection: "simulator" },
      },
      {
        path: "classification",
        name: "ProductClassification",
        component: ProductClassification,
        meta: { section: "analytics", subsection: "classification" },
      },
      {
        path: "ml",
        name: "MLDashboard",
        component: MLDashboard,
        meta: { section: "analytics", subsection: "ml" },
      },
      {
        path: "tiktok-ads",
        name: "TiktokAdsAnalytics",
        component: TiktokAdsAnalytics,
        meta: { section: "analytics", subsection: "tiktok-ads" },
      },
      {
        path: "shopee-ads",
        name: "ShopeeAdsAnalytics",
        component: ShopeeAdsAnalytics,
        meta: { section: "analytics", subsection: "shopee-ads" },
      },
      {
        path: "ai-reports",
        name: "AIReportGallery",
        component: AIReportGallery,
        meta: { section: "analytics", subsection: "ai-reports" },
      },
      // Legacy redirects for backward compatibility
      { path: "shopee", redirect: "/report/shopee" },
      { path: "tiktok", redirect: "/report/tiktok" },
    ],
  },
];

// Settings routes
export const settingsRoutes: RouteRecordRaw[] = [
  {
    path: "/settings",
    name: "Settings",
    component: Settings,
    meta: { section: "settings", subsection: "google-sheets" },
    children: [
      {
        path: "google-sheets",
        name: "SettingsGoogleSheets",
        component: Settings,
        meta: { section: "settings", subsection: "google-sheets" },
      },
      {
        path: "webhook",
        name: "SettingsWebhook",
        component: Settings,
        meta: { section: "settings", subsection: "webhook" },
      },
    ],
  },
];

// Master Product routes
export const masterProductRoutes: RouteRecordRaw[] = [
  {
    path: "/master-products",
    name: "MasterProductList",
    component: MasterProductList,
    meta: { section: "master-product", requiresAuth: true },
  },
  {
    path: "/master-products/add",
    name: "MasterProductAdd",
    component: MasterProductAdd,
    meta: { section: "master-product", subsection: "add", requiresAuth: true },
  },
  {
    path: "/master-products/import",
    name: "MasterProductImport",
    component: MasterProductImport,
    meta: {
      section: "master-product",
      subsection: "import",
      requiresAuth: true,
    },
  },
  {
    path: "/master-products/:id",
    name: "MasterProductEdit",
    component: MasterProductEdit,
    meta: { section: "master-product", subsection: "edit", requiresAuth: true },
  },
];

// Core routes
export const coreRoutes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "Dashboard",
    component: Dashboard,
    meta: { requiresAuth: true },
  },
  {
    path: "/inventory",
    name: "Inventory",
    component: Inventory,
    meta: { section: "inventory" },
  },
  { path: "/route-mapping", name: "RouteMapping", component: RouteMapping },
];

// Combine all routes
export const routes: RouteRecordRaw[] = [
  ...authRoutes,
  ...devRoutes,
  ...coreRoutes,
  ...masterProductRoutes,
  ...productManagerRoutes,
  ...orderManagerRoutes,
  ...scriptMonitorRoutes,
  ...reportRoutes,
  ...analyticsRoutes,
  ...settingsRoutes,
  { path: "/:pathMatch(.*)*", redirect: "/" },
];
