/**
 * Lazy load component definitions for Vue Router
 * Extracted for maintainability (agents.md: max 300 lines per file)
 */

// Core views
export const Dashboard = () =>
  import(/* webpackChunkName: "dashboard" */ "../views/Dashboard.vue");
export const Inventory = () =>
  import(/* webpackChunkName: "inventory" */ "../views/Inventory.vue");
export const Settings = () =>
  import(/* webpackChunkName: "settings" */ "../views/Settings.vue");
export const RouteMapping = () =>
  import(/* webpackChunkName: "route-mapping" */ "../views/RouteMapping.vue");
export const ScriptMonitor = () =>
  import(/* webpackChunkName: "script-monitor" */ "../views/ScriptMonitor.vue");
export const LoginView = () =>
  import(/* webpackChunkName: "login" */ "../views/LoginView.vue");
export const DevPreview = () =>
  import(/* webpackChunkName: "dev-preview" */ "../views/DevPreview.vue");

// Analytics views
export const ShopeeAnalytics = () =>
  import(
    /* webpackChunkName: "analytics" */ "../views/analytics/ShopeeAnalytics.vue"
  );
export const TiktokAnalytics = () =>
  import(
    /* webpackChunkName: "analytics-tiktok" */ "../views/analytics/TiktokAnalytics.vue"
  );
export const TiktokAdsAnalytics = () =>
  import(
    /* webpackChunkName: "analytics-tiktok-ads" */ "../views/analytics/TiktokAdsAnalytics.vue"
  );
export const ShopeeAdsAnalytics = () =>
  import(
    /* webpackChunkName: "analytics-shopee-ads" */ "../views/analytics/ShopeeAdsAnalytics.vue"
  );

// Admin panel views
export const AdminLayout = () =>
  import(/* webpackChunkName: "admin" */ "../views/AdminLayout.vue");
export const AdminDashboardPage = () =>
  import(
    /* webpackChunkName: "admin-dashboard" */ "../views/AdminDashboardPage.vue"
  );
export const AdminUsersPage = () =>
  import(/* webpackChunkName: "admin-users" */ "../views/AdminUsersPage.vue");
export const AdminRolesPage = () =>
  import(/* webpackChunkName: "admin-roles" */ "../views/AdminRolesPage.vue");
export const AdminAuditPage = () =>
  import(/* webpackChunkName: "admin-audit" */ "../views/AdminAuditPage.vue");
export const AdminSettingsPage = () =>
  import(
    /* webpackChunkName: "admin-settings" */ "../views/AdminSettingsPage.vue"
  );
export const AdminShopSetupPage = () =>
  import(
    /* webpackChunkName: "admin-shop-setup" */ "../views/AdminShopSetupPage.vue"
  );
