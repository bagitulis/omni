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
  AdminLayout,
  AdminDashboardPage,
  AdminUsersPage,
  AdminRolesPage,
  AdminAuditPage,
  AdminSettingsPage,
  AdminShopSetupPage,
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

// Operation routes
export const operationRoutes: RouteRecordRaw[] = [
  {
    path: "/operation",
    name: "Operation",
    component: Dashboard,
    children: [
      {
        path: "shopee",
        name: "OperationShopee",
        component: Dashboard,
        meta: { section: "operation", platform: "shopee" },
      },
      {
        path: "lazada",
        name: "OperationLazada",
        component: Dashboard,
        meta: { section: "operation", platform: "lazada" },
      },
      {
        path: "tiktok",
        name: "OperationTiktok",
        component: Dashboard,
        meta: { section: "operation", platform: "tiktok" },
      },
    ],
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
    meta: { section: "script-monitor", subsection: "current" },
    redirect: "/script-monitor/current",
    children: [
      {
        path: "current",
        name: "ScriptMonitorCurrent",
        component: ScriptMonitor,
        meta: {
          section: "script-monitor",
          subsection: "current",
          tab: "current",
        },
      },
      {
        path: "queue",
        name: "ScriptMonitorQueue",
        component: ScriptMonitor,
        meta: { section: "script-monitor", subsection: "queue", tab: "queue" },
      },
      {
        path: "history",
        name: "ScriptMonitorHistory",
        component: ScriptMonitor,
        meta: {
          section: "script-monitor",
          subsection: "history",
          tab: "history",
        },
      },
      {
        path: "auto-functions",
        name: "ScriptMonitorAutoFunctions",
        component: ScriptMonitor,
        meta: {
          section: "script-monitor",
          subsection: "auto-functions",
          tab: "config",
        },
      },
    ],
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

// Analytics routes
export const analyticsRoutes: RouteRecordRaw[] = [
  {
    path: "/analytics",
    name: "Analytics",
    redirect: "/analytics/tiktok-ads",
    meta: { section: "analytics" },
    children: [
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
        path: "logs",
        name: "SettingsLogs",
        component: Settings,
        meta: { section: "settings", subsection: "logs" },
      },
      {
        path: "resources",
        name: "SettingsResources",
        component: Settings,
        meta: { section: "settings", subsection: "resources" },
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

// Admin routes
const adminMeta = { requiresAuth: true, requiresAdminRole: true };
export const adminRoutes: RouteRecordRaw[] = [
  {
    path: "/admin",
    name: "AdminLayout",
    component: AdminLayout,
    meta: adminMeta,
    redirect: "/admin/dashboard",
    children: [
      {
        path: "dashboard",
        name: "AdminDashboard",
        component: AdminDashboardPage,
        meta: adminMeta,
      },
      {
        path: "users",
        name: "AdminUsers",
        component: AdminUsersPage,
        meta: adminMeta,
      },
      {
        path: "roles",
        name: "AdminRoles",
        component: AdminRolesPage,
        meta: adminMeta,
      },
      {
        path: "audit",
        name: "AdminAudit",
        component: AdminAuditPage,
        meta: adminMeta,
      },
      {
        path: "settings",
        name: "AdminSettings",
        component: AdminSettingsPage,
        meta: adminMeta,
      },
      {
        path: "shop-setup",
        name: "AdminShopSetup",
        component: AdminShopSetupPage,
        meta: adminMeta,
      },
    ],
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
  ...operationRoutes,
  ...productManagerRoutes,
  ...orderManagerRoutes,
  ...scriptMonitorRoutes,
  ...reportRoutes,
  ...analyticsRoutes,
  ...settingsRoutes,
  ...adminRoutes,
  { path: "/:pathMatch(.*)*", redirect: "/" },
];
